// Package pg provides a database/sql driver for PostgreSQL databases within Spin components.
package pg

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"time"

	pg "github.com/spinframework/spin-go-sdk/v3/imports/spin_postgres_4_2_0_postgres"
	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

// Open returns a new connection to the database.
func Open(name string) *sql.DB {
	return sql.OpenDB(&connector{name: name})
}

// connector implements driver.Connector.
type connector struct {
	conn *conn
	name string
}

// Connect returns a connection to the database.
func (d *connector) Connect(_ context.Context) (driver.Conn, error) {
	if d.conn != nil {
		return d.conn, nil
	}
	return d.Open(d.name)
}

// Driver returns the underlying Driver of the Connector.
func (d *connector) Driver() driver.Driver {
	return d
}

// Open returns a new connection to the database.
func (d *connector) Open(name string) (driver.Conn, error) {
	results := pg.ConnectionOpenAsync(name)
	if results.IsErr() {
		return nil, toError(results.Err())
	}
	d.conn = &conn{spinConn: *results.Ok()}
	return d.conn, nil
}

// conn implements driver.Conn
type conn struct {
	spinConn pg.Connection
}

var _ driver.Conn = (*conn)(nil)

// Prepare returns a prepared statement, bound to this connection.
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{conn: c, query: query}, nil
}

func (c *conn) Close() error {
	c.spinConn.Drop()
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported by this driver")
}

type stmt struct {
	conn  *conn
	query string
}

var _ driver.Stmt = (*stmt)(nil)
var _ driver.ColumnConverter = (*stmt)(nil)

// Close closes the statement.
func (s *stmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters.
func (s *stmt) NumInput() int {
	// Golang sql won't sanity check argument counts before Query.
	return -1
}

// Query executes a query that may return rows, such as a SELECT.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	rdbmsParams := make([]pg.ParameterValue, len(args))
	for i, v := range args {
		rdbmsParams[i] = toRdbmsParameterValue(v)
	}

	results := s.conn.spinConn.QueryAsync(s.query, rdbmsParams)
	if results.IsErr() {
		return nil, toError(results.Err())
	}

	tuple := results.Ok()
	cols := tuple.F0
	colNames := make([]string, len(cols))
	colTypes := make([]uint8, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name
		colTypes[i] = uint8(c.DataType.Tag())
	}

	rows := &rows{
		columns:    colNames,
		columnType: colTypes,
		stream:     tuple.F1,
		future:     tuple.F2,
	}

	rows.next = rows.pull()
	return rows, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	rdbmsParams := make([]pg.ParameterValue, len(args))
	for i, v := range args {
		rdbmsParams[i] = toRdbmsParameterValue(v)
	}

	queryResult := s.conn.spinConn.ExecuteAsync(s.query, rdbmsParams)
	if queryResult.IsErr() {
		return &result{}, toError(queryResult.Err())
	}

	return &result{}, nil
}

// ColumnConverter returns GlobalParameterConverter to prevent using driver.DefaultParameterConverter.
func (s *stmt) ColumnConverter(_ int) driver.ValueConverter {
	return spindb.GlobalParameterConverter
}

type result struct {
	rowsAffected int64
}

func (r result) LastInsertId() (int64, error) {
	return -1, errors.New("LastInsertId is unsupported by this driver")
}

func (r result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type rows struct {
	columns    []string
	columnType []uint8
	next       []any
	stream     *wittypes.StreamReader[[]pg.DbValue]
	future     *wittypes.FutureReader[wittypes.Result[wittypes.Unit, pg.Error]]
	result     error
}

var _ driver.Rows = (*rows)(nil)
var _ driver.RowsColumnTypeScanType = (*rows)(nil)
var _ driver.RowsNextResultSet = (*rows)(nil)

// Columns returns the column names.
func (r *rows) Columns() []string {
	return r.columns
}

// Close closes the rows iterator.
func (r *rows) Close() error {
	r.stream.Drop()
	r.future.Drop()
	r.stream = nil
	r.future = nil
	r.next = nil
	r.result = io.EOF
	return nil
}

func (r *rows) pull() []any {
	buffer := [][]pg.DbValue{nil}
	if r.stream.Read(buffer) == 1 {
		return toRow(buffer[0])
	}
	result := r.future.Read()
	if result.IsOk() {
		r.result = io.EOF
	} else {
		r.result = toError(result.Err())
	}
	return nil
}

// Next moves the cursor to the next row.
func (r *rows) Next(dest []driver.Value) error {
	if !r.HasNextResultSet() {
		return r.result
	}
	next := r.next
	r.next = r.pull()
	for i := 0; i != len(r.columns); i++ {
		dest[i] = driver.Value(next[i])
	}
	return nil
}

// HasNextResultSet is called at the end of the current result set and
// reports whether there is another result set after the current one.
func (r *rows) HasNextResultSet() bool {
	return r.next != nil
}

// NextResultSet advances the driver to the next result set even
// if there are remaining rows in the current result set.
//
// NextResultSet should return io.EOF when there are no more result sets.
func (r *rows) NextResultSet() error {
	if r.HasNextResultSet() {
		r.next = r.pull()
		return nil
	}
	return r.result
}

// ColumnTypeScanType returns the value type that can be used to scan types into.
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	return colTypeToReflectType(r.columnType[index])
}

func toRdbmsParameterValue(x any) pg.ParameterValue {
	switch v := x.(type) {
	case bool:
		return pg.MakeParameterValueBoolean(v)
	case int8:
		return pg.MakeParameterValueInt8(v)
	case int16:
		return pg.MakeParameterValueInt16(v)
	case int32:
		return pg.MakeParameterValueInt32(v)
	case int64:
		return pg.MakeParameterValueInt64(v)
	case int:
		return pg.MakeParameterValueInt64(int64(v))
	case float32:
		return pg.MakeParameterValueFloating32(v)
	case float64:
		return pg.MakeParameterValueFloating64(v)
	case string:
		return pg.MakeParameterValueStr(v)
	case []byte:
		return pg.MakeParameterValueBinary(v)
	case []string:
		return pg.MakeParameterValueArrayStr(toOptionSlice(v))
	case Int32Range:
		witVal, _ := v.Value()
		return pg.MakeParameterValueRangeInt32(witVal.(wittypes.Tuple2[wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]], wittypes.Option[wittypes.Tuple2[int32, pg.RangeBoundKind]]]))
	case Int64Range:
		witVal, _ := v.Value()
		return pg.MakeParameterValueRangeInt64(witVal.(wittypes.Tuple2[wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]], wittypes.Option[wittypes.Tuple2[int64, pg.RangeBoundKind]]]))
	case []int32:
		return pg.MakeParameterValueArrayInt32(toOptionSlice(v))
	case []int64:
		return pg.MakeParameterValueArrayInt64(toOptionSlice(v))
	case time.Time:
		v = v.UTC()
		return pg.MakeParameterValueDatetime(wittypes.Tuple7[int32, uint8, uint8, uint8, uint8, uint8, uint32]{
			F0: int32(v.Year()),
			F1: uint8(v.Month()),
			F2: uint8(v.Day()),
			F3: uint8(v.Hour()),
			F4: uint8(v.Minute()),
			F5: uint8(v.Second()),
			F6: uint32(v.Nanosecond()),
		})
	case time.Duration:
		// TODO: does this need to be parsed into Micros/Days/Months?
		return pg.MakeParameterValueInterval(pg.Interval{
			Micros: v.Microseconds(),
		})
	case nil:
		return pg.MakeParameterValueDbNull()
	default:
		panic("unknown value type")
	}
}

func toError(err pg.Error) error {
	switch err.Tag() {
	case pg.ErrorBadParameter:
		return errors.New(err.BadParameter())
	case pg.ErrorConnectionFailed:
		return errors.New(err.ConnectionFailed())
	case pg.ErrorQueryFailed:
		return errors.New(err.QueryFailed().Text())
	case pg.ErrorValueConversionFailed:
		return errors.New(err.ValueConversionFailed())
	case pg.ErrorOther:
		return errors.New(err.Other())
	default:
		panic("unknown error from runtime")
	}
}

func toRow(row []pg.DbValue) []any {
	result := make([]any, len(row))
	for i, v := range row {
		switch v.Tag() {
		case pg.DbValueBoolean:
			result[i] = v.Boolean()
		case pg.DbValueInt8:
			result[i] = v.Int8()
		case pg.DbValueInt16:
			result[i] = v.Int16()
		case pg.DbValueInt32:
			result[i] = v.Int32()
		case pg.DbValueInt64:
			result[i] = v.Int64()
		case pg.DbValueFloating32:
			result[i] = v.Floating32()
		case pg.DbValueFloating64:
			result[i] = v.Floating64()
		case pg.DbValueStr:
			result[i] = v.Str()
		case pg.DbValueBinary:
			result[i] = v.Binary()
		case pg.DbValueDate:
			d := v.Date()
			result[i] = time.Date(int(d.F0), time.Month(d.F1), int(d.F2), 0, 0, 0, 0, time.UTC)
		case pg.DbValueTime:
			t := v.Time()
			result[i] = time.Date(0, 1, 1, int(t.F0), int(t.F1), int(t.F2), int(t.F3), time.UTC)
		case pg.DbValueDatetime:
			dt := v.Datetime()
			result[i] = time.Date(int(dt.F0), time.Month(dt.F1), int(dt.F2), int(dt.F3), int(dt.F4), int(dt.F5), int(dt.F6), time.UTC)
		case pg.DbValueTimestamp:
			result[i] = time.Unix(v.Timestamp(), 0).UTC()
		case pg.DbValueUuid:
			result[i] = v.Uuid()

		// TODO make these all go types
		// case pg.DbValueJsonb:
		// 	result[i] = v.Jsonb()
		// case pg.DbValueDecimal:
		// 	result[i] = v.Decimal()
		case pg.DbValueRangeInt32:
			result[i] = v.RangeInt32()
		// case pg.DbValueRangeInt64:
		// 	result[i] = v.RangeInt64()
		// case pg.DbValueRangeDecimal:
		// 	result[i] = v.RangeDecimal()
		case pg.DbValueArrayInt32:
			result[i] = fromOptionSlice(v.ArrayInt32())
		case pg.DbValueArrayInt64:
			result[i] = fromOptionSlice(v.ArrayInt64())
		// case pg.DbValueArrayDecimal:
		// 	result[i] = v.ArrayDecimal()
		case pg.DbValueArrayStr:
			result[i] = fromOptionSlice(v.ArrayStr())

		// TODO: time.duration
		// case pg.DbValueInterval:
		// 	result[i] = v.Interval()
		case pg.DbValueUnsupported:
			result[i] = v.Unsupported()
		case pg.DbValueDbNull:
			result[i] = nil
		default:
			panic("unknown value type")
		}
	}

	return result
}

func colTypeToReflectType(typ uint8) reflect.Type {
	switch typ {
	case uint8(pg.DbDataTypeBoolean):
		return reflect.TypeFor[bool]()
	case uint8(pg.DbDataTypeInt8):
		return reflect.TypeFor[int8]()
	case uint8(pg.DbDataTypeInt16):
		return reflect.TypeFor[int16]()
	case uint8(pg.DbDataTypeInt32):
		return reflect.TypeFor[int32]()
	case uint8(pg.DbDataTypeInt64):
		return reflect.TypeFor[int64]()
	case uint8(pg.DbDataTypeFloating32):
		return reflect.TypeFor[float32]()
	case uint8(pg.DbDataTypeFloating64):
		return reflect.TypeFor[float64]()
	case uint8(pg.DbDataTypeStr):
		return reflect.TypeFor[string]()
	case uint8(pg.DbDataTypeUuid):
		return reflect.TypeFor[string]()
	case uint8(pg.DbDataTypeDecimal):
		// TODO:
		// return reflect.TypeFor[string]()
	case uint8(pg.DbDataTypeBinary):
		return reflect.TypeFor[[]byte]()
	case uint8(pg.DbDataTypeJsonb):
		return reflect.TypeFor[[]byte]()
	case uint8(pg.DbDataTypeDate),
		uint8(pg.DbDataTypeTime),
		uint8(pg.DbDataTypeDatetime),
		uint8(pg.DbDataTypeTimestamp):
		return reflect.TypeFor[time.Time]()
	case uint8(pg.DbDataTypeInterval):
		// TODO
	case uint8(pg.DbDataTypeRangeInt32):
		return reflect.TypeFor[Int32Range]()
	case uint8(pg.DbDataTypeRangeInt64):
		return reflect.TypeFor[Int64Range]()
	case uint8(pg.DbDataTypeRangeDecimal):
		// TODO
	case uint8(pg.DbDataTypeArrayInt32):
		return reflect.TypeFor[[]int32]()
	case uint8(pg.DbDataTypeArrayInt64):
		return reflect.TypeFor[[]int64]()
	case uint8(pg.DbDataTypeArrayDecimal):
		// TODO
	case uint8(pg.DbDataTypeArrayStr):
		return reflect.TypeFor[[]string]()
	case uint8(pg.DbDataTypeOther):
		return reflect.TypeFor[any]().Elem()
	}
	panic("invalid db column type of " + string(typ))
}

func toOptionSlice[T any](v []T) []wittypes.Option[T] {
	values := make([]wittypes.Option[T], len(v))
	for i, x := range v {
		values[i] = wittypes.Some(x)
	}
	return values
}

func fromOptionSlice[T any](v []wittypes.Option[T]) []T {
	values := make([]T, len(v))
	for i, x := range v {
		if x.IsSome() {
			values[i] = x.Some()
		}
	}
	return values
}

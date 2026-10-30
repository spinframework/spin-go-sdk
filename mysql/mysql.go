// Package mysql provides a database/sql driver for MySQL databases within Spin components.
package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"

	mysql "github.com/spinframework/spin-go-sdk/v3/imports/spin_mysql_3_0_0_mysql"
	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	wittypes "go.bytecodealliance.org/pkg/wit/types"
)

// Open returns a new connection to the database.
func Open(name string) *sql.DB {
	return sql.OpenDB(&connector{name: name})
}

type conn struct {
	spinConn mysql.Connection
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{conn: c, query: query}, nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported by this driver")
}

type connector struct {
	conn *conn
	name string
}

var _ driver.Conn = (*conn)(nil)

func (d *connector) Connect(_ context.Context) (driver.Conn, error) {
	if d.conn != nil {
		return d.conn, nil
	}
	return d.Open(d.name)
}

func (d *connector) Driver() driver.Driver {
	return d
}

func (d *connector) Open(name string) (driver.Conn, error) {
	results := mysql.ConnectionOpen(name)
	if results.IsErr() {
		return nil, toError(results.Err())
	}
	d.conn = &conn{spinConn: *results.Ok()}
	return d.conn, nil
}

func (d *connector) Close() error {
	if d.conn != nil {
		d.conn.Close()
	}
	return nil
}

type rows struct {
	columns    []string
	columnType []uint8
	next       []any
	stream     *wittypes.StreamReader[[]mysql.DbValue]
	future     *wittypes.FutureReader[wittypes.Result[wittypes.Unit, mysql.Error]]
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
	buffer := [][]mysql.DbValue{nil}
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

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	wasiParams := make([]mysql.ParameterValue, len(args))
	for i, v := range args {
		wasiParams[i] = toWasiParameterValue(v)
	}

	queryResult := s.conn.spinConn.Execute(s.query, wasiParams)
	if queryResult.IsErr() {
		return &result{}, toError(queryResult.Err())
	}

	return &result{}, nil
}

// Query executes a query that may return rows, such as a SELECT.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	wasiParams := make([]mysql.ParameterValue, len(args))
	for i, v := range args {
		wasiParams[i] = toWasiParameterValue(v)
	}

	results := s.conn.spinConn.Query(s.query, wasiParams)
	if results.IsErr() {
		return nil, toError(results.Err())
	}

	tuple := results.Ok()
	cols := tuple.F0
	colNames := make([]string, len(cols))
	colTypes := make([]uint8, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name
		colTypes[i] = uint8(c.DataType)
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

func toWasiParameterValue(x any) mysql.ParameterValue {
	switch v := x.(type) {
	case bool:
		return mysql.MakeParameterValueBoolean(v)
	case int8:
		return mysql.MakeParameterValueInt8(v)
	case int16:
		return mysql.MakeParameterValueInt16(v)
	case int32:
		return mysql.MakeParameterValueInt32(v)
	case int64:
		return mysql.MakeParameterValueInt64(v)
	case int:
		return mysql.MakeParameterValueInt64(int64(v))
	case uint8:
		return mysql.MakeParameterValueUint8(v)
	case uint16:
		return mysql.MakeParameterValueUint16(v)
	case uint32:
		return mysql.MakeParameterValueUint32(v)
	case uint64:
		return mysql.MakeParameterValueUint64(v)
	case float32:
		return mysql.MakeParameterValueFloating32(v)
	case float64:
		return mysql.MakeParameterValueFloating64(v)
	case string:
		return mysql.MakeParameterValueStr(v)
	case []byte:
		return mysql.MakeParameterValueBinary(v)
	case nil:
		return mysql.MakeParameterValueDbNull()
	default:
		panic("unknown value type")
	}
}

func toError(err mysql.Error) error {
	switch err.Tag() {
	case mysql.Error_BadParameter:
		return errors.New(err.BadParameter())
	case mysql.Error_ConnectionFailed:
		return errors.New(err.ConnectionFailed())
	case mysql.Error_QueryFailed:
		return errors.New(err.QueryFailed())
	case mysql.Error_ValueConversionFailed:
		return errors.New(err.ValueConversionFailed())
	default:
		// TODO: not sure if using "Other" as the default is appropriate
		return errors.New(err.Other())
	}
}

func toRow(row []mysql.DbValue) []any {
	result := make([]any, len(row))
	for i, v := range row {
		switch v.Tag() {
		case mysql.DbValue_Boolean:
			result[i] = v.Boolean()
		case mysql.DbValue_Int8:
			result[i] = v.Int8()
		case mysql.DbValue_Int16:
			result[i] = v.Int16()
		case mysql.DbValue_Int32:
			result[i] = v.Int32()
		case mysql.DbValue_Int64:
			result[i] = v.Int64()
		case mysql.DbValue_Uint8:
			result[i] = v.Uint8()
		case mysql.DbValue_Uint16:
			result[i] = v.Uint16()
		case mysql.DbValue_Uint32:
			result[i] = v.Uint32()
		case mysql.DbValue_Uint64:
			result[i] = v.Uint64()
		case mysql.DbValue_Floating32:
			result[i] = v.Floating32()
		case mysql.DbValue_Floating64:
			result[i] = v.Floating64()
		case mysql.DbValue_Str:
			result[i] = v.Str()
		case mysql.DbValue_Binary:
			result[i] = v.Binary()
		case mysql.DbValue_DbNull:
			result[i] = nil
		default:
			panic("unknown value type")
		}
	}

	return result
}

// ColumnConverter returns GlobalParameterConverter to prevent using driver.DefaultParameterConverter.
func (s *stmt) ColumnConverter(_ int) driver.ValueConverter {
	return spindb.GlobalParameterConverter
}

type result struct{}

func (r result) LastInsertId() (int64, error) {
	return -1, errors.New("LastInsertId is unsupported by this driver")
}

func (r result) RowsAffected() (int64, error) {
	return -1, errors.New("RowsAffected is unsupported by this driver")
}

func colTypeToReflectType(typ uint8) reflect.Type {
	switch typ {
	case uint8(mysql.DbDataType_Boolean):
		return reflect.TypeOf(false)
	case uint8(mysql.DbDataType_Int8):
		return reflect.TypeOf(int8(0))
	case uint8(mysql.DbDataType_Int16):
		return reflect.TypeOf(int16(0))
	case uint8(mysql.DbDataType_Int32):
		return reflect.TypeOf(int32(0))
	case uint8(mysql.DbDataType_Int64):
		return reflect.TypeOf(int64(0))
	case uint8(mysql.DbDataType_Uint8):
		return reflect.TypeOf(uint8(0))
	case uint8(mysql.DbDataType_Uint16):
		return reflect.TypeOf(uint16(0))
	case uint8(mysql.DbDataType_Uint32):
		return reflect.TypeOf(uint32(0))
	case uint8(mysql.DbDataType_Uint64):
		return reflect.TypeOf(uint64(0))
	case uint8(mysql.DbDataType_Str):
		return reflect.TypeOf("")
	case uint8(mysql.DbDataType_Binary):
		return reflect.TypeOf(new([]byte))
	case uint8(mysql.DbDataType_Other):
		return reflect.TypeOf(new(any)).Elem()
	}
	panic("invalid db column type of " + string(typ))
}

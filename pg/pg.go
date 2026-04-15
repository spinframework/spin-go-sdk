// Package pg provides a database/sql driver for PostgreSQL databases within Spin components.
package pg

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"

	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	pg "github.com/spinframework/spin-go-sdk/v3/imports/fermyon_spin_2_0_0_postgres"
	rdbmstypes "github.com/spinframework/spin-go-sdk/v3/imports/fermyon_spin_2_0_0_rdbms_types"
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
	results := pg.ConnectionOpen(name)
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

	results := s.conn.spinConn.Query(s.query, rdbmsParams)
	if results.IsErr() {
		return nil, toError(results.Err())
	}

	rowLen := len(results.Ok().Rows)
	allRows := make([][]any, rowLen)
	for rowNum, row := range results.Ok().Rows {
		allRows[rowNum] = toRow(row)
	}

	cols := results.Ok().Columns
	colNames := make([]string, len(cols))
	colTypes := make([]uint8, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name
		colTypes[i] = uint8(c.DataType)
	}

	rows := &rows{
		columns:    colNames,
		columnType: colTypes,
		rows:       allRows,
		len:        int(rowLen),
	}
	return rows, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	rdbmsParams := make([]pg.ParameterValue, len(args))
	for i, v := range args {
		rdbmsParams[i] = toRdbmsParameterValue(v)
	}

	queryResult := s.conn.spinConn.Execute(s.query, rdbmsParams)
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
	pos        int
	len        int
	rows       [][]any
	closed     bool
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
	r.rows = nil
	r.pos = 0
	r.len = 0
	r.closed = true
	return nil
}

// Next moves the cursor to the next row.
func (r *rows) Next(dest []driver.Value) error {
	if !r.HasNextResultSet() {
		return io.EOF
	}
	for i := 0; i != len(r.columns); i++ {
		dest[i] = driver.Value(r.rows[r.pos][i])
	}
	r.pos++
	return nil
}

// HasNextResultSet is called at the end of the current result set and
// reports whether there is another result set after the current one.
func (r *rows) HasNextResultSet() bool {
	return r.pos < r.len
}

// NextResultSet advances the driver to the next result set even
// if there are remaining rows in the current result set.
//
// NextResultSet should return io.EOF when there are no more result sets.
func (r *rows) NextResultSet() error {
	if r.HasNextResultSet() {
		r.pos++
		return nil
	}
	return io.EOF // Per interface spec.
}

// ColumnTypeScanType returns the value type that can be used to scan types into.
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	return colTypeToReflectType(r.columnType[index])
}

func toRdbmsParameterValue(x any) pg.ParameterValue {
	switch v := x.(type) {
	case bool:
		return rdbmstypes.MakeParameterValueBoolean(v)
	case int8:
		return rdbmstypes.MakeParameterValueInt8(v)
	case int16:
		return rdbmstypes.MakeParameterValueInt16(v)
	case int32:
		return rdbmstypes.MakeParameterValueInt32(v)
	case int64:
		return rdbmstypes.MakeParameterValueInt64(v)
	case int:
		return rdbmstypes.MakeParameterValueInt64(int64(v))
	case uint8:
		return rdbmstypes.MakeParameterValueUint8(v)
	case uint16:
		return rdbmstypes.MakeParameterValueUint16(v)
	case uint32:
		return rdbmstypes.MakeParameterValueUint32(v)
	case uint64:
		return rdbmstypes.MakeParameterValueUint64(v)
	case float32:
		return rdbmstypes.MakeParameterValueFloating32(v)
	case float64:
		return rdbmstypes.MakeParameterValueFloating64(v)
	case string:
		return rdbmstypes.MakeParameterValueStr(v)
	case []byte:
		return rdbmstypes.MakeParameterValueBinary(v)
	case nil:
		return rdbmstypes.MakeParameterValueDbNull()
	default:
		panic("unknown value type")
	}
}

func toError(err pg.Error) error {
	switch err.Tag() {
	case rdbmstypes.ErrorBadParameter:
		return errors.New(err.BadParameter())
	case rdbmstypes.ErrorConnectionFailed:
		return errors.New(err.ConnectionFailed())
	case rdbmstypes.ErrorQueryFailed:
		return errors.New(err.QueryFailed())
	case rdbmstypes.ErrorValueConversionFailed:
		return errors.New(err.ValueConversionFailed())
	default:
		// TODO: not sure if using "Other" as the default is appropriate
		return errors.New(err.Other())
	}
}

func toRow(row []rdbmstypes.DbValue) []any {
	result := make([]any, len(row))
	for i, v := range row {
		switch v.Tag() {
		case rdbmstypes.DbValueBoolean:
			result[i] = v.Boolean()
		case rdbmstypes.DbValueInt8:
			result[i] = v.Int8()
		case rdbmstypes.DbValueInt16:
			result[i] = v.Int16()
		case rdbmstypes.DbValueInt32:
			result[i] = v.Int32()
		case rdbmstypes.DbValueInt64:
			result[i] = v.Int64()
		case rdbmstypes.DbValueUint8:
			result[i] = v.Uint8()
		case rdbmstypes.DbValueUint16:
			result[i] = v.Uint16()
		case rdbmstypes.DbValueUint32:
			result[i] = v.Uint32()
		case rdbmstypes.DbValueUint64:
			result[i] = v.Uint64()
		case rdbmstypes.DbValueFloating32:
			result[i] = v.Floating32()
		case rdbmstypes.DbValueFloating64:
			result[i] = v.Floating64()
		case rdbmstypes.DbValueStr:
			result[i] = v.Str()
		case rdbmstypes.DbValueBinary:
			result[i] = v.Binary()
		case rdbmstypes.DbValueDbNull:
			result[i] = nil
		default:
			panic("unknown value type")
		}
	}

	return result
}

func colTypeToReflectType(typ uint8) reflect.Type {
	switch typ {
	case uint8(rdbmstypes.DbDataTypeBoolean):
		return reflect.TypeOf(false)
	case uint8(rdbmstypes.DbDataTypeInt8):
		return reflect.TypeOf(int8(0))
	case uint8(rdbmstypes.DbDataTypeInt16):
		return reflect.TypeOf(int16(0))
	case uint8(rdbmstypes.DbDataTypeInt32):
		return reflect.TypeOf(int32(0))
	case uint8(rdbmstypes.DbDataTypeInt64):
		return reflect.TypeOf(int64(0))
	case uint8(rdbmstypes.DbDataTypeUint8):
		return reflect.TypeOf(uint8(0))
	case uint8(rdbmstypes.DbDataTypeUint16):
		return reflect.TypeOf(uint16(0))
	case uint8(rdbmstypes.DbDataTypeUint32):
		return reflect.TypeOf(uint32(0))
	case uint8(rdbmstypes.DbDataTypeUint64):
		return reflect.TypeOf(uint64(0))
	case uint8(rdbmstypes.DbDataTypeStr):
		return reflect.TypeOf("")
	case uint8(rdbmstypes.DbDataTypeBinary):
		return reflect.TypeOf(new([]byte))
	case uint8(rdbmstypes.DbDataTypeOther):
		return reflect.TypeOf(new(any)).Elem()
	}
	panic("invalid db column type of " + string(typ))
}

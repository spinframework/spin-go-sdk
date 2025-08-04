package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"

	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	"github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/mysql"
	rdbmstypes "github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/rdbms-types"
	_ "github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/v2.0.0/sqlite"
	"go.bytecodealliance.org/cm"
	_ "go.bytecodealliance.org/cm"
)

// Open returns a new connection to the database.
func Open(address string) *sql.DB {
	return sql.OpenDB(&connector{address})
}

// connector implements driver.Connector.
type connector struct {
	address string
}

// Connect returns a connection to the database.
func (d *connector) Connect(_ context.Context) (driver.Conn, error) {
	return d.Open(d.address)
}

// Driver returns the underlying Driver of the Connector.
func (d *connector) Driver() driver.Driver {
	return d
}

// Open returns a new connection to the database.
func (d *connector) Open(address string) (driver.Conn, error) {
	return &conn{address: address}, nil
}

// conn implements driver.Conn
type conn struct {
	address string
}

var _ driver.Conn = (*conn)(nil)

// Prepare returns a prepared statement, bound to this connection.
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{c: c, query: query}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported by this driver")
}

type stmt struct {
	c     *conn
	query string
}

var _ driver.Stmt = (*stmt)(nil)
var _ driver.ColumnConverter = (*stmt)(nil)   // TODO: remove deprecated?
var _ driver.NamedValueChecker = (*stmt)(nil) // TODO: implement?

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
	params := make([]mysql.ParameterValue, len(args))
	for i := range args {
		params[i] = toWasiValue(args[i])
	}

	results, err, isErr := mysql.Query(s.c.address, s.query, cm.ToList(params)).Result()
	if isErr {
		return nil, toError(&err)
	}

	cols := results.Columns.Slice()
	colNames := make([]string, len(cols))
	colTypes := make([]uint8, len(cols))
	for _, col := range cols {
		colNames = append(colNames, col.Name)
		colTypes = append(colTypes, uint8(col.DataType))
	}

	rowLen := int(results.Rows.Len())
	allRows := make([][]any, rowLen)
	for rowNum, row := range results.Rows.Slice() {
		allRows[rowNum] = toRow(row.Slice())
	}

	rows := &rows{
		columns:    colNames,
		columnType: colTypes,
		len:        rowLen,
		rows:       allRows,
	}

	return rows, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	params := make([]mysql.ParameterValue, len(args))
	for i := range args {
		params[i] = toWasiValue(args[i])
	}
	_, err, isErr := mysql.Execute(s.c.address, s.query, cm.ToList(params)).Result()
	if isErr {
		return nil, toError(&err)
	}

	return &result{}, nil
}

// ColumnConverter returns GlobalParameterConverter to prevent using driver.DefaultParameterConverter.
func (s *stmt) ColumnConverter(_ int) driver.ValueConverter {
	return spindb.GlobalParameterConverter
}

func (s *stmt) NamedValueChecker(v *driver.NamedValue) error {
	// TODO: implement?
}

type result struct{}

func (r result) LastInsertId() (int64, error) {
	return -1, errors.New("LastInsertId is unsupported by this driver")
}

func (r result) RowsAffected() (int64, error) {
	return -1, errors.New("RowsAffected is unsupported by this driver")
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

// Columns return column names.
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

// ColumnTypeScanType return the value type that can be used to scan types into.
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	return colTypeToReflectType(r.columnType[index])
}

func colTypeToReflectType(typ uint8) reflect.Type {
	switch rdbmstypes.DbDataType(typ) {
	case rdbmstypes.DbDataTypeBoolean:
		return reflect.TypeOf(false)
	case rdbmstypes.DbDataTypeInt8:
		return reflect.TypeOf(int8(0))
	case rdbmstypes.DbDataTypeInt16:
		return reflect.TypeOf(int16(0))
	case rdbmstypes.DbDataTypeInt32:
		return reflect.TypeOf(int32(0))
	case rdbmstypes.DbDataTypeInt64:
		return reflect.TypeOf(int64(0))
	case rdbmstypes.DbDataTypeUint8:
		return reflect.TypeOf(uint8(0))
	case rdbmstypes.DbDataTypeUint16:
		return reflect.TypeOf(uint16(0))
	case rdbmstypes.DbDataTypeUint32:
		return reflect.TypeOf(uint32(0))
	case rdbmstypes.DbDataTypeUint64:
		return reflect.TypeOf(uint64(0))
	case rdbmstypes.DbDataTypeStr:
		return reflect.TypeOf("")
	case rdbmstypes.DbDataTypeBinary:
		return reflect.TypeOf(new([]byte))
	case rdbmstypes.DbDataTypeOther:
		return reflect.TypeOf(new(any)).Elem()
	}
	panic("invalid db column type of " + string(typ))
}

func toWasiValue(x any) mysql.ParameterValue {
	switch v := x.(type) {
	case bool:
		return rdbmstypes.ParameterValueBoolean(v)
	case int8:
		return rdbmstypes.ParameterValueInt8(v)
	case int16:
		return rdbmstypes.ParameterValueInt16(v)
	case int32:
		return rdbmstypes.ParameterValueInt32(v)
	case int64:
		return rdbmstypes.ParameterValueInt64(v)
	case int:
		return rdbmstypes.ParameterValueInt64(int64(v))
	case uint8:
		return rdbmstypes.ParameterValueUint8(v)
	case uint16:
		return rdbmstypes.ParameterValueUint16(v)
	case uint32:
		return rdbmstypes.ParameterValueUint32(v)
	case uint64:
		return rdbmstypes.ParameterValueUint64(v)
	case float32:
		return rdbmstypes.ParameterValueFloating32(v)
	case float64:
		return rdbmstypes.ParameterValueFloating64(v)
	case string:
		return rdbmstypes.ParameterValueStr(v)
	case []byte:
		return rdbmstypes.ParameterValueBinary(cm.ToList([]uint8(v)))
	case nil:
		return rdbmstypes.ParameterValueDbNull()
	default:
		panic("unknown value type")
	}
}

func toError(err *mysql.MysqlError) error {
	if err == nil {
		return nil
	}

	return errors.New(err.String())
}

func toRow(row []rdbmstypes.DbValue) []any {
	result := make([]any, len(row))
	for i, v := range row {
		switch v.String() {
		case "boolean":
			result[i] = *v.Boolean()
		case "int8":
			result[i] = *v.Int8()
		case "int16":
			result[i] = *v.Int16()
		case "int32":
			result[i] = *v.Int32()
		case "int64":
			result[i] = *v.Int64()
		case "uint8":
			result[i] = *v.Uint8()
		case "uint16":
			result[i] = *v.Uint16()
		case "uint32":
			result[i] = *v.Uint32()
		case "uint64":
			result[i] = *v.Uint64()
		case "floating32":
			result[i] = *v.Floating32()
		case "floating64":
			result[i] = *v.Floating64()
		case "str":
			result[i] = *v.Str()
		case "binary":
			result[i] = *v.Binary()
		case "db-null":
			result[i] = nil
		default:
			panic("unknown value type")
		}
	}

	return result
}

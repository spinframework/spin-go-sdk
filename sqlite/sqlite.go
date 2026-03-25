package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"

	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	sqlite "github.com/spinframework/spin-go-sdk/v3/internal/fermyon_spin_2_0_0_sqlite"
)

// Open returns a new connection to the database.
func Open(name string) *sql.DB {
	return sql.OpenDB(&connector{name: name})
}

// conn represents a database connection.
type conn struct {
	spinConn sqlite.Connection
}

// Close the connection.
func (c *conn) Close() error {
	return nil
}

// Prepare returns a prepared statement, bound to this connection.
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{conn: c, query: query}, nil
}

// Begin isn't supported.
func (c *conn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported by this driver")
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
	results := sqlite.ConnectionOpen(name)
	if results.IsErr() {
		return nil, toError(results.Err())
	}
	d.conn = &conn{spinConn: *results.Ok()}
	return d.conn, nil
}

// Close closes the connection to the database.
func (d *connector) Close() error {
	if d.conn != nil {
		d.conn.Close()
	}

	return nil
}

type rows struct {
	columns []string
	pos     int
	len     int
	rows    [][]any
	closed  bool
}

var _ driver.Rows = (*rows)(nil)

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
	sqliteParams := make([]sqlite.Value, len(args))
	for i, v := range args {
		sqliteParams[i] = toSqliteValue(v)
	}

	results := s.conn.spinConn.Execute(s.query, sqliteParams)
	if results.IsErr() {
		return nil, toError(results.Err())
	}

	rowLen := len(results.Ok().Rows)
	allRows := make([][]any, rowLen)
	for rowNum, row := range results.Ok().Rows {
		allRows[rowNum] = toRow(row.Values)
	}

	cols := results.Ok().Columns
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = c
	}

	rows := &rows{
		columns: colNames,
		rows:    allRows,
		len:     int(rowLen),
	}
	return rows, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	sqliteParams := make([]sqlite.Value, len(args))
	for i, v := range args {
		sqliteParams[i] = toSqliteValue(v)
	}

	queryResult := s.conn.spinConn.Execute(s.query, sqliteParams)
	if queryResult.IsErr() {
		return &result{}, toError(queryResult.Err())
	}

	return &result{}, nil
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

func toSqliteValue(x any) sqlite.Value {
	switch v := x.(type) {
	case int8:
		return sqlite.MakeValueInteger(int64(v))
	case int16:
		return sqlite.MakeValueInteger(int64(v))
	case int32:
		return sqlite.MakeValueInteger(int64(v))
	case int64:
		return sqlite.MakeValueInteger(v)
	case int:
		return sqlite.MakeValueInteger(int64(v))
	case uint8:
		return sqlite.MakeValueInteger(int64(v))
	case uint16:
		return sqlite.MakeValueInteger(int64(v))
	case uint32:
		return sqlite.MakeValueInteger(int64(v))
	case uint64:
		return sqlite.MakeValueInteger(int64(v))
	case float32:
		return sqlite.MakeValueReal(float64(v))
	case float64:
		return sqlite.MakeValueReal(v)
	case string:
		return sqlite.MakeValueText(v)
	case []byte:
		return sqlite.MakeValueBlob(v)
	case nil:
		return sqlite.MakeValueNull()
	default:
		panic("unknown value type")
	}
}

func toError(err sqlite.Error) error {
	switch err.Tag() {
	case sqlite.ErrorNoSuchDatabase:
		return errors.New("no such database")
	case sqlite.ErrorAccessDenied:
		return errors.New("access denied")
	case sqlite.ErrorInvalidConnection:
		return errors.New("invalid connection")
	case sqlite.ErrorDatabaseFull:
		return errors.New("database full")
	case sqlite.ErrorIo:
		return errors.New(err.Io())
	default:
		panic("unreachable code")
	}
}

func toRow(row []sqlite.Value) []any {
	result := make([]any, len(row))
	for i, v := range row {
		switch v.Tag() {
		case sqlite.ValueInteger:
			result[i] = v.Integer()
		case sqlite.ValueReal:
			result[i] = v.Real()
		case sqlite.ValueText:
			result[i] = v.Text()
		case sqlite.ValueBlob:
			result[i] = v.Blob()
		case sqlite.ValueNull:
			result[i] = nil
		default:
			panic("unreachable code")
		}
	}

	return result
}

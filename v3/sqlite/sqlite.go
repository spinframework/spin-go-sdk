package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	spindb "github.com/spinframework/spin-go-sdk/v3/internal/db"
	"github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/v2.0.0/sqlite"
	"go.bytecodealliance.org/cm"
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
	d.conn = &conn{
		spinConn: *results.OK(),
	}
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
	numRows int
	rows    [][]any
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
	r.numRows = 0
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
	return r.pos < r.numRows
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

func toRow(row []sqlite.Value) []any {
	ret := make([]any, len(row))
	for i, v := range row {
		switch v.String() {
		case "integer":
			ret[i] = *v.Integer()
		case "real":
			ret[i] = *v.Real()
		case "text":
			ret[i] = *v.Text()
		case "blob":
			// TODO: check this
			ret[i] = *v.Blob()
		case "null":
			ret[i] = nil
		default:
			panic("unknown value type")
		}
	}
	return ret
}

func toWasiValue(x any) sqlite.Value {
	switch v := x.(type) {
	case int:
		return sqlite.ValueInteger(int64(v))
	case int64:
		return sqlite.ValueInteger(v)
	case float64:
		return sqlite.ValueReal(v)
	case string:
		return sqlite.ValueText(v)
	case []byte:
		return sqlite.ValueBlob(cm.ToList([]uint8(v)))
	case nil:
		return sqlite.ValueNull()
	default:
		panic("unknown value type")
	}
}

// Query executes a query that may return rows, such as a SELECT.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	params := make([]sqlite.Value, len(args))
	for i := range args {
		params[i] = toWasiValue(args[i])
	}
	results, err, isErr := s.conn.spinConn.Execute(s.query, cm.ToList(params)).Result()
	if isErr {
		return nil, toError(&err)
	}

	cols := results.Columns.Slice()

	rowLen := results.Rows.Len()
	allrows := make([][]any, rowLen)
	for rownum, row := range results.Rows.Slice() {
		allrows[rownum] = toRow(row.Values.Slice())
	}
	rows := &rows{
		columns: cols,
		rows:    allrows,
		numRows: int(rowLen),
	}
	return rows, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or
// UPDATE.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	params := make([]sqlite.Value, len(args))
	for i := range args {
		params[i] = toWasiValue(args[i])
	}
	_, err, isErr := s.conn.spinConn.Execute(s.query, cm.ToList(params)).Result()
	if isErr {
		return nil, toError(&err)
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

func toError(err *sqlite.Error) error {
	if err == nil {
		return nil
	}
	if err.String() == "io" {
		return fmt.Errorf("io: %s", *err.IO())
	}
	return errors.New(err.String())
}

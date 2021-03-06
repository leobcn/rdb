// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package rdb is a relational database interface for products that use SQL.
// Multiple sequential results and types are supported.
//
// Each logical action (for example http request) must have a context
// associated with it that should be cancelled when it is complete. This ensures
// individual connections are never leaked.
package rdb // import "github.com/kardianos/rdb"

import (
	"bytes"

	"golang.org/x/net/context"
)

// Isolation is used to set the isolation of a transaction.
type Isolation byte

// Isolation levels used in databases.
// Not all databases support all isolation levels.
const (
	IsoDefault Isolation = iota
	IsoReadUncommited
	IsoReadCommited
	IsoWriteCommited
	IsoRepeatableRead
	IsoSerializable
	IsoSnapshot
	IsoLinearizable
)

// Queryer queries the database.
type Queryer interface {
	// Query runs a command against the system. The supplied context
	// will close the query when it is cancelled.
	Query(ctx context.Context, cmd *Command, params ...Param) Next
}

// Preparer prepares a statement for execution.
type Preparer interface {
	// Prepare takes a command and returns a prepared statement.
	// The statement is closed when the supplied contex to Prepare
	// is cancelled.
	Prepare(ctx context.Context, cmd *Command) (Statement, error)
}

// Pool represents a pool of database connections.
type Pool interface {
	// BeginLevel starts a Transaction with the specified isolation level.
	// If the context is cancelled before the transaction is committed, the
	// transaction attempts to roll back before closing.
	Begin(ctx context.Context, iso Isolation) (Transaction, error)

	// Close the connection pool.
	Close()

	// Connection returns a dedicated database connection from the connection pool.
	// When the context is cancelled the connection will be closed and returned
	// to the connection pool.
	Connection(ctx context.Context) (Connection, error)

	// Will attempt to connect to the database and disconnect. Must not impact any existing connections.
	Ping(ctx context.Context) error

	// Status of the current pool.
	Status() PoolStatus

	Queryer
	Preparer
}

// SQLError represents a sql error.
// Useful for showing the exact line number of a query error.
type SQLError interface {
	Error() string
	LineNumber() int
	ErrorCode() int
}

// ErrorList represents a list of errors.
type ErrorList struct {
	List []error
}

// Error returns the error message.
func (err ErrorList) Error() string {
	buf := &bytes.Buffer{}
	for index, item := range err.List {
		if index != 0 {
			buf.WriteRune('\n')
		}
		buf.WriteString(item.Error())
	}
	return buf.String()
}

// Connection represents a single connection to the database.
type Connection interface {
	// Close returns the connection to the connection pool
	Close()

	Queryer
}

// Transaction represents a single transaction connection to the database.
type Transaction interface {
	Queryer

	// Rollback to an existing savepoint. Commit or Rollback should still
	// be called after calling RollbackTo.
	RollbackTo(ctx context.Context, savepoint string) error

	// Create a save point in the transaction.
	SavePoint(ctx context.Context, name string) error

	// Commit the transaction.
	Commit(ctx context.Context) error
}

// Statement represents a prepared statement. On most systems this takes out
// a resource on the server and should be closed by closing the associated context
// (see Preparer). It is not advised to use a Statement scoped to an application
// or a long lived object, as a database restart will invalidate all statements.
type Statement interface {
	Exec(ctx context.Context, params ...Param) Next
}

// PoolStatus is the basic interface for database pool information.
type PoolStatus interface {
	Capacity() int
	Available() int
}

// Row is a way to access data from a cached row.
type Row interface {
	Get(name string) interface{}
	Getx(index int) interface{}
	Into(name string, value interface{}) Row
	Intox(index int, value interface{}) Row
}

// Next proceeds to the next Result or buffers the entierty of the next result.
// If there is an error with the query that returns Next, it will be returned in
// the error value in any of it's methods.
// The connection is returned after the last result or buffer has been read or
// Close is explicitly called.
type Next interface {
	Result() (Result, error)

	Buffer() (*Buffer, error)
	BufferSet() (BufferSet, error)

	// Close will allow any connection to return to the pool.
	// Any subsequent calls to Result or Buffer will return an error.
	// If the query context is cancelled the result is also closed and the
	// connection returned to the pool.
	Close() error
}

// Result provides a way to iterate over a query result.
type Result interface {
	// Prep and Prepx should be called before Scan. If value is a io.Writer
	// and the driver supports it, the driver may write directly into the value.
	// Prepared values are not written to the Row buffer returned in Scan.
	Prep(name string, value interface{}) Result
	Prepx(index int, value interface{}) Result

	// Scan will read a Row, populate any values used in Prep.
	// Row will be nil when last row has been read.
	Scan() (Row, error)

	// Return the column schema for result.
	Schema() Schema

	// Close will allow any connection to return to the pool.
	// Same as calling Next.Close().
	Close() error
}

// Param provides values into a query.
// If the Name field is not specified is not specified, then the order
// of the parameter should be used if the driver supports it.
type Param struct {
	// Optional parameter name.
	Name string

	// Parameter Type. Drivers may be able to infer this type.
	// Check the driver documentation used for more information.
	Type Type

	// Set to true if the parameter is an output parameter.
	// If true, the value member should be provided through a pointer.
	Out bool

	// Do not send this value to the trace.
	NoTrace bool

	// Paremeter Length. Useful for variable length types that may check truncation.
	Length int

	// Value for input parameter.
	// If the value is an io.Reader it will read the value directly to the wire.
	Value interface{}
}

// Command represents a SQL command and can be used from many different
// queries at the same time.
// The Command MUST be reused if the Prepare field is true.
type Command struct {
	// The SQL to be used in the command.
	SQL string

	// TextAsBytes should be set to true for strings to be returned as
	// bytes to reduce allocations.
	TextAsBytes bool

	// If set to true silently truncates text longer then the field.
	// If this is set to false text truncation will result in an error.
	TruncLongText bool

	// Set the isolation level for the query or transaction.
	Isolation Isolation

	// Optional name of the command. May be used if logging.
	Name string
}

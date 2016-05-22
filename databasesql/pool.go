// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package databasesql provides a wrapper for "database/sql" drivers.
//
// TODO (DT): complete wrapping this package.
//
//   import _ "github.com/kardianos/rdb/databasesql"
//   import _ "my-database-sql-driver"
package databasesql // import "github.com/kardianos/rdb/databasesql"

import (
	"database/sql"
	"errors"

	"github.com/kardianos/rdb"
	"golang.org/x/net/context"
)

// Pool implements rdb.Pool.
type Pool struct {
	DB *sql.DB
}

// Query sends a database query.
// TODO (DT): Finish Query wrapper.
// TODO (DT): Might want a config parameter that sets if named parameters are
// substituted.
func (p *Pool) Query(ctx context.Context, cmd *rdb.Command, params ...rdb.Param) rdb.Next {
	return nil
}

// Begin starts a transaction.
// TODO (DT): Finish Transaction wrapper.
func (p *Pool) Begin(ctx context.Context, iso rdb.Isolation) (rdb.Transaction, error) {
	return nil, nil
}

// Close the connection pool.
func (p *Pool) Close() {
	p.DB.Close()
}

var errConnectionNotSupported = errors.New("Connection not supported for database/sql based driver.")

// Connection is not supported for database/sql drivers.
func (p *Pool) Connection() (rdb.Connection, error) {
	return nil, errConnectionNotSupported
}

// Ping the server to ensure it is alive.
func (p *Pool) Ping(ctx context.Context) error {
	return p.DB.Ping()
}

// Status returns the number of connections in the pool.
func (p *Pool) Status() rdb.PoolStatus {
	return p
}

// SetTrace sets a trace on the query process.
func (p *Pool) SetTrace(rdb.Tracer) {

}

// Capacity returns the same as Available for database/sql drivers.
func (p *Pool) Capacity() int {
	// TODO (DT): No way to get true capacity...
	return p.DB.Stats().OpenConnections
}

// Available open connections.
func (p *Pool) Available() int {
	return p.DB.Stats().OpenConnections
}

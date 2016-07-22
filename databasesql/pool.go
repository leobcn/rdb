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
	"context"
	"database/sql"
	"sync"

	"github.com/kardianos/rdb"
	"github.com/pkg/errors"
)

var (
	errTODO         = errors.New("TODO")
	errNotSupported = errors.New("not supported for database/sql based driver")
)

// Pool implements rdb.Pool.
type Pool struct {
	DB *sql.DB

	lk   sync.RWMutex
	stmt map[*rdb.Command]*sql.Stmt
}

type next struct {
	err    error
	rows   *sql.Rows
	params []rdb.Param
}

func (n *next) Result() (rdb.Result, error) {
	return nil, errTODO
}
func (n *next) Buffer() (*rdb.Buffer, error) {
	return nil, errTODO
}
func (n *next) BufferSet() (rdb.BufferSet, error) {
	return nil, errNotSupported
}
func (n *next) Close() error {
	return errTODO
}

func (p *Pool) makeArgs(params []rdb.Param) []interface{} {
	return nil
}

// Query sends a database query.
// TODO (DT): Finish Query wrapper.
// TODO (DT): Might want a config parameter that sets if named parameters are
// substituted.
func (p *Pool) Query(ctx context.Context, cmd *rdb.Command, params ...rdb.Param) rdb.Next {
	if cmd.Prepare {
		var err error
		p.lk.RLock()
		s, found := p.stmt[cmd]
		p.lk.Unlock()

		if !found {
			// TODO (DT): add timeout with ctx.
			s, err = p.DB.Prepare(cmd.SQL)
			if err != nil {
				return &next{err: errors.Wrapf(err, "prepare sql %q", cmd.Name)}
			}
			p.lk.Lock()
			p.stmt[cmd] = s
			p.lk.Unlock()
		}
		rows, err := s.Query(p.makeArgs(params)...)
		if err != nil {
			p.lk.Lock()
			delete(p.stmt, cmd)
			p.lk.Unlock()
		}
		return &next{err: err, rows: rows, params: params}
	}
	rows, err := p.DB.Query(cmd.SQL, p.makeArgs(params)...)
	return &next{err: err, rows: rows, params: params}
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

// Connection is not supported for database/sql drivers.
func (p *Pool) Connection() (rdb.Connection, error) {
	return nil, errNotSupported
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
	// TODO (DT): No way to get true capacity.
	return p.DB.Stats().OpenConnections
}

// Available open connections.
func (p *Pool) Available() int {
	return p.DB.Stats().OpenConnections
}

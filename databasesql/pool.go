// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package databasesql provides a wrapper for "database/sql" drivers.
//
// TODO (DT): complete wrapping this package.
// Limitations:
//   Cannot cancel a query in progress due to underlying database/sql limitations.
//   Does not respect rdb.Command.TextAsBytes parameter as the result data type is not available.
//
//   import _ "github.com/kardianos/rdb/databasesql"
//   import _ "my-database-sql-driver"
package databasesql // import "github.com/kardianos/rdb/databasesql"

import (
	"database/sql"

	"github.com/kardianos/rdb"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	errTODO         = errors.New("TODO")
	errNotSupported = errors.New("not supported for database/sql based driver")
)

// Pool implements rdb.Pool.
type Pool struct {
	DB *sql.DB
}

type next struct {
	ctx    context.Context
	err    error
	rows   *sql.Rows
	cancel func()

	textAsBytes bool
}

type statement struct {
	ctx  context.Context
	stmt *sql.Stmt

	truncateLongText bool
	textAsBytes      bool
}

type transaction struct {
	ctx context.Context
	tx  *sql.Tx
}
type result struct {
	rows *sql.Rows
}

func (n *next) init() {
	if n.err != nil {
		return
	}
	n.ctx, n.cancel = context.WithCancel(n.ctx)
	go func() {
		select {
		case <-n.ctx.Done():
			n.rows.Close()
		}
	}()
}
func (n *next) Prep(name string, value interface{}) rdb.Result {
	panic(errTODO)
	return n
}
func (n *next) Prepx(index int, value interface{}) rdb.Result {
	panic(errTODO)
	return n
}
func (n *next) Scan() (rdb.Row, error) {
	return nil, errTODO
}
func (n *next) Schema() rdb.Schema {
	names, _ := n.rows.Columns()
	sch := make([]rdb.Column, len(names))
	for i, name := range names {
		sch[i] = rdb.Column{
			Name:  name,
			Index: i,
		}
	}
	return sch
}

func (n *next) Result() (rdb.Result, error) {
	return n, n.err
}
func (n *next) Buffer() (*rdb.Buffer, error) {
	return nil, errTODO
}

// BufferSet will only return a single buffer for database/sql drivers.
func (n *next) BufferSet() (rdb.BufferSet, error) {
	buf, err := n.Buffer()
	if buf != nil {
		return rdb.BufferSet{buf}, err
	}
	return nil, err
}

func (n *next) Close() error {
	n.cancel()
	return n.err
}

func (st *statement) Exec(ctx context.Context, params ...rdb.Param) rdb.Next {
	if err := ctx.Err(); err != nil {
		return &next{err: err}
	}
	rows, err := st.stmt.Query(makeArgs(st.truncateLongText, params))
	if cerr := ctx.Err(); cerr != nil {
		rows.Close()
		err = cerr
	}
	n := &next{err: err, rows: rows, ctx: ctx, textAsBytes: st.textAsBytes}
	n.init()
	return n
}

func (tx *transaction) Query(ctx context.Context, cmd *rdb.Command, params ...rdb.Param) rdb.Next {
	if err := ctx.Err(); err != nil {
		return &next{err: err}
	}
	rows, err := tx.tx.Query(cmd.SQL, makeArgs(cmd.TruncLongText, params)...)
	if cerr := ctx.Err(); cerr != nil {
		rows.Close()
		err = cerr
	}
	n := &next{err: err, rows: rows, ctx: ctx, textAsBytes: cmd.TextAsBytes}
	n.init()
	return n
}
func (tx *transaction) RollbackTo(ctx context.Context, name string) error {
	return tx.tx.Rollback()
}

// SavePoint is not supported by database/sql.
func (tx *transaction) SavePoint(ctx context.Context, name string) error {
	return errNotSupported
}
func (tx *transaction) Commit(ctx context.Context) error {
	return tx.tx.Commit()
}

func makeArgs(tuncLongText bool, params []rdb.Param) []interface{} {
	out := make([]interface{}, len(params))
	for i := range params {
		out[i] = params[i].Value
	}
	return out
}

// Query sends a database query.
func (p *Pool) Query(ctx context.Context, cmd *rdb.Command, params ...rdb.Param) rdb.Next {
	if err := ctx.Err(); err != nil {
		return &next{err: err}
	}
	rows, err := p.DB.Query(cmd.SQL, makeArgs(cmd.TruncLongText, params)...)
	if cerr := ctx.Err(); cerr != nil {
		rows.Close()
		err = cerr
	}
	n := &next{err: err, rows: rows, ctx: ctx, textAsBytes: cmd.TextAsBytes}
	n.init()
	return n
}

func (p *Pool) Prepare(ctx context.Context, cmd *rdb.Command) (rdb.Statement, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s, err := p.DB.Prepare(cmd.SQL)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		s.Close()
		return nil, err
	}
	st := &statement{
		ctx:  ctx,
		stmt: s,

		truncateLongText: cmd.TruncLongText,
		textAsBytes:      cmd.TextAsBytes,
	}
	return st, nil
}

// Begin starts a transaction.
func (p *Pool) Begin(ctx context.Context, iso rdb.Isolation) (rdb.Transaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tx, err := p.DB.Begin()
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	t := &transaction{
		ctx: ctx,
		tx:  tx,
	}
	return t, nil
}

// Close the connection pool.
func (p *Pool) Close() {
	p.DB.Close()
}

// Connection is not supported for database/sql drivers.
func (p *Pool) Connection(ctx context.Context) (rdb.Connection, error) {
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

// Capacity returns the same as Available for database/sql drivers.
func (p *Pool) Capacity() int {
	// No way to get true capacity.
	return p.DB.Stats().OpenConnections
}

// Available open connections.
func (p *Pool) Available() int {
	return p.DB.Stats().OpenConnections
}

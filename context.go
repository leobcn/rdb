// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"errors"

	"context"
)

type key int

const (
	poolKey key = 0
)

// NewContext wraps a Pool in a context.
func NewContext(ctx context.Context, pool Pool) context.Context {
	return context.WithValue(ctx, poolKey, pool)
}

// FromContext returns a Pool from a context.
func FromContext(ctx context.Context) (Pool, bool) {
	pool, has := ctx.Value(poolKey).(Pool)
	return pool, has
}

var (
	errNoPoolContext = errors.New("No Pool in context")
)

type nextError struct {
	err error
}

func (next nextError) Result() (Result, error) {
	return nil, next.err
}
func (next nextError) Buffer() (*Buffer, error) {
	return nil, next.err
}
func (next nextError) BufferSet() (BufferSet, error) {
	return nil, next.err
}
func (next nextError) Close() error {
	return next.err
}

// Query unwraps the pool from context.
func Query(ctx context.Context, cmd *Command, params ...Param) Next {
	pool, has := FromContext(ctx)
	if !has {
		return nextError{err: errNoPoolContext}
	}
	return pool.Query(ctx, cmd, params...)
}

// Begin starts a transaction from pool in context.
func Begin(ctx context.Context, iso Isolation) (Transaction, error) {
	pool, has := FromContext(ctx)
	if !has {
		return nil, errNoPoolContext
	}
	return pool.Begin(ctx, iso)
}

// QuerySet runs command and returns a list of buffers and closes any connections
// it has opened before returning.
func QuerySet(ctx context.Context, cmd *Command, params ...Param) (BufferSet, error) {
	nx := Query(ctx, cmd, params...)
	defer nx.Close()

	var set BufferSet = make([]*Buffer, 0, 3)
	for {
		b, err := nx.Buffer()
		if err != nil {
			return set, err
		}
		if b == nil {
			return set, nil
		}
		set = append(set, b)
	}
}

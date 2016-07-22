// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package databasesql

import (
	"context"
	"database/sql"

	"github.com/kardianos/rdb"
)

func init() {
	o := &Opener{}
	rdb.RegisterOpener(o)
}

// Opener implements an rdb.Opener.
type Opener struct{}

// CanOpen returns true if this driver can open it.
func (o *Opener) CanOpen(config *rdb.Config) bool {
	for _, name := range sql.Drivers() {
		if name == config.DriverName {
			return true
		}
	}
	return false
}

// Open opens the config.
func (o *Opener) Open(ctx context.Context, config *rdb.Config) (rdb.Pool, error) {
	db, err := sql.Open(config.DriverName, config.Raw)
	if err != nil {
		return nil, err
	}
	pool := &Pool{
		DB: db,
	}
	return pool, nil
}

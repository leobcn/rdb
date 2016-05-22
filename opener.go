// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"errors"
	"sync"

	"golang.org/x/net/context"
)

var (
	errNoOpenerFound = errors.New("Now opener found that could open config")
)

// Open a new database connection pool.
func Open(ctx context.Context, config *Config) (Pool, error) {
	var found Opener

	openerSync.RLock()
	for _, o := range openerList {
		if can := o.CanOpen(config); can {
			found = o
			break
		}
	}
	openerSync.RUnlock()
	if found == nil {
		return nil, errNoOpenerFound
	}
	return found.Open(ctx, config)
}

/*
	driver.Interface refs rdb
	driver injects opener into rdb
	(maybe) a sql driver can choose to use "driver".Pool implementation or use it's own
*/

// Opener is used to create new database connection pools.
type Opener interface {
	CanOpen(config *Config) bool
	Open(ctx context.Context, config *Config) (Pool, error)
}

var (
	openerSync = sync.RWMutex{}
	openerList = make([]Opener, 0, 3)
)

// RegisterOpener database driver pools should call this to register their opener.
func RegisterOpener(opener Opener) {
	openerSync.Lock()
	defer openerSync.Unlock()

	openerList = append(openerList, opener)
}

// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Buffer provides a database table buffer.
type Buffer struct {
	Name   string
	Row    []Row
	Schema Schema
}

// BufferSet is a list of Buffers.
type BufferSet []*Buffer

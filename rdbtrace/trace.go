// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdbtrace

import (
	"time"

	"github.com/kardianos/rdb"
)

// Tracer records a new query event.
type Tracer interface {
	// Event is called when a new query is invoked. Event
	// should return null if no further tracing should be performed.
	//
	// Event is called for every Pool.Begin, Pool.Query, Pool.Prepare, or Pool.Connection.
	Event() TraceEvent
}

// TraceEvent represents a single query, transaction, or connection.
// If any event returns nil no further tracing will be performed on the event.
type TraceEvent interface {
	// QueryBegin is called when a new query is performed. May be called multiple
	// times in the case of a transaction or connection.
	QueryBegin(at time.Time, cmd *rdb.Command, params []rdb.Param) TraceEvent

	// QueryEnd is called when the query returns from the database server.
	QueryEnd(at time.Time) TraceEvent

	// Reports a message from the server.
	Message(at time.Time, messge string) TraceEvent

	// Reports an error.
	Error(at time.Time, err error) TraceEvent

	// The trace event is closed. This is will be called when the
	// underlying connection is returned to the pool.
	Close(at time.Time)
}

// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// +build go1.7

package rdb_test

import (
	"log"
	"net/http"
	"time"

	"github.com/kardianos/rdb"
	"golang.org/x/net/context"
)

func Example() {
	pool := func() rdb.Pool {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		conf, err := rdb.ParseConfigURL(`sqlite:///srv/folder/file.sqlite3?opt1=valA&opt2=valB`)
		if err != nil {
			log.Fatal(err)
		}
		pool, err := rdb.Open(ctx, conf)
		if err != nil {
			log.Fatal(err)
		}
		return pool
	}()

	// In go1.7+ context can be stored in the "http.Request.Context()" method.
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		table, err := rdb.Query(r.Context(), &rdb.Command{
			SQL: `select ID, Message from Log;`,
		}).Buffer()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range table.Row {
			_ = row.Get("ID").(int64)
			_ = row.Get("Message").(string)
		}
	}
	http.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ctx = rdb.NewContext(ctx, pool) // Store the pool in the context.

		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		r = r.WithContext(ctx)

		switch r.URL.Path {
		default:
			http.Error(w, "not found", http.StatusNotFound)
		case "/api/db/log":
			logHandler(w, r) // In go1.7+ context can be stored in "r".
		}
	})
}

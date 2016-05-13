// Copyright 2016 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Schema is a list of columns and related methods.
type Schema struct {
	Column []Column
}

// Column information as reported by the database.
type Column struct {
	Name    string // Columnn name.
	Index   int    // Column zero based index as appearing in result.
	Type    Type   // The data type as reported from the driver.
	Generic Type   // The generic data type as reported from the driver.

	// Length of the column as it makes sense per type.
	// If Length is negative assume unlimited length.
	Length int

	Nullable  bool // True if the column type can be null.
	Key       bool // True if the column is part of the key.
	Serial    bool // True if the column is auto-incrementing.
	Precision int  // For decimal types, the precision.
	Scale     int  // For types with scale, including decimal.
}

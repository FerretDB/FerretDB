// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fsql

import (
	"database/sql"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Rows wraps [*database/sql.Rows] with resource tracking.
//
// It exposes the subset of *sql.Rows methods we use.
type Rows struct {
	sqlRows *sql.Rows
	token   *resource.Token
}

// wrapRows creates new Rows.
func wrapRows(rows *sql.Rows) *Rows {
	if rows == nil {
		return nil
	}

	res := &Rows{
		sqlRows: rows,
		token:   resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Close calls [*sql.Rows.Close].
func (rows *Rows) Close() error {
	resource.Untrack(rows, rows.token)
	return rows.sqlRows.Close()
}

// Err calls [*sql.Rows.Err].
func (rows *Rows) Err() error {
	return rows.sqlRows.Err()
}

// Next calls [*sql.Rows.Next].
func (rows *Rows) Next() bool {
	return rows.sqlRows.Next()
}

// Scan calls [*sql.Rows.Scan].
func (rows *Rows) Scan(dest ...any) error {
	return rows.sqlRows.Scan(dest...)
}

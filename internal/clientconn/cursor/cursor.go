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

// Package cursor provides access to cursor registry.
package cursor

import (
	"sync"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// The implementation of the cursor and registry is quite complicated and entangled.
// That's because there are many cases when cursor / iterator / underlying database connection
// must be closed to free resources, including when no handler and backend code is running;
// for example, when the client disconnects between `getMore` commands.
// At the same time, we want to shift complexity away from the handler and from backend implementations
// because they are already quite complex.
// The current design enables ease of use at the expense of the implementation complexity.

// Cursor allows clients to iterate over a result set.
//
// It implements types.DocumentsIterator interface by wrapping another iterator with documents
// with additional metadata and registration in the registry.
//
// Closing the cursor removes it from the registry.
type Cursor struct {
	// the order of fields is weird to make the struct smaller due to alignment

	iter       types.DocumentsIterator
	r          *Registry
	token      *resource.Token
	DB         string
	Collection string
	Username   string
	ID         int64
	closeOnce  sync.Once
}

// newCursor creates a new cursor.
func newCursor(id int64, db, collection, username string, iter types.DocumentsIterator, r *Registry) *Cursor {
	c := &Cursor{
		ID:         id,
		DB:         db,
		Collection: collection,
		Username:   username,
		iter:       iter,
		r:          r,
		token:      resource.NewToken(),
	}

	resource.Track(c, c.token)

	return c
}

// Next implements types.DocumentsIterator interface.
func (c *Cursor) Next() (struct{}, *types.Document, error) {
	return c.iter.Next()
}

// Close implements types.DocumentsIterator interface.
func (c *Cursor) Close() {
	c.closeOnce.Do(func() {
		c.r.delete(c.ID)
		c.iter.Close()
		resource.Untrack(c, c.token)
	})
}

// check interfaces
var (
	_ types.DocumentsIterator = (*Cursor)(nil)
)

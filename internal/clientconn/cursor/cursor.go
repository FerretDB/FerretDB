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
//
// The implementation of the cursor and registry is quite complicated and entangled.
// That's because there are many cases when cursor / iterator / underlying database connection
// must be closed to free resources, including when no handler and backend code is running;
// for example, when the client disconnects between `getMore` commands.
// At the same time, we want to shift complexity away from the handler and from backend implementations
// because they are already quite complex.
// The current design enables ease of use at the expense of the implementation complexity.
package cursor

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

//go:generate ../../../bin/stringer -linecomment -type Type

type Type int

const (
	_ Type = iota
	Normal
	Tailable
	TailableAwait
)

// Cursor allows clients to iterate over a result set.
//
// It implements types.DocumentsIterator interface by wrapping another iterator with documents
// with additional metadata and registration in the registry.
//
// Closing the cursor removes it from the registry.
type Cursor struct {
	// the order of fields is weird to make the struct smaller due to alignment

	ID   int64
	iter types.DocumentsIterator
	*NewParams
	r *Registry

	token        *resource.Token
	created      time.Time
	closed       chan struct{}
	closeOnce    sync.Once
	lastRecordID atomic.Int64
}

// newCursor creates a new cursor.
func newCursor(id int64, iter types.DocumentsIterator, params *NewParams, r *Registry) *Cursor {
	c := &Cursor{
		ID:        id,
		iter:      iter,
		NewParams: params,
		r:         r,
		token:     resource.NewToken(),
		created:   time.Now(),
		closed:    make(chan struct{}),
	}

	resource.Track(c, c.token)

	return c
}

func (c *Cursor) Reset(iter types.DocumentsIterator) error {
	if c.Type != Tailable && c.Type != TailableAwait {
		panic("Reset called on non-tailable cursor")
	}

	c.iter = iter

	recordID := c.lastRecordID.Load()
	for {
		_, doc, err := c.Next()
		if err != nil {
			// FIXME
			return lazyerrors.Error(err)
		}

		if doc.RecordID() == recordID {
			return nil
		}
	}
}

// Next implements types.DocumentsIterator interface.
func (c *Cursor) Next() (struct{}, *types.Document, error) {
	zero, doc, err := c.iter.Next()
	if doc != nil {
		recordID := doc.RecordID()
		c.lastRecordID.Store(recordID)

		if c.ShowRecordID {
			doc.Set("$recordId", recordID)
		}
	}

	return zero, doc, err
}

// Close implements types.DocumentsIterator interface.
func (c *Cursor) Close() {
	c.closeOnce.Do(func() {
		c.iter.Close()
		c.iter = nil

		c.r.delete(c)

		close(c.closed)

		resource.Untrack(c, c.token)
	})
}

// check interfaces
var (
	_ types.DocumentsIterator = (*Cursor)(nil)
)

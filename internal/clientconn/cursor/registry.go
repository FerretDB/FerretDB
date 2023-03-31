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

package cursor

import (
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

// Global last cursor ID.
var lastCursorID atomic.Uint32

func init() {
	// to make debugging easier
	if !debugbuild.Enabled {
		lastCursorID.Store(rand.Uint32())
	}
}

// Registry stores cursors.
//
// TODO add cleanup
// TODO add metrics
//
//nolint:vet // for readability
type Registry struct {
	rw sync.RWMutex
	m  map[string]map[int64]*cursor // username -> ID -> cursor
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		m: map[string]map[int64]*cursor{},
	}
}

// Cursor returns stored cursor by username and ID, or nil.
func (r *Registry) Cursor(username string, id int64) *cursor {
	r.rw.RLock()
	defer r.rw.RUnlock()

	if u := r.m[username]; u == nil {
		return nil
	}

	return r.m[username][id]
}

// StoreCursor stores cursor and return its ID.
func (r *Registry) StoreCursor(username string, c *cursor) int64 {
	r.rw.Lock()
	defer r.rw.Unlock()

	if u := r.m[username]; u == nil {
		r.m[username] = map[int64]*cursor{}
	}

	// use global, sequential, positive, short cursor IDs to make debugging easier
	var id int64
	for id == 0 || r.m[username][id] != nil {
		id = int64(lastCursorID.Add(1))
	}

	r.m[username][id] = c

	return id
}

// DeleteCursor closes and deletes cursor.
func (r *Registry) DeleteCursor(username string, id int64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	if u := r.m[username]; u == nil {
		return
	}

	c := r.m[username][id]
	if c == nil {
		return
	}

	c.Close()
	delete(r.m[username], id)

	if len(r.m[username]) == 0 {
		delete(r.m, username)
	}
}

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

// Package conninfo provides access to connection-specific information.
package conninfo

import (
	"context"
	"math/rand"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// contextKey is a special type to represent context.WithValue keys a bit more safely.
type contextKey struct{}

var (
	// Context key for WithConnInfo/Get.
	connInfoKey = contextKey{}

	// Keeps track on all ConnInfo objects.
	connInfoProfiles = pprof.NewProfile("github.com/FerretDB/FerretDB/internal/clientconn/conninfo.connInfo")

	// Global last cursor ID.
	lastCursorID atomic.Uint32
)

func init() {
	// to make debugging easier
	if !debugbuild.Enabled {
		lastCursorID.Store(rand.Uint32())
	}
}

// ConnInfo represents connection info.
type ConnInfo struct {
	PeerAddr string

	rw       sync.RWMutex
	cursors  map[int64]Cursor
	username string
	password string

	stack []byte
}

// NewConnInfo return a new ConnInfo.
func NewConnInfo() *ConnInfo {
	connInfo := &ConnInfo{
		cursors: map[int64]Cursor{},
		stack:   debugbuild.Stack(),
	}

	connInfoProfiles.Add(connInfo, 1)

	runtime.SetFinalizer(connInfo, func(connInfo *ConnInfo) {
		msg := "ConnInfo.Close() has not been called"
		if connInfo.stack != nil {
			msg += "\nConnInfo created by " + string(connInfo.stack)
		}

		panic(msg)
	})

	return connInfo
}

// Close frees resources.
func (connInfo *ConnInfo) Close() {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	connInfoProfiles.Remove(connInfo)

	runtime.SetFinalizer(connInfo, nil)

	for _, c := range connInfo.cursors {
		c.Iter.Close()
	}
}

// Cursor allows clients to iterate over a result set.
type Cursor struct {
	Iter   iterator.Interface[int, *types.Document]
	Filter *types.Document
}

// Cursor returns cursor by ID, or nil.
func (connInfo *ConnInfo) Cursor(id int64) *Cursor {
	connInfo.rw.RLock()
	defer connInfo.rw.RUnlock()

	c, ok := connInfo.cursors[id]
	if !ok {
		return nil
	}

	return &c
}

// StoreCursor stores cursor and return its ID.
func (connInfo *ConnInfo) StoreCursor(iter iterator.Interface[int, *types.Document], filter *types.Document) int64 {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	var id int64

	// use global, sequential, positive, short cursor IDs to make debugging easier
	for {
		id = int64(lastCursorID.Add(1))
		if _, ok := connInfo.cursors[id]; id != 0 && !ok {
			break
		}
	}

	connInfo.cursors[id] = Cursor{
		Iter:   iter,
		Filter: filter,
	}

	return id
}

// DeleteCursor deletes cursor by ID, closing its iterator.
func (connInfo *ConnInfo) DeleteCursor(id int64) {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	c := connInfo.cursors[id]

	c.Iter.Close()

	delete(connInfo.cursors, id)
}

// Auth returns stored username and password.
func (connInfo *ConnInfo) Auth() (username, password string) {
	connInfo.rw.RLock()
	defer connInfo.rw.RUnlock()

	return connInfo.username, connInfo.password
}

// SetAuth stores username and password.
func (connInfo *ConnInfo) SetAuth(username, password string) {
	connInfo.rw.Lock()
	defer connInfo.rw.Unlock()

	connInfo.username = username
	connInfo.password = password
}

// WithConnInfo returns a new context with the given ConnInfo.
func WithConnInfo(ctx context.Context, connInfo *ConnInfo) context.Context {
	return context.WithValue(ctx, connInfoKey, connInfo)
}

// Get returns the ConnInfo value stored in ctx.
func Get(ctx context.Context) *ConnInfo {
	value := ctx.Value(connInfoKey)
	if value == nil {
		panic("connInfo is not set in context")
	}

	connInfo, ok := value.(*ConnInfo)
	if !ok {
		panic("connInfo is set in context with wrong value type")
	}

	if connInfo == nil {
		panic("connInfo is set in context with nil value")
	}

	return connInfo
}

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

// Package conninfo provides a ConnInfo struct that is used to handle connection-specificinfo
// and can be shared through context.
package conninfo

import (
	"context"
	"math"
	"math/rand"
	"sync"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// contextKey is a special type to represent context.WithValue keys a bit more safely.
type contextKey struct{}

// connInfoKey stores the key for withConnInfo context value.
var connInfoKey = contextKey{}

// ConnInfo represents connection info.
type ConnInfo struct {
	PeerAddr string

	rw       sync.RWMutex
	username string
	password string

	curRW   sync.RWMutex
	cursors map[int64]Cursor
}

// NewConnInfo return a new ConnInfo.
func NewConnInfo() *ConnInfo {
	return &ConnInfo{
		cursors: map[int64]Cursor{},
	}
}

// Cursor represents a cursor.
type Cursor struct {
	tx     pgx.Tx
	Iter   iterator.Interface[uint32, *types.Document]
	Filter *types.Document
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

// Cursor returns the cursor value stored.
func (connInfo *ConnInfo) Cursor(id int64) *Cursor {
	connInfo.curRW.RLock()
	defer connInfo.curRW.RUnlock()

	cursor, ok := connInfo.cursors[id]
	if !ok {
		return nil
	}

	return &cursor
}

// SetCursor stores the cursor value.
// We use "db.collection" as the key to store the cursor.
func (connInfo *ConnInfo) SetCursor(tx pgx.Tx, iter iterator.Interface[uint32, *types.Document], filter *types.Document) int64 {
	connInfo.curRW.Lock()
	defer connInfo.curRW.Unlock()

	id := connInfo.generateCursorID()

	connInfo.cursors[id] = Cursor{
		Iter:   iter,
		tx:     tx,
		Filter: filter,
	}

	return id
}

// DeleteCursor deletes the cursor from ConnInfo.
func (connInfo *ConnInfo) DeleteCursor(id int64) (err error) {
	connInfo.curRW.Lock()
	defer connInfo.curRW.Unlock()

	cursor := connInfo.cursors[id]

	cursor.Iter.Close()

	var committed bool

	defer func() {
		if committed {
			return
		}

		if rerr := cursor.tx.Rollback(context.Background()); rerr != nil {
			cursor.tx.Conn().Config().Logger.Log(
				context.Background(), pgx.LogLevelError, "failed to perform rollback",
				map[string]any{"error": rerr},
			)

			if err == nil {
				err = rerr
			}
		}
	}()

	err = cursor.tx.Commit(context.Background())
	if err != nil {
		return err
	}

	committed = true

	delete(connInfo.cursors, id)

	return nil
}

func (connInfo *ConnInfo) generateCursorID() int64 {
	var id int64

	for {
		id = rand.Int63()
		if _, ok := connInfo.cursors[id]; !ok {
			break
		}

		if id < 0 {
			id = int64(math.Abs(float64(id)))
		}

		_, ok := connInfo.cursors[id]
		if !ok {
			break
		}
	}

	return id
}

// Close frees all opened cursors.
func (connInfo *ConnInfo) Close() {
	connInfo.curRW.Lock()
	defer connInfo.curRW.Unlock()

	for k := range connInfo.cursors {
		connInfo.DeleteCursor(k)
	}
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

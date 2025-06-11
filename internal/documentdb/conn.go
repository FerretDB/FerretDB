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

package documentdb

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// Conn represents a pooled PostgreSQL connection.
// It wraps [*pgxpool.Conn] with resource tracking.
type Conn struct {
	conn  *pgxpool.Conn
	token *resource.Token
}

// newConn returns [*Conn] for the given [*pgxpool.Conn].
func newConn(conn *pgxpool.Conn) *Conn {
	res := &Conn{
		conn:  conn,
		token: resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Release returns connection back to the pool, unless it was persisted/hijacked.
// It is safe to call this method multiple times.
func (conn *Conn) Release() {
	if conn.conn != nil {
		conn.conn.Release()
		conn.conn = nil
	}

	resource.Untrack(conn, conn.token)
}

// Conn returns the underlying [*pgx.Conn]. It should not be retained by the caller.
func (conn *Conn) Conn() *pgx.Conn {
	must.NotBeZero(conn.conn)

	return conn.conn.Conn()
}

// hijack removes the connection from the pool and returns it.
// [*Conn] should still be [Release]d, and returned connection should be closed.
//
// All code that use persisted/hijacked connections should be in that package.
// The returned connection should be wrapped in a cursors for resource tracking.
func (conn *Conn) hijack() *pgx.Conn {
	must.NotBeZero(conn.conn)

	res := conn.conn.Hijack()
	conn.conn = nil

	return res
}

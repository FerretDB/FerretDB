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

// Package cursor provides access to DocumentDB cursors.
package cursor

import (
	"context"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// cursor stores DocumentDB's cursor state.
type cursor struct {
	// the order of fields is weird to make the struct smaller due to alignment

	created      time.Time
	token        *resource.Token
	conn         *pgx.Conn // only if persisted/hijacked
	continuation wirebson.RawDocument
	persisted    *prometheus.CounterVec
}

// newCursor creates a new cursor for the given continuation and connection (if any).
// It takes counter to track the total number of closed connections.
func newCursor(continuation wirebson.RawDocument, conn *pgx.Conn, persisted *prometheus.CounterVec) *cursor {
	must.BeTrue(len(continuation) > 0)

	res := &cursor{
		continuation: continuation,
		conn:         conn,
		token:        resource.NewToken(),
		created:      time.Now(),
		persisted:    persisted,
	}

	resource.Track(res, res.token)

	return res
}

// close closes the underlying connection, if any.
//
// It attempts a clean close by sending the exit message to PostgreSQL.
// However, this could block so ctx is available to limit the time to wait (up to 3 seconds).
// The underlying net.Conn.close() will always be called regardless of any other errors.
//
// It is safe to call this method multiple times, but not concurrently.
func (c *cursor) close(ctx context.Context) {
	if c.conn != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		_ = c.conn.Close(ctx)
		c.conn = nil
		c.persisted.WithLabelValues().Inc()
	}

	resource.Untrack(c, c.token)
}

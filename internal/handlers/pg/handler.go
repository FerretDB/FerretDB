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

// Package pg provides PostgreSQL handler.
package pg

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Handler data struct.
type Handler struct {
	// TODO replace those fields with
	// opts *NewOpts
	pgPool    *pgdb.Pool
	l         *zap.Logger
	peerAddr  string
	startTime time.Time
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PgPool    *pgdb.Pool
	L         *zap.Logger
	PeerAddr  string
	StartTime time.Time
}

// New returns a new handler.
func New(opts *NewOpts) common.Handler {
	return &Handler{
		pgPool:    opts.PgPool,
		l:         opts.L,
		peerAddr:  opts.PeerAddr,
		startTime: opts.StartTime,
	}
}

// MsgDebugError used for debugging purposes.
func (h *Handler) MsgDebugError(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, errors.New("debug_error")
}

// MsgDebugPanic used for debugging purposes.
func (h *Handler) MsgDebugPanic(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("debug_panic")
}

// Close prepares handler for graceful shutdown: closes connections, channels etc.
func (h *Handler) Close() {
	h.pgPool.Close()
}

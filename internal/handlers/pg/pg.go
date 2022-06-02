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
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// Handler implements handlers.Interface on top of PostgreSQL.
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
func New(opts *NewOpts) handlers.Interface {
	return &Handler{
		pgPool:    opts.PgPool,
		l:         opts.L,
		peerAddr:  opts.PeerAddr,
		startTime: opts.StartTime,
	}
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.pgPool.Close()
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

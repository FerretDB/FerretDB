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
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return common.NewErrorMsg(common.ErrNotImplemented, "I'm a stub, not a real handler for "+command)
}

// Handler implements handlers.Interface on top of PostgreSQL.
type Handler struct {
	// TODO replace those fields with embedded *NewOpts to sync with Tigris handler
	pgPool    *pgdb.Pool
	l         *zap.Logger
	startTime time.Time
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PgPool *pgdb.Pool
	L      *zap.Logger
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	h := &Handler{
		pgPool:    opts.PgPool,
		l:         opts.L,
		startTime: time.Now(),
	}
	return h, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.pgPool.Close()
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

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
	"sync"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Handler implements handlers.Interface on top of PostgreSQL.
type Handler struct {
	*NewOpts

	rw    sync.RWMutex
	pools map[string]*pgdb.Pool
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PostgreSQLURL string
	L             *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	h := &Handler{
		NewOpts: opts,
		pools:   make(map[string]*pgdb.Pool, 1),
	}
	return h, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.rw.Lock()
	defer h.rw.Unlock()

	for k, pgPool := range h.pools {
		pgPool.Close()
		delete(h.pools, k)
	}
}

// DBPool returns database connection pool for the given client connection.
func (h *Handler) DBPool(ctx context.Context) (*pgdb.Pool, error) {
	// TODO make real implementation; the current one is a stub

	h.rw.RLock()
	p, ok := h.pools[h.PostgreSQLURL]
	h.rw.RUnlock()

	if ok {
		return p, nil
	}

	h.rw.Lock()
	defer h.rw.Unlock()

	p, err := pgdb.NewPool(ctx, h.PostgreSQLURL, h.L, true, h.StateProvider)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.pools[h.PostgreSQLURL] = p

	return p, nil
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

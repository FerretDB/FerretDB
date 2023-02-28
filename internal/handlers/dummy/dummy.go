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

// Package dummy provides a basic handler implementation.
//
// The whole package can be copied to start a new handler implementation.
package dummy

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/saphana/hanadb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// / Handler implements handlers.Interface on top of PostgreSQL.
type Handler struct {
	*NewOpts

	// accessed by DBPool(ctx)
	rw   sync.RWMutex
	pool *hanadb.Pool
}

// NewOpts represents handler configuration.
type NewOpts struct {
	HANAInstanceURL string
	L               *zap.Logger
	Metrics         *connmetrics.ConnMetrics
	StateProvider   *state.Provider
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	if opts.HANAInstanceURL == "" {
		return nil, lazyerrors.New("HANA instance URL is not provided")
	}

	h := &Handler{
		NewOpts: opts,
	}

	return h, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.pool.Close()
}

// DBPool returns database connection pool for the given client connection.
//
// Pool is not closed when ctx is canceled.
func (h *Handler) DBPool(ctx context.Context) (*hanadb.Pool, error) {
	url := h.HANAInstanceURL

	h.rw.RLock()
	pool := h.pool
	h.rw.RUnlock()

	if pool != nil {
		return pool, nil
	}

	h.rw.Lock()
	defer h.rw.Unlock()

	pool, err := hanadb.NewPool(ctx, url, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.pool = pool

	return pool, nil
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

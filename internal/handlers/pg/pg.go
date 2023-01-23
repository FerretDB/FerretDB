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
	"net/url"
	"sync"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Handler implements handlers.Interface on top of PostgreSQL.
type Handler struct {
	*NewOpts
	url url.URL

	// accessed by DBPool(ctx)
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
	if opts.PostgreSQLURL == "" {
		return nil, lazyerrors.New("PostgreSQL URL is not provided")
	}

	u, err := url.Parse(opts.PostgreSQLURL)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h := &Handler{
		NewOpts: opts,
		url:     *u,
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
//
// Pool is not closed when ctx is canceled.
func (h *Handler) DBPool(ctx context.Context) (*pgdb.Pool, error) {
	connInfo := conninfo.Get(ctx)
	username, password := connInfo.Auth()

	// do not log password or full URL

	// replace authentication info only if it is present in the connection
	u := h.url
	if username != "" && password != "" {
		u.User = url.UserPassword(username, password)
	}

	// pgx 'defaultMaxConns' is 4, which is not enough for us.
	// Set it to 20 by default if no query parameter is defined.
	// See: https://github.com/FerretDB/FerretDB/issues/1844
	values := u.Query()
	if !values.Has("pool_max_conns") {
		values.Set("pool_max_conns", "20")
	}

	u.RawQuery = values.Encode()

	url := u.String()

	h.rw.RLock()
	p, ok := h.pools[url]
	h.rw.RUnlock()

	if ok {
		h.L.Debug("DBPool: found existing pool", zap.String("username", username))
		return p, nil
	}

	h.rw.Lock()
	defer h.rw.Unlock()

	p, err := pgdb.NewPool(ctx, url, h.L, h.StateProvider)
	if err != nil {
		h.L.Warn("DBPool: authentication failed", zap.String("username", username), zap.Error(err))
		return nil, lazyerrors.Error(err)
	}

	h.L.Info("DBPool: authentication succeed", zap.String("username", username))
	h.pools[url] = p

	return p, nil
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

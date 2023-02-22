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

// Package tigris provides Tigris handler.
package tigris

import (
	"context"
	"sync"
	"time"

	"github.com/tigrisdata/tigris-client-go/config"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Handler implements handlers.Interface on top of Tigris.
//
//nolint:vet // for readability
type Handler struct {
	*NewOpts

	// accessed by DBPool(ctx)
	rw    sync.RWMutex
	pools map[AuthParams]*tigrisdb.TigrisDB
}

// NewOpts represents handler configuration.
type NewOpts struct {
	AuthParams

	L               *zap.Logger
	Metrics         *connmetrics.ConnMetrics
	StateProvider   *state.Provider
	DisablePushdown bool
}

// AuthParams represents authentication parameters.
type AuthParams struct {
	URL          string
	ClientID     string
	ClientSecret string
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	if opts.URL == "" {
		return nil, lazyerrors.New("Tigris URL is not provided")
	}

	return &Handler{
		NewOpts: opts,
		pools:   make(map[AuthParams]*tigrisdb.TigrisDB, 1),
	}, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.rw.Lock()
	defer h.rw.Unlock()

	for k, p := range h.pools {
		p.Driver.Close()
		delete(h.pools, k)
	}
}

// DBPool returns database connection pool for the given client connection.
//
// Pool is not closed when ctx is canceled.
func (h *Handler) DBPool(ctx context.Context) (*tigrisdb.TigrisDB, error) {
	connInfo := conninfo.Get(ctx)
	username, password := connInfo.Auth()

	// do not log client secret

	// replace authentication info only if it is present in the connection
	ap := h.AuthParams
	if username != "" && password != "" {
		ap.ClientID = username
		ap.ClientSecret = password
	}

	// fast path

	h.rw.RLock()
	p := h.pools[ap]
	h.rw.RUnlock()

	if p != nil {
		h.L.Debug("DBPool: found existing pool", zap.String("username", username))
		return p, nil
	}

	// slow path

	h.rw.Lock()
	defer h.rw.Unlock()

	// a concurrent connection might have created a pool already; check again
	if p = h.pools[ap]; p != nil {
		h.L.Debug("DBPool: found existing pool (after acquiring lock)", zap.String("username", username))
		return p, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cfg := &config.Driver{
		URL:          h.URL,
		ClientID:     h.ClientID,
		ClientSecret: h.ClientSecret,
	}
	p, err := tigrisdb.New(ctx, cfg, h.L)
	if err != nil {
		h.L.Warn("DBPool: authentication failed", zap.String("username", username), zap.Error(err))
		return nil, lazyerrors.Error(err)
	}

	h.L.Info("DBPool: authentication succeed", zap.String("username", username))
	h.pools[ap] = p

	return p, nil
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

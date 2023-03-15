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

// Package hana provides SAP HANA handler.
package hana

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/hana/hanadb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return commonerrors.NewCommandErrorMsg(commonerrors.ErrNotImplemented, "I'm a stub, not a real handler for "+command)
}

// Handler implements handlers.Interface on top of SAP HANA.
type Handler struct {
	pools map[string]*hanadb.Pool
	*NewOpts

	// accessed by DBPool(ctx)
	rw sync.RWMutex
}

// NewOpts represents handler configuration.
type NewOpts struct {
	L             *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider
	HANAURL       string
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	if opts.HANAURL == "" {
		return nil, lazyerrors.New("HANA instance URL is not provided")
	}

	h := &Handler{
		NewOpts: opts,
		pools:   make(map[string]*hanadb.Pool, 1),
	}

	return h, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.rw.Lock()
	defer h.rw.Unlock()

	for k, p := range h.pools {
		p.Close()
		delete(h.pools, k)
	}
}

// DBPool returns database connection pool for the given client connection.
//
// Pool is not closed when ctx is canceled.
func (h *Handler) DBPool(ctx context.Context) (*hanadb.Pool, error) {
	// TODO make real implementation; the current one is a stub.
	// Used for the basic setup for connecting to HANA
	url := h.HANAURL

	h.rw.RLock()
	p, ok := h.pools[url]
	h.rw.RUnlock()

	if ok {
		return p, nil
	}

	h.rw.Lock()
	defer h.rw.Unlock()

	// a concurrent connection might have created a pool already; check again
	if p = h.pools[url]; p != nil {
		return p, nil
	}

	p, err := hanadb.NewPool(ctx, url, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.pools[url] = p

	return p, nil
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

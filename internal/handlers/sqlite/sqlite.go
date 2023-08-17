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

// Package sqlite provides SQLite handler.
//
// It is being converted into universal handler for all backends.
package sqlite

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return commonerrors.NewCommandErrorMsg(
		commonerrors.ErrNotImplemented,
		"I'm a stub, not a real handler for "+command,
	)
}

// Handler implements handlers.Interface.
type Handler struct {
	*NewOpts

	b backends.Backend

	cursors *cursor.Registry
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	URI string

	L             *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// test options
	DisableFilterPushdown bool
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	b, err := sqlite.NewBackend(&sqlite.NewBackendParams{
		URI: opts.URI,
		L:   opts.L,
	})
	if err != nil {
		return nil, err
	}

	return &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),
	}, nil
}

// Close implements handlers.Interface.
func (h *Handler) Close() {
	h.cursors.Close()
	h.b.Close()
}

// Describe implements handlers.Interface.
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.b.Describe(ch)
	h.cursors.Describe(ch)
}

// Collect implements handlers.Interface.
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.b.Collect(ch)
	h.cursors.Collect(ch)
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

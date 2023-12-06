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

// Package handler provides a universal handler implementation for all backends.
package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/decorators/oplog"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Handler provides a set of methods to process clients' requests sent over wire protocol.
//
// MsgXXX methods handle OP_MSG commands.
// CmdQuery handles a limited subset of OP_QUERY messages.
//
// Handler instance is shared between all client connections.
type Handler struct {
	*NewOpts

	b backends.Backend

	cursors  *cursor.Registry
	commands map[string]command
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Backend backends.Backend

	L             *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// test options
	DisablePushdown bool
	EnableOplog     bool
	EnableNewAuth   bool
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	b := opts.Backend

	if opts.EnableOplog {
		b = oplog.NewBackend(b, opts.L.Named("oplog"))
	}

	h := &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),
	}

	h.initCommands()

	return h, nil
}

// Close gracefully shutdowns handler.
// It should be called after listener closes all client connections and stops listening.
func (h *Handler) Close() {
	h.cursors.Close()
}

// Describe implements prometheus.Collector interface.
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.b.Describe(ch)
	h.cursors.Describe(ch)
}

// Collect implements prometheus.Collector interface.
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.b.Collect(ch)
	h.cursors.Collect(ch)
}

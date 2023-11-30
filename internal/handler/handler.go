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
	"fmt"

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
// Handler is shared between all connections! Be careful when you need connection-specific information.
// Currently, we pass connection information through context, see `ConnInfo` and its usage.
type Handler struct {
	*NewOpts

	b backends.Backend

	cursors *cursor.Registry
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
	DisableFilterPushdown bool
	EnableOplog           bool
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	b := opts.Backend

	if opts.EnableOplog {
		b = oplog.NewBackend(b, opts.L.Named("oplog"))
	}

	return &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),
	}, nil
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

// newHandlerFunc represents a function that constructs a new handler.
type newHandlerFunc func(opts *NewHandlerOpts) (*Handler, CloseBackendFunc, error)

// CloseBackendFunc represents a function that closes a backend.
type CloseBackendFunc func()

// registry maps handler names to constructors.
//
// Map values must be added through the `init()` functions in separate files
// so that we can control which handlers will be included in the build with build tags.
var registry = map[string]newHandlerFunc{}

// NewHandlerOpts represents configuration for constructing handlers.
type NewHandlerOpts struct {
	// for all backends
	Logger        *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// for `postgresql` handler
	PostgreSQLURL string

	// for `sqlite` handler
	SQLiteURL string

	// for `hana` handler
	HANAURL string

	// for `mysql` handler
	MySQLURL string

	TestOpts
}

// TestOpts represents experimental configuration options.
type TestOpts struct {
	DisableFilterPushdown bool
	EnableOplog           bool
}

// NewHandler constructs a new handler.
//
// The caller is responsible to call CloseBackendFunc when the handler is no longer needed.
func NewHandler(name string, opts *NewHandlerOpts) (*handler.Handler, CloseBackendFunc, error) {
	if opts == nil {
		return nil, nil, fmt.Errorf("opts is nil")
	}

	// handle deprecated variant
	if name == "pg" {
		name = "postgresql"
	}

	newHandler := registry[name]
	if newHandler == nil {
		return nil, nil, fmt.Errorf("unknown handler %q", name)
	}

	return newHandler(opts)
}

// Handlers returns a list of all handlers registered at compile-time.
func Handlers() []string {
	res := make([]string, 0, len(registry))

	// double check registered names and return them in the right order
	for _, h := range []string{"postgresql", "sqlite", "hana", "mysql"} {
		if _, ok := registry[h]; !ok {
			continue
		}

		res = append(res, h)
	}

	if len(res) != len(registry) {
		panic("registry is not in sync")
	}

	return res
}

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

// Package wiring provides proper setup of FerretDB components.
package wiring

import (
	"context"
	"log/slog"
	"time"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// WireOpts represents options for creating and wiring together FerretDB components.
//
//nolint:vet // for readability
type WireOpts struct {
	Logger *slog.Logger

	StateProvider *state.Provider

	PostgreSQLURL string

	Auth                   bool
	ReplSetName            string
	SessionCleanupInterval time.Duration

	TCPAddr          string
	UnixAddr         string
	TLSAddr          string
	TLSCertFile      string
	TLSKeyFile       string
	TLSCAFile        string
	Mode             middleware.Mode
	ProxyAddr        string
	ProxyTLSCertFile string
	ProxyTLSKeyFile  string
	ProxyTLSCAFile   string
	RecordsDir       string
}

// WireResult represents [Wire] result.
type WireResult struct {
	Pool        *documentdb.Pool
	Listener    *clientconn.Listener
	ConnMetrics *connmetrics.ConnMetrics
}

// Wire creates and wires together:
//   - PostgreSQL/DocumentDB connection pool ([*dociumentdb.Pool]);
//   - DocumentDB handler ([*handler.Handler]);
//   - wire protocol listener ([*clientconn.Listener]).
//
// It does not change the global state or creates components that are different in tests.
// For example, it does not:
//   - change global logger (it is different in tests).
//
// It returns nil if any of the components could not be created.
// The error is already logged, so the caller may just exit.
func Wire(ctx context.Context, opts *WireOpts) *WireResult {
	must.NotBeZero(opts)

	p, err := documentdb.NewPool(opts.PostgreSQLURL, logging.WithName(opts.Logger, "pool"), opts.StateProvider)
	if err != nil {
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct connection pool", logging.Error(err))
		return nil
	}

	lm := connmetrics.NewListenerMetrics()

	//exhaustruct:enforce
	h, err := handler.New(&handler.NewOpts{
		Pool: p, // handler takes over the pool

		Auth:        opts.Auth,
		TCPHost:     opts.TCPAddr,
		ReplSetName: opts.ReplSetName,

		L:             logging.WithName(opts.Logger, "handler"),
		ConnMetrics:   lm.ConnMetrics,
		StateProvider: opts.StateProvider,

		SessionCleanupInterval: opts.SessionCleanupInterval,
	})
	if err != nil {
		p.Close()
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct handler", logging.Error(err))

		return nil
	}

	//exhaustruct:enforce
	lis, err := clientconn.Listen(&clientconn.ListenerOpts{
		Handler: h, // listener takes over the handler
		Metrics: lm,
		Logger:  opts.Logger,

		TCP:  opts.TCPAddr,
		Unix: opts.UnixAddr,

		TLS:         opts.TLSAddr,
		TLSCertFile: opts.TLSCertFile,
		TLSKeyFile:  opts.TLSKeyFile,
		TLSCAFile:   opts.TLSCAFile,

		Mode:             opts.Mode,
		ProxyAddr:        opts.ProxyAddr,
		ProxyTLSCertFile: opts.ProxyTLSCertFile,
		ProxyTLSKeyFile:  opts.ProxyTLSKeyFile,
		ProxyTLSCAFile:   opts.ProxyTLSCAFile,

		TestRecordsDir: opts.RecordsDir,
	})
	if err != nil {
		p.Close()
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct listener", logging.Error(err))

		return nil
	}

	//exhaustruct:enforce
	return &WireResult{
		Pool:        p,
		Listener:    lis,
		ConnMetrics: lm.ConnMetrics,
	}
}

// Run runs all components until ctx is canceled.
//
// When this method returns, all components are stopped.
func (wr *WireResult) Run(ctx context.Context) {
	wr.Listener.Run(ctx)
}

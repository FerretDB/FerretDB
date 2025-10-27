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

// Package setup provides proper setup of FerretDB components.
package setup

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/handlers/proxy"
	"github.com/FerretDB/FerretDB/v2/internal/mcp"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// SetupOpts represents options for creating and setting up FerretDB components.
//
//nolint:vet // for readability
type SetupOpts struct {
	Logger        *slog.Logger
	StateProvider *state.Provider
	Metrics       *middleware.Metrics

	// DocumentDB handler
	PostgreSQLURL          string
	Auth                   bool
	ReplSetName            string
	SessionCleanupInterval time.Duration

	// Proxy handler
	ProxyAddr        string
	ProxyTLSCertFile string
	ProxyTLSKeyFile  string
	ProxyTLSCAFile   string

	// Wire protocol listener
	TCPAddr        string // empty value disables TCP listener
	UnixAddr       string // empty value disables Unix listener
	TLSAddr        string // empty value disables TLS listener
	TLSCertFile    string
	TLSKeyFile     string
	TLSCAFile      string
	Mode           middleware.Mode
	TestRecordsDir string // empty value disables recording

	// DataAPI listener
	DataAPIAddr string // empty value disables Data API listener

	// MCPAddr listener
	MCPAddr string // empty value disables MCP listener
}

// SetupResult represents [Setup] result.
type SetupResult struct {
	docdbH          middleware.Handler
	proxyH          middleware.Handler
	m               *middleware.Middleware
	WireListener    *clientconn.Listener
	DataAPIListener *dataapi.Listener
	MCPListener     *mcp.Listener
}

// Setup creates and sets up:
//   - DocumentDB handler ([*handler.Handler]);
//   - proxy handler ([*proxy.Handler]);
//   - middleware ([*middleware.Middleware]);
//   - wire protocol listener ([*clientconn.Listener]);
//   - Data API listener ([*dataapi.Listener]);
//   - MCP listener ([*mcp.Listener]);
//   - unregistered Prometheus collector for the above components.
//
// It does not change the global state or creates components that are different in tests.
// For example, it does not:
//   - change global logger (it is different in tests);
//   - set up state provider (it is different in tests);
//   - set up telemetry reporter (it is not needed in tests);
//   - set up debug handler (it is global and uses the global Prometheus gatherer);
//   - set up OpenTelemetry trace exporter (it is global).
//
// It returns nil if any of the components could not be created.
// The error is already logged, so the caller may just exit.
func Setup(ctx context.Context, opts *SetupOpts) *SetupResult {
	must.NotBeZero(opts)

	var res SetupResult
	var err error

	// If we exit early, we must Run what we already created to avoid leaks:
	// components create resources like listening sockets, database pools, etc in constructors.
	exitCtx, exitCancel := context.WithCancel(ctx)
	exitCancel() // no defer - we need canceled context

	//exhaustruct:enforce
	res.docdbH, err = handler.New(&handler.NewOpts{
		PostgreSQLURL: opts.PostgreSQLURL,
		Auth:          opts.Auth,

		// That might require a started listener.
		// TODO https://github.com/FerretDB/FerretDB/issues/4965
		TCPHost: opts.TCPAddr,

		ReplSetName: opts.ReplSetName,

		L:             logging.WithName(opts.Logger, "documentdb"),
		Metrics:       opts.Metrics,
		StateProvider: opts.StateProvider,

		SessionCleanupInterval: opts.SessionCleanupInterval,
	})
	if err != nil {
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct DocumentDB handler", logging.Error(err))
		res.Run(exitCtx)

		return nil
	}

	if opts.ProxyAddr != "" {
		//exhaustruct:enforce
		res.proxyH, err = proxy.New(&proxy.NewOpts{
			Addr:        opts.ProxyAddr,
			TLSCertFile: opts.ProxyTLSCertFile,
			TLSKeyFile:  opts.ProxyTLSKeyFile,
			TLSCAFile:   opts.ProxyTLSCAFile,
			L:           logging.WithName(opts.Logger, "proxy"),
		})
		if err != nil {
			opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct proxy handler", logging.Error(err))
			res.Run(exitCtx)

			return nil
		}
	}

	//exhaustruct:enforce
	res.m = middleware.New(&middleware.NewOpts{
		Mode:    opts.Mode,
		DocDB:   res.docdbH,
		Proxy:   res.proxyH,
		Metrics: opts.Metrics,
		L:       logging.WithName(opts.Logger, "middleware"),
	})

	//exhaustruct:enforce
	res.WireListener, err = clientconn.Listen(&clientconn.ListenerOpts{
		M:      res.m,
		Logger: opts.Logger,

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

		TestRecordsDir: opts.TestRecordsDir,
	})
	if err != nil {
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct wire protocol listener", logging.Error(err))
		res.Run(exitCtx)

		return nil
	}

	if opts.DataAPIAddr != "" {
		res.DataAPIListener, err = dataapi.Listen(&dataapi.ListenOpts{
			L:       logging.WithName(opts.Logger, "dataapi"),
			M:       res.m,
			TCPAddr: opts.DataAPIAddr,
			Auth:    opts.Auth,
		})
		if err != nil {
			opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct DataAPI listener", logging.Error(err))
			res.Run(exitCtx)

			return nil
		}
	}

	if opts.MCPAddr != "" {
		res.MCPListener, err = mcp.Listen(&mcp.ListenOpts{
			L:       logging.WithName(opts.Logger, "mcp"),
			M:       res.m,
			TCPAddr: opts.MCPAddr,
			Auth:    opts.Auth,
		})
		if err != nil {
			opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct MCP listener", logging.Error(err))
			res.Run(exitCtx)

			return nil
		}
	}

	return &res
}

// Run runs all components until ctx is canceled.
//
// When this method returns, all components are stopped.
func (sr *SetupResult) Run(ctx context.Context) {
	// Handlers should be started first to prevent commands from clients
	// reach handlers before they are Run.
	//
	// Listeners should be stopped first to allow graceful disconnect.
	// Listeners' Run methods already implement graceful shutdown;
	// we just need to wait for them to stop before stopping other components.

	hDone := make(chan struct{})

	hCtx, hCancel := context.WithCancel(context.WithoutCancel(ctx)) // inherit values
	defer hCancel()

	go func() {
		defer close(hDone)
		sr.runHandlers(hCtx)
	}()

	lDone := make(chan struct{})

	lCtx, lCancel := context.WithCancel(ctx)
	defer lCancel()

	go func() {
		defer close(lDone)
		sr.runListeners(lCtx)
	}()

	<-lDone
	hCancel()
	<-hDone
}

// runListeners runs all listeners until ctx is canceled.
func (sr *SetupResult) runListeners(ctx context.Context) {
	var wg sync.WaitGroup

	if sr.WireListener != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.WireListener.Run(ctx)
		}()
	}

	if sr.DataAPIListener != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.DataAPIListener.Run(ctx)
		}()
	}

	if sr.MCPListener != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.MCPListener.Run(ctx)
		}()
	}

	wg.Wait()
}

// runHandlers runs all handlers until ctx is canceled.
func (sr *SetupResult) runHandlers(ctx context.Context) {
	var wg sync.WaitGroup

	{
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.m.Run(ctx)
		}()
	}

	if sr.docdbH != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.docdbH.Run(ctx)
		}()
	}

	if sr.proxyH != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.proxyH.Run(ctx)
		}()
	}

	wg.Wait()
}

// Describe implements [prometheus.Collector].
func (sr *SetupResult) Describe(ch chan<- *prometheus.Desc) {
	if sr.docdbH != nil {
		sr.docdbH.Describe(ch)
	}

	if sr.proxyH != nil {
		sr.proxyH.Describe(ch)
	}

	sr.m.Describe(ch)

	if sr.WireListener != nil {
		sr.WireListener.Describe(ch)
	}

	if sr.DataAPIListener != nil {
		sr.DataAPIListener.Describe(ch)
	}

	if sr.MCPListener != nil {
		sr.MCPListener.Describe(ch)
	}
}

// Collect implements [prometheus.Collector].
func (sr *SetupResult) Collect(ch chan<- prometheus.Metric) {
	if sr.docdbH != nil {
		sr.docdbH.Collect(ch)
	}

	if sr.proxyH != nil {
		sr.proxyH.Collect(ch)
	}

	sr.m.Collect(ch)

	if sr.WireListener != nil {
		sr.WireListener.Collect(ch)
	}

	if sr.DataAPIListener != nil {
		sr.DataAPIListener.Collect(ch)
	}

	if sr.MCPListener != nil {
		sr.MCPListener.Collect(ch)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*SetupResult)(nil)
)

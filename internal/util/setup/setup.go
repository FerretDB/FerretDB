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

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// SetupOpts represents options for creating and setting up FerretDB components.
//
//nolint:vet // for readability
type SetupOpts struct {
	Logger *slog.Logger

	StateProvider   *state.Provider
	ListenerMetrics *connmetrics.ListenerMetrics

	PostgreSQLURL string

	// DocumentDB handler
	Auth                   bool
	ReplSetName            string
	SessionCleanupInterval time.Duration

	// Wire protocol listener
	TCPAddr          string // empty value disables TCP listener
	UnixAddr         string // empty value disables Unix listener
	TLSAddr          string // empty value disables TLS listener
	TLSCertFile      string
	TLSKeyFile       string
	TLSCAFile        string
	Mode             middleware.Mode
	ProxyAddr        string
	ProxyTLSCertFile string
	ProxyTLSKeyFile  string
	ProxyTLSCAFile   string
	RecordsDir       string

	// DataAPI listener
	DataAPIAddr string // empty value disables Data API listener
}

// SetupResult represents [Setup] result.
type SetupResult struct {
	WireListener    *clientconn.Listener
	DataAPIListener *dataapi.Listener
}

// Setup creates and sets up:
//   - PostgreSQL/DocumentDB connection pool ([*documentdb.Pool]);
//   - DocumentDB handler ([*handler.Handler]);
//   - wire protocol listener ([*clientconn.Listener]);
//   - Data API listener ([*dataapi.Listener]).
//
// It does not change the global state or creates components that are different in tests.
// For example, it does not:
//   - change global logger (it is different in tests);
//   - set up telemetry reporter (it is not needed in tests).
//
// It returns nil if any of the components could not be created.
// The error is already logged, so the caller may just exit.
func Setup(ctx context.Context, opts *SetupOpts) *SetupResult {
	must.NotBeZero(opts)

	p, err := documentdb.NewPool(opts.PostgreSQLURL, logging.WithName(opts.Logger, "pool"), opts.StateProvider)
	if err != nil {
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct connection pool", logging.Error(err))
		return nil
	}

	//exhaustruct:enforce
	h, err := handler.New(&handler.NewOpts{
		Pool: p, // handler takes over the pool

		Auth:        opts.Auth,
		TCPHost:     opts.TCPAddr,
		ReplSetName: opts.ReplSetName,

		L:             logging.WithName(opts.Logger, "handler"),
		ConnMetrics:   opts.ListenerMetrics.ConnMetrics,
		StateProvider: opts.StateProvider,

		SessionCleanupInterval: opts.SessionCleanupInterval,
	})
	if err != nil {
		p.Close()
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct handler", logging.Error(err))

		return nil
	}

	//exhaustruct:enforce
	wireLis, err := clientconn.Listen(&clientconn.ListenerOpts{
		Handler: h, // listener takes over the handler
		Metrics: opts.ListenerMetrics,
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
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct wire protocol listener", logging.Error(err))

		return nil
	}

	var dataapiLis *dataapi.Listener
	if opts.DataAPIAddr != "" {
		dataapiLis, err = dataapi.Listen(&dataapi.ListenOpts{
			L:       logging.WithName(opts.Logger, "dataapi"),
			Handler: h, // does not takes over
			TCPAddr: opts.DataAPIAddr,
			Auth:    opts.Auth,
		})
		if err != nil {
			p.Close()
			opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct DataAPI listener", logging.Error(err))

			return nil
		}
	}

	//exhaustruct:enforce
	return &SetupResult{
		WireListener:    wireLis,
		DataAPIListener: dataapiLis,
	}
}

// Run runs all components until ctx is canceled.
//
// When this method returns, all components are stopped.
func (sr *SetupResult) Run(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		sr.WireListener.Run(ctx)
	}()

	if sr.DataAPIListener != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sr.DataAPIListener.Run(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()
}

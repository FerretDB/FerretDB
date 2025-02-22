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

// Package ferretdb provides embeddable FerretDB implementation.
//
// See [`build/version` package documentation]
// for information about Go build tags that affect this package.
//
// [`build/version` package documentation]: https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version
package ferretdb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/telemetry"
)

// Keep structure and order of Config in sync with the main package and documentation.
// Avoid breaking changes.

// Config represents FerretDB configuration.
type Config struct {
	// PostgreSQL URL. Required.
	PostgreSQLURL string

	// Listen TCP address for MongoDB protocol.
	// If empty, TCP listener is disabled.
	ListenAddr string

	// State directory. Required.
	StateDir string
}

// FerretDB represents an instance of embedded FerretDB implementation.
type FerretDB struct {
	tl  *telemetry.Reporter
	lis *clientconn.Listener
}

// New creates a new instance of embedded FerretDB implementation.
func New(config *Config) (*FerretDB, error) {
	version.Get().Package = "embedded"

	sp, err := state.NewProviderDir(config.StateDir)
	if err == nil {
		err = sp.Update(func(s *state.State) {
			s.TelemetryLocked = true
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to set up state provider: %w", err)
	}

	// Note that the current implementation requires `*logging.Handler` in the `getLog` command implementation.
	// TODO https://github.com/FerretDB/FerretDB/issues/4750
	lOpts := &logging.NewHandlerOpts{
		Base:          "text",
		Level:         slog.LevelError + 100500, // effectively disables logging
		RemoveTime:    true,
		RemoveLevel:   true,
		RemoveSource:  true,
		CheckMessages: false,
	}
	logger := logging.Logger(io.Discard, lOpts, "")

	metrics := connmetrics.NewListenerMetrics()

	tl, err := telemetry.NewReporter(&telemetry.NewReporterOpts{
		URL:            "https://beacon.ferretdb.com/",
		Dir:            config.StateDir,
		F:              &telemetry.Flag{},
		DNT:            os.Getenv("DO_NOT_TRACK"),
		ExecName:       os.Args[0],
		P:              sp,
		ConnMetrics:    metrics.ConnMetrics,
		L:              logging.WithName(logger, "telemetry"),
		UndecidedDelay: time.Hour,
		ReportInterval: 24 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry reporter: %w", err)
	}

	p, err := documentdb.NewPool(config.PostgreSQLURL, logging.WithName(logger, "pool"), sp)
	if err != nil {
		return nil, fmt.Errorf("failed to construct pool: %w", err)
	}

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: false,

		TCPHost:     "",
		ReplSetName: "",

		L:             logging.WithName(logger, "handler"),
		ConnMetrics:   metrics.ConnMetrics,
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("failed to construct handler: %w", err)
	}

	lis, err := clientconn.Listen(&clientconn.ListenerOpts{
		Handler: h,
		Metrics: metrics,
		Logger:  logger,

		TCP: config.ListenAddr,

		Mode: clientconn.NormalMode,
	})
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("failed to construct listener: %w", err)
	}

	return &FerretDB{
		tl:  tl,
		lis: lis,
	}, nil
}

// Run runs FerretDB until ctx is canceled.
//
// When this method returns, all listeners, all client connections, and all PostgreSQL connections are closed.
//
// It is required to run this method in order to initialize listeners with their respective IP addresses and ports.
// Calling [*FerretDB.MongoDBURI] before calling this method will result in a deadlock.
func (f *FerretDB) Run(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		f.tl.Run(ctx)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		f.lis.Run(ctx)
	}()

	wg.Wait()
}

// MongoDBURI returns MongoDB URI for this FerretDB instance.
func (f *FerretDB) MongoDBURI() string {
	u := &url.URL{
		Scheme: "mongodb",
		Host:   f.lis.TCPAddr().String(),
		Path:   "/",
	}

	return u.String()
}

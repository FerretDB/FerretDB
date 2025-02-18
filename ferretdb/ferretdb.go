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

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// Keep structure and order of Config in sync with the main package and documentation.
// Avoid breaking changes.

// Config represents FerretDB configuration.
type Config struct {
	// PostgreSQL URL. Required.
	PostgreSQLURL string

	Listen struct {
		// Listen TCP address for MongoDB protocol.
		// If empty, TCP listener is disabled.
		Addr string
	}
}

// FerretDB represents an instance of embeddable FerretDB implementation.
type FerretDB struct {
	config *Config

	l *clientconn.Listener
}

// New creates a new instance of embeddable FerretDB implementation.
func New(config *Config) (*FerretDB, error) {
	version.Get().Package = "embedded"

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

	sp, err := state.NewProvider("")
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	// Telemetry reporter is not created or running anyway,
	// but disable telemetry explicitly to disable confusing startupWarnings.
	err = sp.Update(func(s *state.State) {
		s.DisableTelemetry()
		s.TelemetryLocked = true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	metrics := connmetrics.NewListenerMetrics()

	h, closeBackend, err := registry.NewHandler(config.Handler, &registry.NewHandlerOpts{
		Logger:        logger,
		ConnMetrics:   metrics.ConnMetrics,
		StateProvider: sp,
		TCPHost:       config.Listener.TCP,

		PostgreSQLURL: config.PostgreSQLURL,

		SQLiteURL: config.SQLiteURL,

		TestOpts: registry.TestOpts{
			CappedCleanupPercentage: 10,
			BatchSize:               100,
		},
	})
	if err != nil {
		if closeBackend != nil {
			closeBackend()
		}
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	l, err := clientconn.Listen(&clientconn.NewListenerOpts{
		TCP:  config.Listener.TCP,
		Unix: config.Listener.Unix,

		TLS:         config.Listener.TLS,
		TLSCertFile: config.Listener.TLSCertFile,
		TLSKeyFile:  config.Listener.TLSKeyFile,
		TLSCAFile:   config.Listener.TLSCAFile,

		Mode:    clientconn.NormalMode,
		Metrics: metrics,
		Handler: h,
		Logger:  logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	return &FerretDB{
		config:       config,
		closeBackend: closeBackend,
		l:            l,
	}, nil
}

// Run runs FerretDB until ctx is canceled.
//
// When this method returns, listener and all connections, as well as handler are closed.
//
// It is required to run this method in order to initialize the listeners with their respective
// IP address and port. Calling methods which require the listener's address (eg: [*FerretDB.MongoDBURI]
// requires it for configuring its Host URL) before calling this method might result in a deadlock.
func (f *FerretDB) Run(ctx context.Context) error {
	defer f.closeBackend()

	f.l.Run(ctx)

	return nil
}

// MongoDBURI returns MongoDB URI for this FerretDB instance.
//
// TCP's connection string is returned if both TCP and Unix listeners are enabled.
// TLS is preferred over both.
func (f *FerretDB) MongoDBURI() string {
	var u *url.URL

	switch {
	case f.config.Listener.TLS != "":
		q := url.Values{
			"tls": []string{"true"},
		}

		u = &url.URL{
			Scheme:   "mongodb",
			Host:     f.l.TLSAddr().String(),
			Path:     "/",
			RawQuery: q.Encode(),
		}
	case f.config.Listener.TCP != "":
		u = &url.URL{
			Scheme: "mongodb",
			Host:   f.l.TCPAddr().String(),
			Path:   "/",
		}
	case f.config.Listener.Unix != "":
		// MongoDB really wants Unix domain socket path in the host part of the URI
		u = &url.URL{
			Scheme: "mongodb",
			Host:   f.l.UnixAddr().String(),
			Path:   "/",
		}
	}

	return u.String()
}

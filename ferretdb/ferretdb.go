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
// See [github.com/FerretDB/FerretDB/build/version] package documentation
// for information about Go build tags that affect this package.
package ferretdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Config represents FerretDB configuration.
type Config struct {
	Listener ListenerConfig

	// Handler to use; one of `postgresql` or `sqlite`.
	Handler string

	// PostgreSQL connection string for `postgresql` handler.
	// See:
	//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ParseConfig
	//   - https://pkg.go.dev/github.com/jackc/pgx/v5#ParseConfig
	//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#ParseConfig
	PostgreSQLURL string // For example: `postgres://hostname:5432/ferretdb`.

	// SQLite URI (directory) for `sqlite` handler.
	// See https://www.sqlite.org/uri.html.
	SQLiteURL string // For example: `file:data/`.
}

// ListenerConfig represents listener configuration.
type ListenerConfig struct {
	// Listen TCP address.
	// If empty, TCP listener is disabled.
	TCP string

	// Listen Unix domain socket path.
	// If empty, Unix listener is disabled.
	Unix string

	// Listen TLS address.
	// If empty, TLS listener is disabled.
	TLS string

	// Server certificate path.
	TLSCertFile string

	// Server key path.
	TLSKeyFile string

	// Root CA certificate path.
	TLSCAFile string
}

// FerretDB represents an instance of embeddable FerretDB implementation.
type FerretDB struct {
	config *Config

	l *clientconn.Listener
}

// New creates a new instance of embeddable FerretDB implementation.
func New(config *Config) (*FerretDB, error) {
	version.Get().Package = "embedded"

	if config.Listener.TCP == "" &&
		config.Listener.Unix == "" &&
		config.Listener.TLS == "" {
		return nil, errors.New("Listener TCP, Unix and TLS are empty")
	}

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

	h, err := registry.NewHandler(config.Handler, &registry.NewHandlerOpts{
		Logger:        logger,
		ConnMetrics:   metrics.ConnMetrics,
		StateProvider: sp,

		PostgreSQLURL: config.PostgreSQLURL,

		SQLiteURL: config.SQLiteURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		TCP:         config.Listener.TCP,
		Unix:        config.Listener.Unix,
		TLS:         config.Listener.TLS,
		TLSCertFile: config.Listener.TLSCertFile,
		TLSKeyFile:  config.Listener.TLSKeyFile,
		TLSCAFile:   config.Listener.TLSCAFile,

		Mode:    clientconn.NormalMode,
		Metrics: metrics,
		Handler: h,
		Logger:  logger,
	})

	return &FerretDB{
		config: config,
		l:      l,
	}, nil
}

// Run runs FerretDB until ctx is canceled.
//
// When this method returns, listener and all connections, as well as handler are closed.
//
// It is required to run this method in order to initialise the listeners with their respective
// IP address and port. Calling methods which require the listener's address (eg: [*FerretDB.MongoDBURI]
// requires it for configuring its Host URL) before calling this method might result in a deadlock.
func (f *FerretDB) Run(ctx context.Context) error {
	err := f.l.Run(ctx)
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	if err != nil {
		// Do not expose internal error details.
		// If you need stable error values and/or types for some cases, please create an issue.
		err = errors.New(err.Error())
	}

	return err
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
		// MongoDB really wants Unix socket path in the host part of the URI
		u = &url.URL{
			Scheme: "mongodb",
			Host:   f.l.UnixAddr().String(),
			Path:   "/",
		}
	}

	return u.String()
}

// logger is a global logger used by FerretDB.
//
// If it is a problem for you, please create an issue.
var logger *zap.Logger

// Initialize the global logger there to avoid creating too many issues for zap users that initialize it in their
// `main()` functions. It is still not a full solution; eventually, we should remove the usage of the global logger.
func init() {
	l := zap.ErrorLevel
	if version.Get().DebugBuild {
		l = zap.DebugLevel
	}

	logging.Setup(l, "")
	logger = zap.L()
}

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
// See [build/version package documentation] for information about Go build tags that affect this package.
//
// [build/version package documentation]: https://pkg.go.dev/github.com/FerretDB/FerretDB/build/version
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

	// Handler to use; one of `pg` or `tigris` (if enabled at compile-time).
	Handler string

	// PostgreSQL connection string for `pg` handler.
	PostgreSQLURL string // For example: `postgres://hostname:5432/ferretdb`.

	// Tigris parameters for `tigris` handler.
	// See https://docs.tigrisdata.com/overview/authentication
	// and https://docs.tigrisdata.com/golang/getting-started.
	TigrisClientID     string
	TigrisClientSecret string
	TigrisToken        string
	TigrisURL          string
}

// ListenerConfig represents listener configuration.
type ListenerConfig struct {
	// Listen TCP address.
	// If empty, TCP listener is disabled.
	Addr string

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
	if config.Listener.Addr == "" &&
		config.Listener.Unix == "" &&
		config.Listener.TLS == "" {
		return nil, errors.New("Listener Addr, Unix and TLS are empty")
	}

	p, err := state.NewProvider("")
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	metrics := connmetrics.NewListenerMetrics()

	h, err := registry.NewHandler(config.Handler, &registry.NewHandlerOpts{
		Ctx:           context.Background(),
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: config.PostgreSQLURL,

		TigrisClientID:     config.TigrisClientID,
		TigrisClientSecret: config.TigrisClientSecret,
		TigrisToken:        config.TigrisToken,
		TigrisURL:          config.TigrisURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct handler: %s", err)
	}

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		Listener: clientconn.ListenerOpts{
			Addr:        config.Listener.Addr,
			Unix:        config.Listener.Unix,
			TLS:         config.Listener.TLS,
			TLSCertFile: config.Listener.TLSCertFile,
			TLSKeyFile:  config.Listener.TLSKeyFile,
			TLSCAFile:   config.Listener.TLSCAFile,
		},
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

// Run runs FerretDB until ctx is done.
//
// When this method returns, listener and all connections are closed.
func (f *FerretDB) Run(ctx context.Context) error {
	defer f.l.Handler.Close()

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
func (f *FerretDB) MongoDBURI() string {
	var u *url.URL

	switch {
	case f.config.Listener.TLS != "":
		q := make(url.Values)

		q.Set("tls", "true")

		u = &url.URL{
			Scheme:   "mongodb",
			Host:     f.l.TLS().String(),
			Path:     "/",
			RawQuery: q.Encode(),
		}
	case f.config.Listener.Addr != "":
		u = &url.URL{
			Scheme: "mongodb",
			Host:   f.l.Addr().String(),
			Path:   "/",
		}
	case f.config.Listener.Unix != "":
		// MongoDB really wants Unix socket path in the host part of the URI
		u = &url.URL{
			Scheme: "mongodb",
			Host:   f.l.Unix().String(),
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

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
package ferretdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Config represents FerretDB configuration.
type Config struct {
	// TCP Listen address.
	ListenAddr string

	// Listen Unix domain socket path; defaults to "".
	ListenUnix string

	// Handler to use; one of `pg` or `tigris` (if enabled at compile-time).
	Handler string

	// PostgreSQL connection string for `pg` handler.
	PostgreSQLURL string // For example: `postgres://username:password@hostname:5432/ferretdb`.

	// Tigris parameters for `tigris` handler.
	// See https://docs.tigrisdata.com/overview/authentication
	// and https://docs.tigrisdata.com/golang/getting-started.
	TigrisClientID     string
	TigrisClientSecret string
	TigrisToken        string
	TigrisURL          string
}

// FerretDB represents an instance of embeddable FerretDB implementation.
type FerretDB struct {
	config     *Config
	listenAddr string
}

// New creates a new instance of embeddable FerretDB implementation.
func New(config *Config) (*FerretDB, error) {
	listenAddr := config.ListenAddr
	if listenAddr == "" {
		listenAddr = "127.0.0.1:27017"
	}

	return &FerretDB{
		config:     config,
		listenAddr: listenAddr,
	}, nil
}

// Run runs FerretDB until ctx is done.
//
// When this method returns, listener and all connections are closed.
func (f *FerretDB) Run(ctx context.Context) error {
	p, err := state.NewProvider("")
	if err != nil {
		return fmt.Errorf("failed to construct handler: %s", err)
	}

	cmdsList := maps.Keys(common.Commands)
	sort.Strings(cmdsList)

	metrics := connmetrics.NewListenerMetrics(cmdsList)

	newOpts := registry.NewHandlerOpts{
		Ctx:           context.Background(),
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: f.config.PostgreSQLURL,

		TigrisClientID:     f.config.TigrisClientID,
		TigrisClientSecret: f.config.TigrisClientSecret,
		TigrisToken:        f.config.TigrisToken,
		TigrisURL:          f.config.TigrisURL,
	}
	h, err := registry.NewHandler(f.config.Handler, &newOpts)
	if err != nil {
		return fmt.Errorf("failed to construct handler: %s", err)
	}
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr: f.listenAddr,
		ListenUnix: f.config.ListenUnix,
		Mode:       clientconn.NormalMode,
		Metrics:    metrics,
		Handler:    h,
		Logger:     logger,
	})

	if err = l.Run(ctx); err != nil {
		// Do not expose internal error details.
		// If you need stable error values and/or types for some cases, please create an issue.
		err = errors.New(err.Error())
	}
	return err
}

// MongoDBURI returns MongoDB URI for this FerretDB instance.
// If it was spawned on domain socket then it will be returned,
// in case of listening both on TCP and a socket TCP connection string will be returned.
func (f *FerretDB) MongoDBURI() string {
	var u url.URL
	if f.isListeningOnlyOnSock() {
		u = url.URL{
			Scheme: "mongodb",
			Host:   url.PathEscape(f.config.ListenUnix),
		}
	} else {
		u = url.URL{
			Scheme: "mongodb",
			Host:   f.listenAddr,
			Path:   "/",
		}
	}

	return u.String()
}

// TODO
func (f *FerretDB) isListeningOnlyOnSock() bool {
	return f.config.ListenAddr == "" && f.config.ListenUnix != ""
}

// logger is a global logger used by FerretDB.
//
// If it is a problem for you, please create an issue.
var logger *zap.Logger

// Initialize the global logger there to avoid creating too many issues for zap users that initialize it in their
// `main()` functions. It is still not a full solution; eventually, we should remove the usage of the global logger.
func init() {
	logging.Setup(zap.FatalLevel, "")
	logger = zap.L()
}

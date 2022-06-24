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

package ferretdb

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

// Config contains a backend connection strings.
type Config struct {
	Handler     string
	PostgresURL string
	TigrisURL   string
}

// FerretDB proxy.
type FerretDB struct {
	config Config
}

// New registers backend handler and returns a new FerretDB.
func New(conf Config) FerretDB {
	switch conf.Handler {
	case "pg":
		// already initialized in init()
	case "tigris":
		registry.RegisterTigris()
	default:
		panic(fmt.Sprintf("Unknown backend handler %q.", conf.Handler))
	}
	return FerretDB{
		config: conf,
	}
}

// GetConnectionString returns the backend connection string.
func (fdb *FerretDB) GetConnectionString() string {
	return "mongodb://127.0.0.1:27017"
}

// Run runs the FerretDB proxy as a library with:
// * error level logging
// * monitoring disabled
// * handler PostgreSQL.
func (fdb *FerretDB) Run(ctx context.Context) error {
	listenAddr := "127.0.0.1:27017"
	mode := clientconn.NormalMode
	testConnTimeout := time.Duration(0)

	logging.Setup(zapcore.ErrorLevel)
	logger := zap.L()

	newHandler := registry.Handlers[fdb.config.Handler]
	if newHandler == nil {
		logger.Sugar().Fatalf("Unknown backend handler %q.", fdb.config.Handler)
	}

	opts := registry.NewHandlerOpts{
		PostgresURL: fdb.config.PostgresURL,
		TigrisURL:   fdb.config.TigrisURL,
		Ctx:         ctx,
		Logger:      logger,
	}
	h := registry.New(fdb.config.Handler, opts)
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      listenAddr,
		Mode:            clientconn.Mode(mode),
		Handler:         h,
		Logger:          logger,
		TestConnTimeout: testConnTimeout,
	})

	err := l.Run(ctx)
	if err != nil && err != context.Canceled {
		logger.Error("Listener stopped", zap.Error(err))
	}
	return nil
}

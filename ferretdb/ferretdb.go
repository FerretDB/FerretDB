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
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

// Config contains a backend connection strings.
type Config struct {
	PostgresURL string
	TigrisURL   string
}

// FerretDB proxy.
type FerretDB struct {
	config Config
}

// New returns a new FerretDB.
func New(conf Config) FerretDB {
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
	proxyAddr := "127.0.0.1:37017"
	mode := clientconn.NormalMode
	handler := "pg"
	testConnTimeout := time.Duration(0)

	_, ok := registry.Handlers["pg"]
	if !ok {
		panic("no pg handler registered")
	}

	logging.Setup(zapcore.ErrorLevel)
	logger := zap.L()

	newHandler := registry.Handlers[handler]
	if newHandler == nil {
		logger.Sugar().Fatalf("Unknown backend handler %q.", handler)
	}
	h, err := newHandler(&registry.NewHandlerOpts{
		PostgresURL: fdb.config.PostgresURL,
		TigrisURL:   fdb.config.TigrisURL,
		Ctx:         ctx,
		Logger:      logger,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      listenAddr,
		ProxyAddr:       proxyAddr,
		Mode:            clientconn.Mode(mode),
		Handler:         h,
		Logger:          logger,
		TestConnTimeout: testConnTimeout,
	})

	err = l.Run(ctx)
	if err == nil || err == context.Canceled {
		logger.Info("Listener stopped")
	} else {
		logger.Error("Listener stopped", zap.Error(err))
	}
	return nil
}

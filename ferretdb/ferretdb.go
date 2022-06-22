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
	"net"
	"net/url"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/common/register"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// Config contains a PostgreSQL backend connection string.
type Config struct {
	PostgreSQLConnectionString string
}

// FerretDB proxy.
type FerretDB struct {
	config   Config
	mongoURL string
}

// New returns a new FerretDB.
func New(conf Config) FerretDB {
	return FerretDB{
		config:   conf,
		mongoURL: transformConnStringPgToMongo(conf.PostgreSQLConnectionString),
	}
}

// Run runs the FerretDB proxy as a library with logging error level, monitoring disabled, handler PostgreSQL.
func (fdb *FerretDB) Run(ctx context.Context, conf Config) (string, error) {
	listenAddr := "127.0.0.1:27017"
	proxyAddr := "127.0.0.1:37017"
	debugAddr := "127.0.0.1:8088"
	mode := clientconn.NormalMode
	handler := "pg"
	testConnTimeout := time.Duration(0)

	_, ok := register.HandlerFunc["pg"]
	if !ok {
		panic("no pg handler registered")
	}

	logging.Setup(zapcore.ErrorLevel)
	logger := zap.L()

	info := version.Get()

	startFields := []zap.Field{
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
		zap.String("branch", info.Branch),
		zap.Bool("dirty", info.Dirty),
	}
	for _, k := range info.BuildEnvironment.Keys() {
		v := must.NotFail(info.BuildEnvironment.Get(k))
		startFields = append(startFields, zap.Any(k, v))
	}

	go debug.RunHandler(ctx, debugAddr, logger.Named("debug"))

	newHandler := register.HandlerFunc[handler]
	if newHandler == nil {
		logger.Sugar().Fatalf("Unknown backend handler %q.", handler)
	}
	h, err := newHandler(&register.NewHandlerOpts{
		PostgreSQLConnectionString: conf.PostgreSQLConnectionString,
		Ctx:                        ctx,
		Logger:                     logger,
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
	return fdb.mongoURL, nil
}

// transformConnStringPgToMongo parses postgresql connection string and returns a corresponded MongoDB connection string for the driver.
// I.e. transforms "postgres://postgres@127.0.0.1:5432/ferretdb" into mongodb://127.0.0.1:27017
func transformConnStringPgToMongo(connectionURL string) string {
	u, err := url.Parse(connectionURL)
	if err != nil {
		panic(err)
	}

	// change scheme
	if u.Scheme != "postgres" {
		panic("connection url: postgres scheme required")
	}
	u.Scheme = "mongodb"

	// MongoDB collections corresponds to scheme in PostgreSQL
	u.Path = ""

	// set the port to the FerreDB port
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		panic(err)
	}
	port := "27017"
	u.Host = host + ":" + port

	u.User = nil

	return u.String()
}

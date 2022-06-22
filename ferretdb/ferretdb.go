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
	"github.com/FerretDB/FerretDB/internal/handlers/common/register"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// config is a FerretDB library config.
var config Config

// Config ConnectionString contains a backend connection string.
// "postgres://user@postgres:5432/ferretdb" - then it's postgres.
type Config struct {
	ConnectionString string
}

// GetConnectionString returns the backend connection string.
func GetConnectionString() string {
	return config.ConnectionString
}

// Run runs the FerretDB proxy as a library.
func Run(ctx context.Context, conf Config) error {
	config = conf
	listenAddr := "127.0.0.1:27017"
	proxyAddr := "127.0.0.1:37017"
	debugAddr := "127.0.0.1:8088"
	mode := clientconn.NormalMode
	handler := "pg"
	testConnTimeout := time.Duration(0)

	_, ok := register.RegisteredHandlers["pg"]
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

	newHandler := register.RegisteredHandlers[handler]
	if newHandler == nil {
		logger.Sugar().Fatalf("Unknown backend handler %q.", handler)
	}
	h, err := newHandler(&register.NewHandlerOpts{
		Ctx:    ctx,
		Logger: logger,
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

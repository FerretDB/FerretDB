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

// Package provides support for the embedded use-case.
package ferretdb

import (
	"context"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// config with a Connection string
var config Config

// Config ConnectionString contains a string connecting to the backend.
// "postgres://user@postgres:5432/ferretdb" - then it's postgres.
type Config struct {
	ConnectionString string
}

// registeredHandlers maps handler names to constructors.
// The values for `registeredHandlers` must be set through the `init()` functions of the corresponding handlers
// so that we can control which handlers will be included in the build with build tags.
var registeredHandlers = map[string]newHandler{}

// newHandler represents a function that constructs a new handler.
type newHandler func(opts *newHandlerOpts) (handlers.Interface, error)

// newHandlerOpts represents common configuration for constructing handlers.
//
// Handler-specific configuration is passed via command-line flags directly.
type newHandlerOpts struct {
	ctx    context.Context
	logger *zap.Logger
}

// run function that runs embedded proxy until ctx is canceled.
func run(ctx context.Context, conf Config) error {
	config = conf
	listenAddr := "127.0.0.1:27017"
	proxyAddr := "127.0.0.1:37017"
	debugAddr := "127.0.0.1:8088"
	mode := clientconn.NormalMode
	handler := "pg"
	testConnTimeout := time.Duration(0)

	_, ok := registeredHandlers["pg"]
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
	logger.Info("Starting FerretDB "+info.Version+"...", startFields...)

	ctx, stop := notifyAppTermination(context.Background())
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, debugAddr, logger.Named("debug"))

	newHandler := registeredHandlers[handler]
	if newHandler == nil {
		logger.Sugar().Fatalf("Unknown backend handler %q.", handler)
	}
	h, err := newHandler(&newHandlerOpts{
		ctx:    ctx,
		logger: logger,
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

	prometheus.DefaultRegisterer.MustRegister(l)

	err = l.Run(ctx)
	if err == nil || err == context.Canceled {
		logger.Info("Listener stopped")
	} else {
		logger.Error("Listener stopped", zap.Error(err))
	}

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic(err)
	}
	for _, mf := range mfs {
		if _, err := expfmt.MetricFamilyToText(os.Stderr, mf); err != nil {
			panic(err)
		}
	}
	return nil
}

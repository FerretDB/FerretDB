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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

var (
	versionF = flag.Bool("version", false, "print version to stdout (full version, commit, branch, dirty flag) and exit")

	listenAddrF = flag.String("listen-addr", "127.0.0.1:27017", "listen address")
	proxyAddrF  = flag.String("proxy-addr", "127.0.0.1:37017", "proxy address")
	debugAddrF  = flag.String("debug-addr", "127.0.0.1:8088", "debug address")
	modeF       = flag.String("mode", string(clientconn.AllModes[0]), fmt.Sprintf("operation mode: %v", clientconn.AllModes))
	testRecordF = flag.String("test-record", "", "directory of record files with binary data coming from connected clients")

	handlerF = flag.String("handler", "<set in initFlags()>", "<set in initFlags()>")

	postgreSQLURLF = flag.String("postgresql-url", "postgres://postgres@127.0.0.1:5432/ferretdb", "PostgreSQL URL")

	logLevelF = flag.String("log-level", "<set in initFlags()>", "<set in initFlags()>")

	testConnTimeoutF = flag.Duration("test-conn-timeout", 0, "test: set connection timeout")
)

// tigrisURL is a Tigris URL. It is set in the main_tigris.go.
var tigrisURL string

// initFlags improves flags settings after all global flags are initialized
// and all handler constructors are registered.
func initFlags() {
	f := flag.Lookup("handler")
	f.Usage = "backend handler: " + strings.Join(registry.Handlers(), ", ")
	f.DefValue = "pg"
	must.NoError(f.Value.Set(f.DefValue))

	levels := []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
	}

	f = flag.Lookup("log-level")
	f.Usage = "log level: " + strings.Join(levels, ", ")
	f.DefValue = zapcore.DebugLevel.String()
	must.NoError(f.Value.Set(f.DefValue))
}

func main() {
	initFlags()
	flag.Parse()

	level, err := zapcore.ParseLevel(*logLevelF)
	if err != nil {
		log.Fatal(err)
	}
	logging.Setup(level)
	logger := zap.L()

	info := version.Get()

	if *versionF {
		fmt.Fprintln(os.Stdout, "version:", info.Version)
		fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)
		return
	}

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

	var found bool
	for _, m := range clientconn.AllModes {
		if *modeF == string(m) {
			found = true
			break
		}
	}
	if !found {
		logger.Sugar().Fatalf("Unknown mode %q.", *modeF)
	}

	ctx, stop := notifyAppTermination(context.Background())
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, *debugAddrF, logger.Named("debug"))

	h, err := registry.NewHandler(*handlerF, &registry.NewHandlerOpts{
		Ctx:           ctx,
		Logger:        logger,
		PostgreSQLURL: *postgreSQLURLF,
		TigrisURL:     tigrisURL,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      *listenAddrF,
		ProxyAddr:       *proxyAddrF,
		Mode:            clientconn.Mode(*modeF),
		Handler:         h,
		Logger:          logger,
		TestConnTimeout: *testConnTimeoutF,
		TestRecordPath:  *testRecordF,
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
}

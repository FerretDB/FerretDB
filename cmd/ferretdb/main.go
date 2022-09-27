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
	"time"

	"github.com/alecthomas/kong"
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

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
var cli struct {
	Version bool `default:"false" help:"print version to stdout (full version, commit, branch, dirty flag) and exit."`

	ListenAddr string `name:"listen-addr" default:"127.0.0.1:27017" help:"listen address."`
	ProxyAddr  string `name:"proxy-addr" default:"127.0.0.1:37017" help:"proxy address."`
	DebugAddr  string `name:"debug-addr" default:"127.0.0.1:8088" help:"debug address."`
	Mode       string `default:"${default_mode}" help:"${help_mode}."`
	TestRecord string `name:"test-record" default:"" help:"directory of record files with binary data coming from connected clients."`

	Handler string `default:"pg" help:"${help_handler}."`

	PostgresURL string `name:"postgresql-url" default:"postgres://postgres@127.0.0.1:5432/ferretdb" help:"PostgreSQL URL."`

	LogLevel string `name:"log-level" default:"${default_logLevel}" help:"${help_logLevel}."`

	TestConnTimeout time.Duration `name:"test-conn-timeout" default:"0" help:"test: set connection timeout."`
}

var (
//versionF = flag.Bool("version", false, "print version to stdout (full version, commit, branch, dirty flag) and exit")

//listenAddrF = flag.String("listen-addr", "127.0.0.1:27017", "listen address")
//proxyAddrF  = flag.String("proxy-addr", "127.0.0.1:37017", "proxy address")
//debugAddrF  = flag.String("debug-addr", "127.0.0.1:8088", "debug address")
//modeF = flag.String("mode", string(clientconn.AllModes[0]), fmt.Sprintf("operation mode: %v", clientconn.AllModes))
//testRecordF = flag.String("test-record", "", "directory of record files with binary data coming from connected clients")

//handlerF = flag.String("handler", "<set in initFlags()>", "<set in initFlags()>")

//postgreSQLURLF = flag.String("postgresql-url", "postgres://postgres@127.0.0.1:5432/ferretdb", "PostgreSQL URL")

//logLevelF = flag.String("log-level", "<set in initFlags()>", "<set in initFlags()>")

//testConnTimeoutF = flag.Duration("test-conn-timeout", 0, "test: set connection timeout")
)

// Tigris parameters that are set at main_tigris.go.
var (
	tigrisClientID     string
	tigrisClientSecret string
	tigrisToken        string
	tigrisURL          string
)

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
	levels := []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
	}

	kongCtx := kong.Parse(&cli,
		kong.Vars{
			"default_mode":     string(clientconn.AllModes[0]),
			"help_mode":        fmt.Sprintf("operation mode: %v", clientconn.AllModes),
			"help_handler":     "backend handler: " + strings.Join(registry.Handlers(), ", "),
			"default_logLevel": zapcore.DebugLevel.String(),
			"help_logLevel":    "log level: " + strings.Join(levels, ", "),
		})
	kongCtx.Flags()[0].Help = "LEDU"

	log.Print(kongCtx.Command())

	//initFlags()
	//flag.Parse()

	level, err := zapcore.ParseLevel(cli.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logging.Setup(level)
	logger := zap.L()

	info := version.Get()

	if cli.Version {
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
		if cli.Mode == string(m) {
			found = true
			break
		}
	}
	if !found {
		logger.Sugar().Fatalf("Unknown mode %q.", cli.Mode)
	}

	ctx, stop := notifyAppTermination(context.Background())
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, cli.DebugAddr, logger.Named("debug"))

	h, err := registry.NewHandler(cli.Handler, &registry.NewHandlerOpts{
		Ctx:    ctx,
		Logger: logger,

		PostgreSQLURL: cli.PostgresURL,

		TigrisClientID:     tigrisClientID,
		TigrisClientSecret: tigrisClientSecret,
		TigrisToken:        tigrisToken,
		TigrisURL:          tigrisURL,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      cli.ListenAddr,
		ProxyAddr:       cli.ProxyAddr,
		Mode:            clientconn.Mode(cli.Mode),
		Handler:         h,
		Logger:          logger,
		TestConnTimeout: cli.TestConnTimeout,
		TestRecordPath:  cli.TestRecord,
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

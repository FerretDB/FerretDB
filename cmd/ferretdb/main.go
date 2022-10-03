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
	"golang.org/x/exp/slices"

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
	Version bool `default:"false" help:"Print version to stdout (full version, commit, branch, dirty flag) and exit."`

	ListenAddr string `default:"127.0.0.1:27017" help:"Listen address."`
	ProxyAddr  string `default:"127.0.0.1:37017" help:"Proxy address."`
	DebugAddr  string `default:"127.0.0.1:8088" help:"Debug address."`
	Mode       string `default:"${default_mode}" help:"${help_mode}."`
	TestRecord string `default:"" help:"Directory of record files with binary data coming from connected clients."`

	Handler string `default:"pg" help:"${help_handler}."`

	PostgresURL string `name:"postgresql-url" default:"postgres://postgres@127.0.0.1:5432/ferretdb" help:"PostgreSQL URL."`

	LogLevel string `default:"${default_logLevel}" help:"${help_logLevel}."`

	TestConnTimeout time.Duration `default:"0" help:"Test: set connection timeout."`

	kong.Plugins
}

// Additional variables for the kong parsers.
var (
	logLevels = []string{
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
	}

	kongOptions = []kong.Option{
		kong.Vars{
			"default_logLevel": zapcore.DebugLevel.String(),
			"default_mode":     string(clientconn.AllModes[0]),
			"help_handler":     "Backend handler: " + strings.Join(registry.Handlers(), ", "),
			"help_logLevel":    "Log level: " + strings.Join(logLevels, ", "),
			"help_mode":        fmt.Sprintf("Operation mode: %v", clientconn.AllModes),
		},
	}
)

// Tigris parameters that are set at main_tigris.go.
var (
	tigrisClientID     string
	tigrisClientSecret string
	tigrisToken        string
	tigrisURL          string
)

func main() {
	kong.Parse(&cli, kongOptions...)

	run()
}

// run sets up environment based on provided flags and runs FerretDB.
func run() {
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

	if !slices.Contains(clientconn.AllModes, clientconn.Mode(cli.Mode)) {
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

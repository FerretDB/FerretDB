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
	"path/filepath"
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
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
var cli struct {
	ListenAddr string `default:"127.0.0.1:27017"      help:"Listen address."`
	ProxyAddr  string `default:"127.0.0.1:37017"      help:"Proxy address."`
	DebugAddr  string `default:"127.0.0.1:8088"       help:"${help_debug_addr}"`
	StateDir   string `default:"."                    help:"Process state directory."`
	Mode       string `default:"${default_mode}"      help:"${help_mode}"             enum:"${enum_mode}"`

	Log struct {
		Level string `default:"${default_log_level}" help:"${help_log_level}"`
		UUID  bool   `default:"false"                help:"Add instance UUID to all log messages."`
	} `embed:"" prefix:"log-"`

	MetricsUUID bool `default:"false" help:"Add instance UUID to all metrics."`

	Handler string `default:"pg" help:"${help_handler}"`

	PostgresURL string `default:"postgres://postgres@127.0.0.1:5432/ferretdb" help:"PostgreSQL URL for 'pg' handler."`

	// Put flags for other handlers there, between --postgres-url and --version in the help output.
	kong.Plugins

	Version bool `default:"false" help:"Print version to stdout and exit."`

	Test struct {
		ConnTimeout time.Duration `default:"0" help:"Testing flag: client connection timeout."`
		RecordsDir  string        `default:""  help:"Testing flag: directory for record files."`
	} `embed:"" prefix:"test-"`
}

// Additional variables for the kong parsers.
var (
	logLevels = []string{
		zap.DebugLevel.String(),
		zap.InfoLevel.String(),
		zap.WarnLevel.String(),
		zap.ErrorLevel.String(),
	}

	kongOptions = []kong.Option{
		kong.Vars{
			"default_log_level": zap.DebugLevel.String(),
			"default_mode":      clientconn.AllModes[0],

			"help_debug_addr": "Debug address for /debug/metrics, /debug/pprof, and similar HTTP handlers.",
			"help_log_level": fmt.Sprintf(
				"Log level: '%s'. Debug level also enables development mode.",
				strings.Join(logLevels, "', '"),
			),
			"help_mode":    fmt.Sprintf("Operation mode: '%s'.", strings.Join(clientconn.AllModes, "', '")),
			"help_handler": fmt.Sprintf("Backend handler: '%s'.", strings.Join(registry.Handlers(), "', '")),

			"enum_mode": strings.Join(clientconn.AllModes, ","),
		},
		kong.DefaultEnvars("FERRETDB"),
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

// setupState setups state provider.
func setupState() (*state.Provider, string) {
	f, err := filepath.Abs(filepath.Join(cli.StateDir, "state.json"))
	if err != nil {
		log.Fatalf("Failed to get path for state file: %s.", err)
	}

	p, err := state.NewProvider(f)
	if err != nil {
		log.Fatalf("Failed to create state provider: %s.", err)
	}

	s, err := p.Get()
	if err != nil {
		log.Fatalf("Failed to get state: %s.", err)
	}

	return p, s.UUID
}

// setupMetrics setups Prometheus metrics registerer with some metrics.
func setupMetrics(stateProvider *state.Provider, uuid string) prometheus.Registerer {
	r := prometheus.WrapRegistererWith(
		prometheus.Labels{"uuid": uuid},
		prometheus.DefaultRegisterer,
	)
	m := stateProvider.MetricsCollector(false)

	// Unless requested, don't add UUID to all metrics, but add it to one.
	// See https://prometheus.io/docs/instrumenting/writing_exporters/#target-labels-not-static-scraped-labels
	if !cli.MetricsUUID {
		r = prometheus.DefaultRegisterer
		m = stateProvider.MetricsCollector(true)
	}

	r.MustRegister(m)

	return r
}

// setupLogger setups zap logger.
func setupLogger(uuid string) *zap.Logger {
	info := version.Get()

	logUUID := uuid
	startupFields := []zap.Field{
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
		zap.String("branch", info.Branch),
		zap.Bool("dirty", info.Dirty),
		zap.Bool("debug", info.Debug),
		zap.Reflect("buildEnvironment", info.BuildEnvironment.Map()),
	}

	// Similarly to Prometheus, unless requested, don't add UUID to all messages, but log it once at startup.
	if !cli.Log.UUID {
		logUUID = ""
		startupFields = append(startupFields, zap.String("uuid", uuid))
	}

	level, err := zapcore.ParseLevel(cli.Log.Level)
	if err != nil {
		log.Fatal(err)
	}

	logging.Setup(level, logUUID)
	l := zap.L()

	l.Info("Starting FerretDB "+info.Version+"...", startupFields...)

	return l
}

// run sets up environment based on provided flags and runs FerretDB.
func run() {
	if cli.Version {
		info := version.Get()

		fmt.Fprintln(os.Stdout, "version:", info.Version)
		fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)

		return
	}

	stateProvider, uuid := setupState()

	metricsRegisterer := setupMetrics(stateProvider, uuid)

	logger := setupLogger(uuid)

	ctx, stop := notifyAppTermination(context.Background())
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, cli.DebugAddr, metricsRegisterer, logger.Named("debug"))

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
		TestConnTimeout: cli.Test.ConnTimeout,
		TestRecordsDir:  cli.Test.RecordsDir,
	})

	metricsRegisterer.MustRegister(l)

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

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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/telemetry"
)

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
//
// Keep order in sync with documentation.
var cli struct {
	Version  bool   `default:"false" help:"Print version to stdout and exit." env:"-"`
	Handler  string `default:"pg" help:"${help_handler}"`
	Mode     string `default:"${default_mode}" help:"${help_mode}" enum:"${enum_mode}"`
	StateDir string `default:"."               help:"Process state directory."`

	Listen struct {
		Addr        string `default:"127.0.0.1:27017" help:"Listen TCP address."`
		Unix        string `default:""                help:"Listen Unix domain socket path."`
		TLS         string `default:""                help:"Listen TLS address."`
		TLSCertFile string `default:""                help:"TLS cert file path."`
		TLSKeyFile  string `default:""                help:"TLS key file path."`
		TLSCAFile   string `default:""                help:"TLS CA file path." name:"tls-ca-file"`
	} `embed:"" prefix:"listen-"`

	ProxyAddr string `default:""                help:"Proxy address."`
	DebugAddr string `default:"127.0.0.1:8088"  help:"Listen address for HTTP handlers for metrics, pprof, etc."`

	PostgreSQLURL string `name:"postgresql-url" default:"${default_postgresql_url}" help:"PostgreSQL URL for 'pg' handler."`

	// see setCLIPlugins
	kong.Plugins

	Log struct {
		Level string `default:"${default_log_level}" help:"${help_log_level}"`
		UUID  bool   `default:"false"                help:"Add instance UUID to all log messages." negatable:""`
	} `embed:"" prefix:"log-"`

	MetricsUUID bool `default:"false" help:"Add instance UUID to all metrics." negatable:""`

	Telemetry telemetry.Flag `default:"undecided" help:"Enable or disable basic telemetry. See https://beacon.ferretdb.io."`

	Test struct {
		RecordsDir      string `default:"" help:"Experimental: directory for record files."`
		DisablePushdown bool   `default:"false" help:"Experimental: disable query pushdown."`
		EnableCursors   bool   `default:"false" help:"Experimental: enable cursors."`

		//nolint:lll // for readability
		Telemetry struct {
			URL            string        `default:"https://beacon.ferretdb.io/" help:"Experimental: telemetry: reporting URL."`
			UndecidedDelay time.Duration `default:"1h"                          help:"${help_telemetry_undecided_delay}"`
			ReportInterval time.Duration `default:"24h"                         help:"Experimental: telemetry: report interval."`
			ReportTimeout  time.Duration `default:"5s"                          help:"Experimental: telemetry: report timeout."`
			Package        string        `default:""                            help:"Experimental: telemetry: custom package type."`
		} `embed:"" prefix:"telemetry-"`
	} `embed:"" prefix:"test-"`
}

// The tigrisFlags struct represents flags that are used specifically by "tigris" handler.
//
// See main_tigris.go.
var tigrisFlags struct {
	TigrisURL          string `default:"127.0.0.1:8081" help:"Tigris URL for 'tigris' handler."`
	TigrisClientID     string `default:""               help:"Tigris Client ID."`
	TigrisClientSecret string `default:""               help:"Tigris Client secret."`
}

// The hanaFlags struct represents flags that are used specifically by "hana" handler.
//
// See main_hana.go.
var hanaFlags struct {
	HANAURL string `name:"hana-url" help:"SAP HANA URL for 'hana' handler"`
}

// handlerFlags is a map of handler names to their flags.
var handlerFlags = map[string]any{}

// setCLIPlugins adds Kong flags for handlers in the stable order.
func setCLIPlugins() {
	handlers := maps.Keys(handlerFlags)
	slices.Sort(handlers)

	for _, h := range handlers {
		cli.Plugins = append(cli.Plugins, handlerFlags[h])
	}
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
			"default_log_level":      defaultLogLevel().String(),
			"default_mode":           clientconn.AllModes[0],
			"default_postgresql_url": "postgres://127.0.0.1:5432/ferretdb",

			"help_log_level":                 fmt.Sprintf("Log level: '%s'.", strings.Join(logLevels, "', '")),
			"help_mode":                      fmt.Sprintf("Operation mode: '%s'.", strings.Join(clientconn.AllModes, "', '")),
			"help_handler":                   fmt.Sprintf("Backend handler: '%s'.", strings.Join(registry.Handlers(), "', '")),
			"help_telemetry_undecided_delay": "Experimental: telemetry: delay for undecided state.",

			"enum_mode": strings.Join(clientconn.AllModes, ","),
		},
		kong.DefaultEnvars("FERRETDB"),
	}
)

func main() {
	setCLIPlugins()
	kong.Parse(&cli, kongOptions...)

	run()
}

// defaultLogLevel returns the default log level.
func defaultLogLevel() zapcore.Level {
	if version.Get().DebugBuild {
		return zap.DebugLevel
	}

	return zap.InfoLevel
}

// setupState setups state provider.
func setupState() *state.Provider {
	f, err := filepath.Abs(filepath.Join(cli.StateDir, "state.json"))
	if err != nil {
		log.Fatalf("Failed to get path for state file: %s.", err)
	}

	p, err := state.NewProvider(f)
	if err != nil {
		log.Fatalf("Failed to create state provider: %s.", err)
	}

	return p
}

// setupMetrics setups Prometheus metrics registerer with some metrics.
func setupMetrics(stateProvider *state.Provider) prometheus.Registerer {
	r := prometheus.WrapRegistererWith(
		prometheus.Labels{"uuid": stateProvider.Get().UUID},
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
func setupLogger(stateProvider *state.Provider) *zap.Logger {
	info := version.Get()

	startupFields := []zap.Field{
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
		zap.String("branch", info.Branch),
		zap.Bool("dirty", info.Dirty),
		zap.String("package", info.Package),
		zap.Bool("debugBuild", info.DebugBuild),
		zap.Any("buildEnvironment", info.BuildEnvironment.Map()),
	}
	logUUID := stateProvider.Get().UUID

	// Similarly to Prometheus, unless requested, don't add UUID to all messages, but log it once at startup.
	if !cli.Log.UUID {
		startupFields = append(startupFields, zap.String("uuid", logUUID))
		logUUID = ""
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

// runTelemetryReporter runs telemetry reporter until ctx is canceled.
func runTelemetryReporter(ctx context.Context, opts *telemetry.NewReporterOpts) {
	r, err := telemetry.NewReporter(opts)
	if err != nil {
		opts.L.Sugar().Fatalf("Failed to create telemetry reporter: %s.", err)
	}

	r.Run(ctx)
}

// dumpMetrics dumps all Prometheus metrics to stderr.
func dumpMetrics() {
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

// run sets up environment based on provided flags and runs FerretDB.
func run() {
	info := version.Get()

	if p := cli.Test.Telemetry.Package; p != "" {
		info.Package = p
	}

	if cli.Version {
		fmt.Fprintln(os.Stdout, "version:", info.Version)
		fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)
		fmt.Fprintln(os.Stdout, "package:", info.Package)
		fmt.Fprintln(os.Stdout, "debugBuild:", info.DebugBuild)

		return
	}

	stateProvider := setupState()

	metricsRegisterer := setupMetrics(stateProvider)

	logger := setupLogger(stateProvider)

	ctx, stop := notifyAppTermination(context.Background())

	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		debug.RunHandler(ctx, cli.DebugAddr, metricsRegisterer, logger.Named("debug"))
	}()

	metrics := connmetrics.NewListenerMetrics()

	wg.Add(1)

	go func() {
		defer wg.Done()
		runTelemetryReporter(
			ctx,
			&telemetry.NewReporterOpts{
				URL:            cli.Test.Telemetry.URL,
				F:              &cli.Telemetry,
				DNT:            os.Getenv("DO_NOT_TRACK"),
				ExecName:       os.Args[0],
				P:              stateProvider,
				ConnMetrics:    metrics.ConnMetrics,
				L:              logger.Named("telemetry"),
				UndecidedDelay: cli.Test.Telemetry.UndecidedDelay,
				ReportInterval: cli.Test.Telemetry.ReportInterval,
				ReportTimeout:  cli.Test.Telemetry.ReportTimeout,
			},
		)
	}()

	h, err := registry.NewHandler(cli.Handler, &registry.NewHandlerOpts{
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: stateProvider,

		PostgreSQLURL: cli.PostgreSQLURL,

		TigrisURL:          tigrisFlags.TigrisURL,
		TigrisClientID:     tigrisFlags.TigrisClientID,
		TigrisClientSecret: tigrisFlags.TigrisClientSecret,

		HANAURL: hanaFlags.HANAURL,

		TestOpts: registry.TestOpts{
			DisablePushdown: cli.Test.DisablePushdown,
			EnableCursors:   cli.Test.EnableCursors,
		},
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer h.Close()

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		TCP:         cli.Listen.Addr,
		Unix:        cli.Listen.Unix,
		TLS:         cli.Listen.TLS,
		TLSCertFile: cli.Listen.TLSCertFile,
		TLSKeyFile:  cli.Listen.TLSKeyFile,
		TLSCAFile:   cli.Listen.TLSCAFile,

		ProxyAddr:      cli.ProxyAddr,
		Mode:           clientconn.Mode(cli.Mode),
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: cli.Test.RecordsDir,
	})

	metricsRegisterer.MustRegister(l)

	err = l.Run(ctx)
	if err == nil || errors.Is(err, context.Canceled) {
		logger.Info("Listener stopped")
	} else {
		logger.Error("Listener stopped", zap.Error(err))
	}

	stop()

	wg.Wait()

	if info.DebugBuild {
		dumpMetrics()
	}
}

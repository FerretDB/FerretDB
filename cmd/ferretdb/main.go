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
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/automaxprocs/maxprocs"
	_ "golang.org/x/crypto/x509roots/fallback" // register root TLS certificates for production Docker image

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handler/registry"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/telemetry"
)

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
//
// Keep order in sync with documentation.
var cli struct {
	// We hide `run` command to show only `ping` in the help message.
	Run  struct{} `cmd:"" default:"1"                             hidden:""`
	Ping struct{} `cmd:"" help:"Ping existing FerretDB instance."`

	Version     bool   `default:"false"           help:"Print version to stdout and exit." env:"-"`
	Handler     string `default:"postgresql"      help:"${help_handler}"`
	Mode        string `default:"${default_mode}" help:"${help_mode}"                      enum:"${enum_mode}"`
	StateDir    string `default:"."               help:"Process state directory."`
	ReplSetName string `default:""                help:"Replica set name."`

	Listen struct {
		Addr        string `default:"127.0.0.1:27017" help:"Listen TCP address."`
		Unix        string `default:""                help:"Listen Unix domain socket path."`
		TLS         string `default:""                help:"Listen TLS address."`
		TLSCertFile string `default:""                help:"TLS cert file path."`
		TLSKeyFile  string `default:""                help:"TLS key file path."`
		TLSCaFile   string `default:""                help:"TLS CA file path."`
	} `embed:"" prefix:"listen-"`

	Proxy struct {
		Addr        string `default:"" help:"Proxy address."`
		TLSCertFile string `default:"" help:"Proxy TLS cert file path."`
		TLSKeyFile  string `default:"" help:"Proxy TLS key file path."`
		TLSCaFile   string `default:"" help:"Proxy TLS CA file path."`
	} `embed:"" prefix:"proxy-"`

	DebugAddr string `default:"127.0.0.1:8088" help:"Listen address for HTTP handlers for metrics, profiling, etc."`

	// see setCLIPlugins
	kong.Plugins

	Setup struct {
		Database string        `default:""    help:"Setup database during backend initialization."`
		Username string        `default:""    help:"Setup user during backend initialization."`
		Password string        `default:""    help:"Setup user's password."`
		Timeout  time.Duration `default:"30s" help:"Setup timeout."`
	} `embed:"" prefix:"setup-"`

	Log struct {
		Level  string `default:"${default_log_level}" help:"${help_log_level}"`
		Format string `default:"console"              help:"${help_log_format}"                     enum:"${enum_log_format}"`
		UUID   bool   `default:"false"                help:"Add instance UUID to all log messages." negatable:""`
	} `embed:"" prefix:"log-"`

	MetricsUUID bool `default:"false" help:"Add instance UUID to all metrics." negatable:""`

	OTel struct {
		Traces struct {
			URL string `default:"" help:"OpenTelemetry OTLP/HTTP traces endpoint URL (e.g. 'http://host:4318/v1/traces')."`
		} `embed:"" prefix:"traces-"`
	} `embed:"" prefix:"otel-"`

	Telemetry telemetry.Flag `default:"undecided" help:"Enable or disable basic telemetry. See https://beacon.ferretdb.com."`

	Test struct {
		RecordsDir string `default:"" help:"Testing: directory for record files."`

		DisablePushdown      bool `default:"false" help:"Experimental: disable pushdown."`
		EnableNestedPushdown bool `default:"false" help:"Experimental: enable pushdown for dot notation."`

		CappedCleanup struct {
			Interval   time.Duration `default:"1m" help:"Experimental: capped collections cleanup interval."`
			Percentage uint8         `default:"10" help:"Experimental: percentage of documents to cleanup."`
		} `embed:"" prefix:"capped-cleanup-"`

		EnableNewAuth bool `default:"false" help:"Experimental: enable new authentication."`

		BatchSize            int `default:"100" help:"Experimental: maximum insertion batch size."`
		MaxBsonObjectSizeMiB int `default:"16"  help:"Experimental: maximum BSON object size in MiB."`

		Telemetry struct {
			URL            string        `default:"https://beacon.ferretdb.com/" help:"Telemetry: reporting URL."`
			UndecidedDelay time.Duration `default:"1h"                           help:"Telemetry: delay for undecided state."`
			ReportInterval time.Duration `default:"24h"                          help:"Telemetry: report interval."`
			ReportTimeout  time.Duration `default:"5s"                           help:"Telemetry: report timeout."`
			Package        string        `default:""                             help:"Telemetry: custom package type."`
		} `embed:"" prefix:"telemetry-"`
	} `embed:"" prefix:"test-"`
}

// The postgreSQLFlags struct represents flags that are used by the "postgresql" backend.
//
// See main_postgresql.go.
//
//nolint:lll // some tags are long
var postgreSQLFlags struct {
	PostgreSQLURL string `name:"postgresql-url" default:"postgres://127.0.0.1:5432/ferretdb" help:"PostgreSQL URL for 'postgresql' handler."`
}

// The sqliteFlags struct represents flags that are used by the "sqlite" backend.
//
// See main_sqlite.go.
var sqliteFlags struct {
	SQLiteURL string `name:"sqlite-url" default:"file:data/" help:"SQLite URI (directory) for 'sqlite' handler."`
}

// The mySQLFlags struct represents flags that are used by the "mysql" backend.
//
// See main_mysql.go.
var mySQLFlags struct {
	MySQLURL string `name:"mysql-url" default:"mysql://127.0.0.1:3306/ferretdb" help:"MySQL URL for 'mysql' handler" hidden:""`
}

// The hanaFlags struct represents flags that are used by the "hana" handler.
//
// See main_hana.go.
var hanaFlags struct {
	HANAURL string `name:"hana-url" help:"SAP HANA URL for 'hana' handler"`
}

// handlerFlags is a map of handler names to their flags.
var handlerFlags = map[string]any{}

// setCLIPlugins adds Kong flags for handlers in the right order.
func setCLIPlugins() {
	handlers := registry.Handlers()

	if len(handlers) != len(handlerFlags) {
		panic("handlers and handlerFlags are not in sync")
	}

	for _, h := range handlers {
		f := handlerFlags[h]
		if f == nil {
			panic(fmt.Sprintf("handler %q has no flags", h))
		}

		cli.Plugins = append(cli.Plugins, f)
	}
}

// Additional variables for the kong parsers.
var (
	logLevels = []string{
		slog.LevelDebug.String(),
		slog.LevelInfo.String(),
		slog.LevelWarn.String(),
		slog.LevelError.String(),
	}

	logFormats = []string{"console", "text", "json"}

	kongOptions = []kong.Option{
		kong.HelpOptions{
			Compact: true,
		},
		kong.Vars{
			"default_log_level": defaultLogLevel().String(),
			"default_mode":      clientconn.AllModes[0],

			"enum_log_format": strings.Join(logFormats, ","),
			"enum_mode":       strings.Join(clientconn.AllModes, ","),

			"help_handler":    fmt.Sprintf("Backend handler: '%s'.", strings.Join(registry.Handlers(), "', '")),
			"help_log_format": fmt.Sprintf("Log format: '%s'.", strings.Join(logFormats, "', '")),
			"help_log_level":  fmt.Sprintf("Log level: '%s'.", strings.Join(logLevels, "', '")),
			"help_mode":       fmt.Sprintf("Operation mode: '%s'.", strings.Join(clientconn.AllModes, "', '")),
		},
		kong.DefaultEnvars("FERRETDB"),
	}
)

func main() {
	setCLIPlugins()
	ctx := kong.Parse(&cli, kongOptions...)

	switch ctx.Command() {
	case "run":
		run()
	case "ping":
		logger := setupLogger(cli.Log.Format, "")
		checkFlags(logger)

		ready := ReadyZ{
			l: logger,
		}

		ctx, stop := ctxutil.SigTerm(context.Background())
		defer stop()

		if !ready.Probe(ctx) {
			os.Exit(1)
		}

	default:
		panic("unknown sub-command")
	}
}

// defaultLogLevel returns the default log level.
func defaultLogLevel() slog.Level {
	if version.Get().DebugBuild {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

// setupState setups state provider.
func setupState() *state.Provider {
	var f string

	if dir := cli.StateDir; dir != "" && dir != "-" {
		var err error
		if f, err = filepath.Abs(filepath.Join(dir, "state.json")); err != nil {
			log.Fatalf("Failed to get path for state file: %s.", err)
		}
	}

	sp, err := state.NewProvider(f)
	if err != nil {
		log.Fatal(stateFileProblem(f, err))
	}

	return sp
}

// setupMetrics setups Prometheus metrics registerer with some metrics.
func setupMetrics(stateProvider *state.Provider) prometheus.Registerer {
	r := prometheus.DefaultRegisterer
	m := stateProvider.MetricsCollector(true)

	// we don't do it by default due to
	// https://prometheus.io/docs/instrumenting/writing_exporters/#target-labels-not-static-scraped-labels
	if cli.MetricsUUID {
		r = prometheus.WrapRegistererWith(
			prometheus.Labels{"uuid": stateProvider.Get().UUID},
			prometheus.DefaultRegisterer,
		)
		m = stateProvider.MetricsCollector(false)
	}

	r.MustRegister(m)

	return r
}

// setupLogger creates a logger with the level defined from cli.
func setupLogger(format string, uuid string) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cli.Log.Level)); err != nil {
		log.Fatal(err)
	}

	opts := &logging.NewHandlerOpts{
		Base:  format,
		Level: level,
	}
	logging.Setup(opts, uuid)

	return slog.Default()
}

// checkFlags checks that CLI flags are not self-contradictory.
func checkFlags(l *slog.Logger) {
	ctx := context.Background()

	if cli.Setup.Database != "" && !cli.Test.EnableNewAuth {
		l.LogAttrs(ctx, logging.LevelFatal, "--setup-database requires --test-enable-new-auth")
	}

	if (cli.Setup.Database == "") != (cli.Setup.Username == "") {
		l.LogAttrs(ctx, logging.LevelFatal, "--setup-database should be used together with --setup-username")
	}

	if cli.Test.DisablePushdown && cli.Test.EnableNestedPushdown {
		l.LogAttrs(
			ctx,
			logging.LevelFatal,
			"--test-disable-pushdown and --test-enable-nested-pushdown should not be set at the same time",
		)
	}
}

// dumpMetrics dumps all Prometheus metrics to stderr.
func dumpMetrics() {
	mfs := must.NotFail(prometheus.DefaultGatherer.Gather())

	for _, mf := range mfs {
		must.NotFail(expfmt.MetricFamilyToText(os.Stderr, mf))
	}
}

// run sets up environment based on provided flags and runs FerretDB.
func run() {
	// to increase a chance of resource finalizers to spot problems
	if debugbuild.Enabled {
		defer func() {
			runtime.GC()
			runtime.GC()
		}()
	}

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

	// safe to always enable
	runtime.SetBlockProfileRate(10000)

	stateProvider := setupState()

	metricsRegisterer := setupMetrics(stateProvider)

	startupFields := []slog.Attr{
		slog.String("version", info.Version),
		slog.String("commit", info.Commit),
		slog.String("branch", info.Branch),
		slog.Bool("dirty", info.Dirty),
		slog.String("package", info.Package),
		slog.Bool("debugBuild", info.DebugBuild),
		slog.Any("buildEnvironment", info.BuildEnvironment),
	}
	logUUID := stateProvider.Get().UUID

	// Similarly to Prometheus, unless requested, don't add UUID to all messages, but log it once at startup.
	if !cli.Log.UUID {
		startupFields = append(startupFields, slog.String("uuid", logUUID))
		logUUID = ""
	}

	logger := setupLogger(cli.Log.Format, logUUID)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "Starting FerretDB "+info.Version+"...", startupFields...)

	if debugbuild.Enabled {
		logger.Info("This is a debug build. The performance will be affected.")
	}

	if logger.Enabled(context.Background(), slog.LevelDebug) {
		logger.Info("Debug logging enabled. The security and performance will be affected.")
	}

	checkFlags(logger)

	if _, err := maxprocs.Set(maxprocs.Logger(func(format string, a ...any) {
		logger.Debug(fmt.Sprintf(format, a...))
	})); err != nil {
		logger.Warn("Failed to set GOMAXPROCS", logging.Error(err))
	}

	ctx, stop := ctxutil.SigTerm(context.Background())

	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")

		// second SIGTERM should immediately stop the process
		stop()
	}()

	// used to start debug handler with probes as soon as possible, even before listener is created
	var listener atomic.Pointer[clientconn.Listener]

	var wg sync.WaitGroup

	if addr := cli.DebugAddr; addr != "" && addr != "-" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			l := logging.WithName(logger, "debug")
			ready := ReadyZ{
				l: l,
			}

			h, err := debug.Listen(&debug.ListenOpts{
				TCPAddr: addr,
				L:       l,
				R:       metricsRegisterer,
				Livez: func(context.Context) bool {
					if listener.Load() == nil {
						return false
					}

					return listener.Load().Listening()
				},
				Readyz: ready.Probe,
			})
			if err != nil {
				l.LogAttrs(ctx, logging.LevelFatal, "Failed to create debug handler", logging.Error(err))
			}

			h.Serve(ctx)
		}()
	}

	if u := cli.OTel.Traces.URL; u != "" && u != "-" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			l := logging.WithName(logger, "otel")

			ot, err := observability.NewOTelTraceExporter(&observability.OTelTraceExporterOpts{
				Logger:  l,
				Service: "ferretdb",
				Version: version.Get().Version,
				URL:     u,
			})
			if err != nil {
				l.LogAttrs(ctx, logging.LevelFatal, "Failed to create Otel tracer", logging.Error(err))
			}

			ot.Run(ctx)
		}()
	}

	metrics := connmetrics.NewListenerMetrics()

	wg.Add(1)

	go func() {
		defer wg.Done()

		l := logging.WithName(logger, "telemetry")
		opts := &telemetry.NewReporterOpts{
			URL:            cli.Test.Telemetry.URL,
			F:              &cli.Telemetry,
			DNT:            os.Getenv("DO_NOT_TRACK"),
			ExecName:       os.Args[0],
			P:              stateProvider,
			ConnMetrics:    metrics.ConnMetrics,
			L:              l,
			UndecidedDelay: cli.Test.Telemetry.UndecidedDelay,
			ReportInterval: cli.Test.Telemetry.ReportInterval,
			ReportTimeout:  cli.Test.Telemetry.ReportTimeout,
		}

		r, err := telemetry.NewReporter(opts)
		if err != nil {
			l.LogAttrs(ctx, logging.LevelFatal, "Failed to create telemetry reporter", logging.Error(err))
		}

		r.Run(ctx)
	}()

	h, closeBackend, err := registry.NewHandler(cli.Handler, &registry.NewHandlerOpts{
		Logger:        logger,
		ConnMetrics:   metrics.ConnMetrics,
		StateProvider: stateProvider,
		TCPHost:       cli.Listen.Addr,
		ReplSetName:   cli.ReplSetName,

		SetupDatabase: cli.Setup.Database,
		SetupUsername: cli.Setup.Username,
		SetupPassword: password.WrapPassword(cli.Setup.Password),
		SetupTimeout:  cli.Setup.Timeout,

		PostgreSQLURL: postgreSQLFlags.PostgreSQLURL,

		SQLiteURL: sqliteFlags.SQLiteURL,

		HANAURL: hanaFlags.HANAURL,

		MySQLURL: mySQLFlags.MySQLURL,

		TestOpts: registry.TestOpts{
			DisablePushdown:         cli.Test.DisablePushdown,
			EnableNestedPushdown:    cli.Test.EnableNestedPushdown,
			CappedCleanupInterval:   cli.Test.CappedCleanup.Interval,
			CappedCleanupPercentage: cli.Test.CappedCleanup.Percentage,
			EnableNewAuth:           cli.Test.EnableNewAuth,
			BatchSize:               cli.Test.BatchSize,
			MaxBsonObjectSizeBytes:  cli.Test.MaxBsonObjectSizeMiB * 1024 * 1024,
		},
	})
	if err != nil {
		logger.LogAttrs(ctx, logging.LevelFatal, "Failed to construct handler", logging.Error(err))
	}

	defer closeBackend()

	l, err := clientconn.Listen(&clientconn.NewListenerOpts{
		TCP:  cli.Listen.Addr,
		Unix: cli.Listen.Unix,

		TLS:         cli.Listen.TLS,
		TLSCertFile: cli.Listen.TLSCertFile,
		TLSKeyFile:  cli.Listen.TLSKeyFile,
		TLSCAFile:   cli.Listen.TLSCaFile,

		ProxyAddr:        cli.Proxy.Addr,
		ProxyTLSCertFile: cli.Proxy.TLSCertFile,
		ProxyTLSKeyFile:  cli.Proxy.TLSKeyFile,
		ProxyTLSCAFile:   cli.Proxy.TLSCaFile,

		Mode:           clientconn.Mode(cli.Mode),
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: cli.Test.RecordsDir,
	})
	if err != nil {
		logger.LogAttrs(ctx, logging.LevelFatal, "Failed to construct listener", logging.Error(err))
	}

	listener.Store(l)

	metricsRegisterer.MustRegister(l)

	l.Run(ctx)

	logger.Info("Listener stopped")

	wg.Wait()

	if info.DebugBuild {
		dumpMetrics()
	}
}

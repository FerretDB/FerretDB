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

// Package ferretdb provides embeddable FerretDB implementation.
package ferretdb

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/debug"
	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/telemetry"
)

// RunOpts represents configurable options for running FerretDB.
type RunOpts struct { //nolint:vet // used only for configuration
	Version bool

	PostgreSQLURL string

	Listen ListenOpts

	Proxy ProxyOpts

	DebugAddr string

	Mode     string
	StateDir string
	Auth     bool

	Log LogOpts

	MetricsUUID bool

	OTel OTelOpts

	Telemetry telemetry.Flag

	Dev DevOpts
}

// ListenOpts represents configurable options for listening.
type ListenOpts struct {
	Addr        string
	Unix        string
	TLS         string
	TLSCertFile string
	TLSKeyFile  string
	TLSCaFile   string
	DataAPIAddr string
}

// ProxyOpts represents configurable options for proxy.
type ProxyOpts struct {
	Addr        string
	TLSCertFile string
	TLSKeyFile  string
	TLSCaFile   string
}

// LogOpts represents configurable options for logging.
type LogOpts struct {
	Level  string
	Format string
	UUID   bool
}

// OTelOpts represents configurable options for OpenTelemetry.
type OTelOpts struct {
	Traces OTelTracesOpts
}

// OTelTracesOpts represents configurable options for OpenTelemetry traces.
type OTelTracesOpts struct {
	URL string
}

// DevOpts represents configurable options for development.
type DevOpts struct {
	ReplSetName string
	RecordsDir  string
	Telemetry   TelemetryOpts
}

// TelemetryOpts represents configurable options for telemetry.
type TelemetryOpts struct {
	URL            string
	Package        string
	UndecidedDelay time.Duration
	ReportInterval time.Duration
}

// DefaultLogLevel returns the default log level.
func DefaultLogLevel() slog.Level {
	if version.Get().DevBuild {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

// setupState setups state provider.
func setupState(stateDir string) *state.Provider {
	if stateDir == "" || stateDir == "-" {
		log.Fatal("State directory must be set.")
	}

	f, err := filepath.Abs(filepath.Join(stateDir, "state.json"))
	if err != nil {
		log.Fatalf("Failed to get path for state file: %s.", err)
	}

	sp, err := state.NewProvider(f)
	if err != nil {
		log.Fatal(stateFileProblem(f, err))
	}

	return sp
}

// setupMetrics setups Prometheus metrics registerer with some metrics.
func setupMetrics(uuid bool, stateProvider *state.Provider) prometheus.Registerer {
	r := prometheus.DefaultRegisterer
	m := stateProvider.MetricsCollector(true)

	// we don't do it by default due to
	// https://prometheus.io/docs/instrumenting/writing_exporters/#target-labels-not-static-scraped-labels
	if uuid {
		r = prometheus.WrapRegistererWith(
			prometheus.Labels{"uuid": stateProvider.Get().UUID},
			prometheus.DefaultRegisterer,
		)
		m = stateProvider.MetricsCollector(false)
	}

	r.MustRegister(m)

	return r
}

// setupLogger setups slog logger.
func setupLogger(logLevel string, format string, uuid string) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		log.Fatal(err)
	}

	opts := &logging.NewHandlerOpts{
		Base:          format,
		Level:         level,
		CheckMessages: false, // TODO https://github.com/FerretDB/FerretDB/issues/4511
	}
	logging.Setup(opts, uuid)
	logger := slog.Default()

	return logger
}

// checkOptions checks and logs options set for run.
func checkOptions(ctx context.Context, logger *slog.Logger, opts *RunOpts) {
	if devbuild.Enabled {
		logger.WarnContext(ctx, "This is a development build. The performance will be affected.")
	}

	if logger.Enabled(ctx, slog.LevelDebug) {
		logger.WarnContext(ctx, "Debug logging enabled. The performance will be affected.")
	}

	if !opts.Auth {
		logger.WarnContext(ctx, "Authentication is disabled. The server will accept any connection.")
	}
}

// dumpMetrics dumps all Prometheus metrics to stderr.
func dumpMetrics() {
	mfs := must.NotFail(prometheus.DefaultGatherer.Gather())

	for _, mf := range mfs {
		must.NotFail(expfmt.MetricFamilyToText(os.Stderr, mf))
	}
}

// Run sets up environment based on provided RunOpts and runs FerretDB.
func Run(ctx context.Context, opts *RunOpts) {
	// to increase a chance of resource finalizers to spot problems
	if devbuild.Enabled {
		defer func() {
			runtime.GC()
			runtime.GC()
		}()
	}

	info := version.Get()

	if p := opts.Dev.Telemetry.Package; p != "" {
		info.Package = p
	}

	if opts.Version {
		_, _ = fmt.Fprintln(os.Stdout, "version:", info.Version)
		_, _ = fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		_, _ = fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		_, _ = fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)
		_, _ = fmt.Fprintln(os.Stdout, "package:", info.Package)
		_, _ = fmt.Fprintln(os.Stdout, "devBuild:", info.DevBuild)

		return
	}

	// safe to always enable
	runtime.SetBlockProfileRate(10000)

	stateProvider := setupState(opts.StateDir)

	metricsRegisterer := setupMetrics(opts.MetricsUUID, stateProvider)

	startupFields := []slog.Attr{
		slog.String("version", info.Version),
		slog.String("commit", info.Commit),
		slog.String("branch", info.Branch),
		slog.Bool("dirty", info.Dirty),
		slog.String("package", info.Package),
		slog.Bool("devBuild", info.DevBuild),
		slog.Any("buildEnvironment", info.BuildEnvironment),
	}
	logUUID := stateProvider.Get().UUID

	// Similarly to Prometheus, unless requested, don't add UUID to all messages, but log it once at startup.
	if !opts.Log.UUID {
		startupFields = append(startupFields, slog.String("uuid", logUUID))
		logUUID = ""
	}

	logger := setupLogger(opts.Log.Level, opts.Log.Format, logUUID)

	logger.LogAttrs(ctx, slog.LevelInfo, "Starting FerretDB "+info.Version+"...", startupFields...)

	checkOptions(ctx, logger, opts)

	maxprocsOpts := []maxprocs.Option{
		maxprocs.Min(2),
		maxprocs.RoundQuotaFunc(func(v float64) int {
			return int(math.Ceil(v))
		}),
		maxprocs.Logger(func(format string, a ...any) {
			logger.Info(fmt.Sprintf(format, a...))
		}),
	}
	if _, err := maxprocs.Set(maxprocsOpts...); err != nil {
		logger.WarnContext(ctx, "Failed to set GOMAXPROCS", logging.Error(err))
	}

	ctx, stop := ctxutil.SigTerm(ctx)

	go func() {
		<-ctx.Done()
		logger.InfoContext(ctx, "Stopping")

		// second SIGTERM should immediately stop the process
		stop()
	}()

	// used to start debug handler with probes as soon as possible, even before listener is created
	var listener atomic.Pointer[clientconn.Listener]

	var wg sync.WaitGroup

	if addr := opts.DebugAddr; addr != "" && addr != "-" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			l := logging.WithName(logger, "debug")
			ready := ReadyZ{
				l:    l,
				opts: opts,
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

	if u := opts.OTel.Traces.URL; u != "" && u != "-" {
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

	{
		wg.Add(1)

		go func() {
			defer wg.Done()

			l := logging.WithName(logger, "telemetry")

			file, err := filepath.Abs(filepath.Join(opts.StateDir, "telemetry.json"))
			if err != nil {
				l.LogAttrs(ctx, logging.LevelFatal, "Failed to get path for local telemetry report file", logging.Error(err))
			}

			r, err := telemetry.NewReporter(&telemetry.NewReporterOpts{
				URL:            opts.Dev.Telemetry.URL,
				File:           file,
				F:              &opts.Telemetry,
				DNT:            os.Getenv("DO_NOT_TRACK"),
				ExecName:       os.Args[0],
				P:              stateProvider,
				ConnMetrics:    metrics.ConnMetrics,
				L:              l,
				UndecidedDelay: opts.Dev.Telemetry.UndecidedDelay,
				ReportInterval: opts.Dev.Telemetry.ReportInterval,
			})
			if err != nil {
				l.LogAttrs(ctx, logging.LevelFatal, "Failed to create telemetry reporter", logging.Error(err))
			}

			r.Run(ctx)
		}()
	}

	p, err := documentdb.NewPool(opts.PostgreSQLURL, logging.WithName(logger, "pool"), stateProvider)
	if err != nil {
		logger.LogAttrs(ctx, logging.LevelFatal, "Failed to construct pool", logging.Error(err))
	}

	defer p.Close()

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: opts.Auth,

		TCPHost:     opts.Listen.Addr,
		ReplSetName: opts.Dev.ReplSetName,

		L:             logging.WithName(logger, "handler"),
		ConnMetrics:   metrics.ConnMetrics,
		StateProvider: stateProvider,
	}

	h, err := handler.New(handlerOpts)
	if err != nil {
		handlerOpts.L.LogAttrs(ctx, logging.LevelFatal, "Failed to construct handler", logging.Error(err))
	}

	lis, err := clientconn.Listen(&clientconn.NewListenerOpts{
		TCP:  opts.Listen.Addr,
		Unix: opts.Listen.Unix,

		TLS:         opts.Listen.TLS,
		TLSCertFile: opts.Listen.TLSCertFile,
		TLSKeyFile:  opts.Listen.TLSKeyFile,
		TLSCAFile:   opts.Listen.TLSCaFile,

		ProxyAddr:        opts.Proxy.Addr,
		ProxyTLSCertFile: opts.Proxy.TLSCertFile,
		ProxyTLSKeyFile:  opts.Proxy.TLSKeyFile,
		ProxyTLSCAFile:   opts.Proxy.TLSCaFile,

		Mode:           clientconn.Mode(opts.Mode),
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: opts.Dev.RecordsDir,
	})
	if err != nil {
		logger.LogAttrs(ctx, logging.LevelFatal, "Failed to construct listener", logging.Error(err))
	}

	if addr := opts.Listen.DataAPIAddr; addr != "" && addr != "-" {
		wg.Add(1)

		go func() {
			defer wg.Done()

			l := logging.WithName(logger, "dataapi")

			var lis *dataapi.Listener

			lis, err = dataapi.Listen(&dataapi.ListenOpts{
				TCPAddr: addr,
				L:       l,
				Handler: h,
			})
			if err != nil {
				l.LogAttrs(ctx, logging.LevelFatal, "Failed to construct DataAPI listener", logging.Error(err))
			}

			lis.Run(ctx)
		}()
	}

	listener.Store(lis)

	metricsRegisterer.MustRegister(lis)

	lis.Run(ctx)

	wg.Wait()

	if info.DevBuild {
		dumpMetrics()
	}
}

// Ping sets up environment based on provided RunOpts and runs FerretDB.
func Ping(ctx context.Context, opts *RunOpts) bool {
	l := setupLogger(opts.Log.Level, opts.Log.Format, "")
	checkOptions(ctx, l, opts)

	ready := ReadyZ{
		l:    l,
		opts: opts,
	}

	ctx, stop := ctxutil.SigTerm(ctx)
	defer stop()

	return ready.Probe(ctx)
}

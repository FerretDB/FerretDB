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

// Package debug provides debug facilities.
package debug

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	_ "expvar" // for metrics
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"slices"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "debug"
)

// setup ensures that debug handler is set up only once.
var setup atomic.Bool

// Probe should return true on success and false on failure (or context cancellation).
// It may log additional information if needed.
//
// It must be thread-safe.
type Probe func(ctx context.Context) bool

// Handler represents debug handler.
//
//nolint:vet // for readability
type Handler struct {
	opts     *ListenOpts
	lis      net.Listener
	handlers map[string]string
	stdL     *log.Logger
}

// ListenOpts represents [Listen] options.
//
//nolint:vet // for readability
type ListenOpts struct {
	TCPAddr string
	L       *slog.Logger
	R       prometheus.Registerer
	Livez   Probe
	Readyz  Probe
}

// Listen creates a new debug handler and starts listener on the given TCP address.
//
// This function can be called only once because it affects [http.DefaultServeMux].
func Listen(opts *ListenOpts) (*Handler, error) {
	if setup.Swap(true) {
		panic("debug handler is already set up")
	}

	must.NotBeZero(opts)

	l := opts.L

	stdL := slog.NewLogLogger(l.Handler(), slog.LevelError)

	probeDurations := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "probe_response_seconds",
			Help:      "Probe response time seconds.",
			Buckets:   []float64{0.1, 0.5, 1, 5},
		},
		[]string{"probe", "code"},
	)

	opts.R.MustRegister(probeDurations)

	http.Handle("/debug/metrics", promhttp.InstrumentMetricHandler(
		opts.R, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          opts.R,
			EnableOpenMetrics: true,
		}),
	))

	http.HandleFunc("/debug/archive", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/zip")

		rw.Header().Set(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename=FerretDB-debug-%d.zip", time.Now().UnixMilli()),
		)

		ctx := req.Context()
		zipWriter := zip.NewWriter(rw)

		defer func() {
			if err := zipWriter.Close(); err != nil {
				l.ErrorContext(ctx, "Archive handler failed", logging.Error(err))
				return
			}
		}()

		metricsFile, err := zipWriter.Create("metrics")
		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed", logging.Error(err))
			return
		}

		debugAddr := ctx.Value(http.LocalAddrContextKey).(net.Addr)

		// we use *http.Request instead of http.Get function to provide the ctx
		scrapeReq := must.NotFail(http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://%s%s", debugAddr.String(), "/debug/metrics"),
			nil,
		))

		resp, err := http.DefaultClient.Do(scrapeReq)
		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed - metrics failed", logging.Error(err))
			return
		}

		_, err = io.Copy(metricsFile, resp.Body)

		_ = resp.Body.Close()

		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed", logging.Error(err))
			return
		}

		heapFile, err := zipWriter.Create("heap")
		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed", logging.Error(err))
			return
		}

		scrapeReq.URL.Path = "/debug/pprof/heap"

		resp, err = http.DefaultClient.Do(scrapeReq)
		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed - pprof failed", logging.Error(err))
			return
		}

		_, err = io.Copy(heapFile, resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			l.ErrorContext(ctx, "Archive handler failed", logging.Error(err))
			return
		}
	})

	svOpts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
		// TODO https://github.com/FerretDB/FerretDB/issues/3600
	}
	must.NoError(statsviz.Register(http.DefaultServeMux, svOpts...))

	http.Handle("/debug/livez", promhttp.InstrumentHandlerDuration(
		probeDurations.MustCurryWith(prometheus.Labels{"probe": "livez"}),
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			if !opts.Livez(ctx) {
				l.Warn("Livez probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			l.Debug("Livez probe succeeded")
			rw.WriteHeader(http.StatusOK)
		}),
	))

	http.Handle("/debug/readyz", promhttp.InstrumentHandlerDuration(
		probeDurations.MustCurryWith(prometheus.Labels{"probe": "readyz"}),
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			if !opts.Livez(ctx) {
				l.Warn("Readyz probe failed - livez probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			if !opts.Readyz(ctx) {
				l.Warn("Readyz probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			l.Debug("Readyz probe succeeded")
			rw.WriteHeader(http.StatusOK)
		}),
	))

	handlers := map[string]string{
		// custom handlers registered above
		"/debug/metrics": "Metrics in Prometheus format",
		"/debug/archive": "Metrics and pprof heap profile packed into the zip archive",
		"/debug/graphs":  "Visualize metrics",
		"/debug/livez":   "Liveness probe",
		"/debug/readyz":  "Readiness probe",

		// stdlib handlers
		"/debug/vars":  "Expvar package metrics",
		"/debug/pprof": "Runtime profiling data for pprof",
	}

	var page bytes.Buffer
	must.NoError(template.Must(template.New("debug").Parse(`
	<html>
	<body>
	<ul>
	{{range $path, $desc := .}}
		<li><a href="{{$path}}">{{$path}}</a>: {{$desc}}</li>
	{{end}}
	</ul>
	</body>
	</html>
	`)).Execute(&page, handlers))

	http.HandleFunc("/debug", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write(page.Bytes())
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, "/debug", http.StatusSeeOther)
	})

	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Handler{
		opts:     opts,
		lis:      lis,
		handlers: handlers,
		stdL:     stdL,
	}, nil
}

// Serve runs debug handler until ctx is canceled.
//
// It exits when handler is stopped and listener closed.
func (h *Handler) Serve(ctx context.Context) {
	s := http.Server{
		Addr:     h.opts.TCPAddr,
		Handler:  http.DefaultServeMux,
		ErrorLog: h.stdL,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	l := h.opts.L

	root := fmt.Sprintf("http://%s", h.lis.Addr())

	l.InfoContext(ctx, fmt.Sprintf("Starting debug server on %s...", root))

	paths := maps.Keys(h.handlers)
	slices.Sort(paths)

	for _, path := range paths {
		l.InfoContext(ctx, fmt.Sprintf("%s%s - %s", root, path, h.handlers[path]))
	}

	go func() {
		if err := s.Serve(h.lis); !errors.Is(err, http.ErrServerClosed) {
			l.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	stopCtx, stopCancel := ctxutil.WithDelay(ctx)
	defer stopCancel(nil)

	if err := s.Shutdown(stopCtx); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Shutdown exited with unexpected error", logging.Error(err))
	}

	if err := s.Close(); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Close exited with unexpected error", logging.Error(err))
	}

	l.InfoContext(ctx, "Debug server stopped")
}

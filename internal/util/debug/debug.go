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
	"maps"
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

	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
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

// addToZip adds a new file to the zip archive.
//
// Passed [io.ReadCloser] is always closed.
func addToZip(w *zip.Writer, name string, r io.ReadCloser) (err error) {
	defer func() {
		if e := r.Close(); e != nil && err == nil {
			err = e
		}
	}()

	f, err := w.CreateHeader(&zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	})
	if err != nil {
		return
	}

	_, err = io.Copy(f, r)

	return
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
			fmt.Sprintf("attachment; filename=ferretdb-%s.zip", time.Now().Format("2006-01-02-15-04-05")),
		)

		ctx := req.Context()
		zipWriter := zip.NewWriter(rw)
		errs := map[string]error{}

		defer func() {
			files := slices.Sorted(maps.Keys(errs))

			var b bytes.Buffer
			for _, f := range files {
				b.WriteString(fmt.Sprintf("%s: %v\n", f, errs[f]))
			}

			if err := addToZip(zipWriter, "errors.txt", io.NopCloser(&b)); err != nil {
				l.ErrorContext(ctx, "Failed to add errors.txt to archive", logging.Error(err))
			}

			if err := zipWriter.Close(); err != nil {
				l.ErrorContext(ctx, "Failed to close archive", logging.Error(err))
			}
		}()

		host := ctx.Value(http.LocalAddrContextKey).(net.Addr)
		getReq := must.NotFail(http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/", host), nil))

		for file, u := range map[string]struct {
			path  string
			query string
		}{
			"goroutine.pprof": {path: "/debug/pprof/goroutine"},
			"heap.pprof":      {path: "/debug/pprof/heap", query: "gc=1"},
			"profile.pprof":   {path: "/debug/pprof/profile", query: "seconds=5"},
			"metrics.txt":     {path: "/debug/metrics"}, // includes version, UUID, etc
			"vars.json":       {path: "/debug/vars"},
		} {
			getReq.URL.Path = u.path
			getReq.URL.RawQuery = u.query

			l.DebugContext(ctx, "Fetching file for archive", slog.String("file", file), slog.String("url", getReq.URL.String()))

			resp, err := http.DefaultClient.Do(getReq)
			if err == nil {
				err = addToZip(zipWriter, file, resp.Body)
			}

			errs[file] = err
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
			ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
			defer cancel()

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
			ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
			defer cancel()

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
		"/debug/archive": "Zip archive with debugging information",
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

	l.InfoContext(ctx, fmt.Sprintf("Starting debug server on %s/debug ...", root))

	paths := slices.Sorted(maps.Keys(h.handlers))

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
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := s.Shutdown(shutdownCtx); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Shutdown exited with unexpected error", logging.Error(err))
	}

	if err := s.Close(); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Close exited with unexpected error", logging.Error(err))
	}

	l.InfoContext(ctx, "Debug server stopped")
}

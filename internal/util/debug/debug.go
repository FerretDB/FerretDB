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
	"bytes"
	"context"
	"errors"
	_ "expvar" // for metrics
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"slices"
	"sync/atomic"
	"text/template"

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "debug"
)

// setup ensures that debug handler is set up only once.
var setup atomic.Bool

// Probe should return true on success and false on failure.
//
// It must be thread-safe.
type Probe func() bool

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
	L       *zap.Logger
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

	stdL := must.NotFail(zap.NewStdLogAt(opts.L, zap.WarnLevel))

	http.Handle("/debug/metrics", promhttp.InstrumentMetricHandler(
		opts.R, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          opts.R,
			EnableOpenMetrics: true,
		}),
	))

	svOpts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
		// TODO https://github.com/FerretDB/FerretDB/issues/3600
	}
	must.NoError(statsviz.Register(http.DefaultServeMux, svOpts...))

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

	must.NoError(opts.R.Register(probeDurations))

	http.Handle("/debug/livez", promhttp.InstrumentHandlerDuration(
		probeDurations.MustCurryWith(prometheus.Labels{"probe": "livez"}),
		http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			if !opts.Livez() {
				opts.L.Warn("Livez probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			opts.L.Debug("Livez probe succeeded")
			rw.WriteHeader(http.StatusOK)
		}),
	))

	http.Handle("/debug/readyz", promhttp.InstrumentHandlerDuration(
		probeDurations.MustCurryWith(prometheus.Labels{"probe": "readyz"}),
		http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			if !opts.Livez() {
				opts.L.Warn("Readyz probe failed - livez probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			if !opts.Readyz() {
				opts.L.Warn("Readyz probe failed")
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			opts.L.Debug("Readyz probe succeeded")
			rw.WriteHeader(http.StatusOK)
		}),
	))

	handlers := map[string]string{
		// custom handlers registered above
		"/debug/metrics": "Metrics in Prometheus format",
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

	root := fmt.Sprintf("http://%s", h.lis.Addr())

	h.opts.L.Sugar().Infof("Starting debug server on %s ...", root)

	paths := maps.Keys(h.handlers)
	slices.Sort(paths)

	for _, path := range paths {
		h.opts.L.Sugar().Infof("%s%s - %s", root, path, h.handlers[path])
	}

	go func() {
		if err := s.Serve(h.lis); !errors.Is(err, http.ErrServerClosed) {
			h.opts.L.DPanic("Serve exited with unexpected error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	stopCtx, stopCancel := ctxutil.WithDelay(ctx)
	defer stopCancel(nil)

	if err := s.Shutdown(stopCtx); err != nil {
		h.opts.L.DPanic("Shutdown exited with unexpected error", zap.Error(err))
	}

	if err := s.Close(); err != nil {
		h.opts.L.DPanic("Close exited with unexpected error", zap.Error(err))
	}

	h.opts.L.Sugar().Info("Debug server stopped.")
}

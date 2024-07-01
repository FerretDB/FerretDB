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

var setup atomic.Bool

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

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of debug handler requests.",
		},
		[]string{"handler", "code"},
	)

	must.NoError(opts.R.Register(requestCount))

	startedHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		// TODO https://github.com/FerretDB/FerretDB/issues/4306
		rw.WriteHeader(http.StatusOK)
	})

	// healthz handler, which is used for liveness probe, returns StatusOK when reached.
	// This ensures that listener is running.
	healthzHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	readyHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		// TODO https://github.com/FerretDB/FerretDB/issues/4306
		rw.WriteHeader(http.StatusOK)
	})

	http.Handle("/debug/started", promhttp.InstrumentHandlerCounter(
		requestCount.MustCurryWith(prometheus.Labels{"handler": "/debug/started"}),
		startedHandler,
	))

	http.Handle("/debug/healthz", promhttp.InstrumentHandlerCounter(
		requestCount.MustCurryWith(prometheus.Labels{"handler": "/debug/healthz"}),
		healthzHandler,
	))

	http.Handle("/debug/ready", promhttp.InstrumentHandlerCounter(
		requestCount.MustCurryWith(prometheus.Labels{"handler": "/debug/ready"}),
		readyHandler,
	))

	handlers := map[string]string{
		// custom handlers registered above
		"/debug/graphs":  "Visualize metrics",
		"/debug/metrics": "Metrics in Prometheus format",

		// custom handlers for Kubernetes probes
		"/debug/started": "Check if listener have started",
		"/debug/healthz": "Check if listener is healthy",
		"/debug/ready":   "Check if listener and backend are ready for queries",

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
			panic(err)
		}
	}()

	<-ctx.Done()

	stopCtx, stopCancel := ctxutil.WithDelay(ctx)
	defer stopCancel(nil)

	_ = s.Shutdown(stopCtx)
	_ = s.Close()

	h.opts.L.Sugar().Info("Debug server stopped.")
}

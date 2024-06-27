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
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"slices"
	"text/template"
	"time"

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "debug"
)

// RunHandler runs debug handler.
func RunHandler(ctx context.Context, addr string, r prometheus.Registerer, l *zap.Logger, started <-chan struct{}) {
	stdL := must.NotFail(zap.NewStdLogAt(l, zap.WarnLevel))

	http.Handle("/debug/metrics", promhttp.InstrumentMetricHandler(
		r, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          r,
			EnableOpenMetrics: true,
		}),
	))

	opts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
		// TODO https://github.com/FerretDB/FerretDB/issues/3600
	}
	must.NoError(statsviz.Register(http.DefaultServeMux, opts...))

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of debug handler requests.",
		},
		[]string{"handler", "code"},
	)

	must.NoError(r.Register(requestCount))

	// started handler, which is used for startup probe, returns StatusOK when FerretDB listener were initialized.
	// If it wasn't yet, the StatusInternalServerError is returned.
	//
	// If started channel is nil, the handler always returns StatusOK.
	startedHandler := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if started == nil {
			rw.WriteHeader(http.StatusOK)
		}

		select {
		case <-started:
			rw.WriteHeader(http.StatusOK)
		default:
			rw.WriteHeader(http.StatusInternalServerError)
		}
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

	s := http.Server{
		Addr:     addr,
		ErrorLog: stdL,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		lis := must.NotFail(net.Listen("tcp", addr))

		root := fmt.Sprintf("http://%s", lis.Addr())

		l.Sugar().Infof("Starting debug server on %s ...", root)

		paths := maps.Keys(handlers)
		slices.Sort(paths)

		for _, path := range paths {
			l.Sugar().Infof("%s%s - %s", root, path, handlers[path])
		}

		if err := s.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	s.Shutdown(stopCtx) //nolint:contextcheck // use new context for cancellation

	s.Close()
	l.Sugar().Info("Debug server stopped.")
}

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

const (
	archivePath = "/debug/archive"
	graphsPath  = "/debug/graphs"
	metricsPath = "/debug/metrics"
	pprofPath   = "/debug/pprof"
	varsPath    = "/debug/vars"
)

// RunHandler runs debug handler.
func RunHandler(ctx context.Context, addr string, r prometheus.Registerer, l *zap.Logger) {
	stdL := must.NotFail(zap.NewStdLogAt(l, zap.WarnLevel))

	s := http.Server{
		Addr:     addr,
		ErrorLog: stdL,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler: debugHandler(ctx, r, l),
	}

	go func() {
		lis := must.NotFail(net.Listen("tcp", addr))

		root := fmt.Sprintf("http://%s", lis.Addr())

		l.Sugar().Infof("Starting debug server on %s ...", root)

		handlers := handlersList()

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

// debugHandler returns the main handler for debugging endpoints.
func debugHandler(ctx context.Context, r prometheus.Registerer, l *zap.Logger) http.Handler {
	stdL := must.NotFail(zap.NewStdLogAt(l, zap.WarnLevel))

	http.Handle(metricsPath, promhttp.InstrumentMetricHandler(
		r, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          r,
			EnableOpenMetrics: true,
		}),
	))

	opts := []statsviz.Option{
		statsviz.Root(graphsPath),
		// TODO https://github.com/FerretDB/FerretDB/issues/3600
	}

	must.NoError(statsviz.Register(http.DefaultServeMux, opts...))

	http.HandleFunc(archivePath, archiveHandler)

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
	`)).Execute(&page, handlersList()))

	http.HandleFunc("/debug", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write(page.Bytes())
	})

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, "/debug", http.StatusSeeOther)
	})

	return http.DefaultServeMux
}

// handlersList returns the map of handler paths and their descriptions.
func handlersList() map[string]string {
	return map[string]string{
		// custom handlers registered above
		graphsPath:  "Visualize metrics",
		metricsPath: "Metrics in Prometheus format",
		archivePath: "Metrics and pprof data in zip format",

		// stdlib handlers
		varsPath:  "Expvar package metrics",
		pprofPath: "Runtime profiling data for pprof",
	}
}

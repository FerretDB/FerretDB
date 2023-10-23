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
	_ "expvar" // for metrics
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"text/template"
	"time"

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// RunHandler runs debug handler.
func RunHandler(ctx context.Context, addr string, r prometheus.Registerer, l *zap.Logger) {
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

	handlers := map[string]string{
		// custom handlers registered above
		"/debug/graphs":  "Visualize metrics",
		"/debug/metrics": "Metrics in Prometheus format",

		// stdlib handlers
		"/debug/vars":  "Expvar package metrics",
		"/debug/pprof": "Runtime profiling data for pprof",
	}

	var page bytes.Buffer
	must.NoError(template.Must(template.New("debug").Parse(`
	<html>
	<body>
	<ul>
	{{range $key, $value := .}}
		<li><a href="{{$key}}">{{$key}}</a>: {{$value}}</li>
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

		l.Sugar().Infof("Starting debug server on http://%s/", lis.Addr())

		if err := s.Serve(lis); err != http.ErrServerClosed {
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

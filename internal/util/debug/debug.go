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
	"log"
	"math"
	"math/rand"
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
		statsviz.TimeseriesPlot(scatterPlot()),
		statsviz.TimeseriesPlot(barPlot()),
		statsviz.TimeseriesPlot(stackedPlot()),
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

func scatterPlot() statsviz.TimeSeriesPlot {
	// Describe the 'sine' time series.
	sine := statsviz.TimeSeries{
		Name:     "short sin",
		Unitfmt:  "%{y:.4s}B",
		GetValue: updateSine,
	}

	// Build a new plot, showing our sine time series
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:  "sine",
		Title: "Sine",
		Type:  statsviz.Scatter,
		InfoText: `This is an example of a 'scatter' type plot, showing a single time series.<br>
InfoText field (this) accepts any HTML tags like <b>bold</b>, <i>italic</i>, etc.`,
		YAxisTitle: "y unit",
		Series:     []statsviz.TimeSeries{sine},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}

	return plot
}

func barPlot() statsviz.TimeSeriesPlot {
	// Describe the 'user logins' time series.
	logins := statsviz.TimeSeries{
		Name:     "user logins",
		Unitfmt:  "%{y:.4s}",
		GetValue: logins,
	}

	// Describe the 'user signins' time series.
	signins := statsviz.TimeSeries{
		Name:     "user signins",
		Unitfmt:  "%{y:.4s}",
		GetValue: signins,
	}

	// Build a new plot, showing both time series at once.
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:  "users",
		Title: "Users",
		Type:  statsviz.Bar,
		InfoText: `This is an example of a 'bar' type plot, showing 2 time series.<br>
InfoText field (this) accepts any HTML tags like <b>bold</b>, <i>italic</i>, etc.`,
		YAxisTitle: "users",
		Series:     []statsviz.TimeSeries{logins, signins},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}

	return plot
}

func stackedPlot() statsviz.TimeSeriesPlot {
	// Describe the 'user logins' time series.
	logins := statsviz.TimeSeries{
		Name:     "user logins",
		Unitfmt:  "%{y:.4s}",
		Type:     statsviz.Bar,
		GetValue: logins,
	}

	// Describe the 'user signins' time series.
	signins := statsviz.TimeSeries{
		Name:     "user signins",
		Unitfmt:  "%{y:.4s}",
		Type:     statsviz.Bar,
		GetValue: signins,
	}

	// Build a new plot, showing both time series at once.
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:    "users-stack",
		Title:   "Stacked Users",
		Type:    statsviz.Bar,
		BarMode: statsviz.Stack,
		InfoText: `This is an example of a 'bar' plot showing 2 time series stacked on top of each other with <b>BarMode:Stack</b>.<br>
InfoText field (this) accepts any HTML tags like <b>bold</b>, <i>italic</i>, etc.`,
		YAxisTitle: "users",
		Series:     []statsviz.TimeSeries{logins, signins},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}

	return plot
}

var val = 0.

func updateSine() float64 {
	val += 0.5
	return math.Sin(val)
}

func logins() float64 {
	return (rand.Float64() + 2) * 1000
}

func signins() float64 {
	return (rand.Float64() + 1.5) * 100
}

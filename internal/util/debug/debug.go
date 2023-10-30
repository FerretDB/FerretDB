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
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof" // for profiling
	"text/template"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

// RunHandler runs debug handler.
func RunHandler(ctx context.Context, addr string, r prometheus.Registerer, l *zap.Logger) {
	stdL := must.NotFail(zap.NewStdLogAt(l, zap.WarnLevel))

	metricHandler := promhttp.InstrumentMetricHandler(
		r, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:          stdL,
			ErrorHandling:     promhttp.ContinueOnError,
			Registry:          r,
			EnableOpenMetrics: true,
		}),
	)
	prometheus_metrics, err := prometheus.DefaultGatherer.Gather()
	fmt.Printf("prometheus_metrics: %v\n", prometheus_metrics)
	if err != nil {
		log.Fatalf("error in Gathering prometheus metrics: %v", err)
	}

	http.Handle("/debug/metrics", metricHandler)

	opts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
		statsviz.TimeseriesPlot(ferretdb_postgresql_metadata_databases_barPlot()),
		statsviz.TimeseriesPlot(ferretdb_postgresql_pool_size_barPlot()),
		statsviz.TimeseriesPlot(promhttp_metric_handler_requests_total_stackPlot()),
		statsviz.TimeseriesPlot(promhttp_metric_handler_requests_in_flight_scatterPlot()),
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

// gauge
func promhttp_metric_handler_requests_in_flight_scatterPlot() statsviz.TimeSeriesPlot {
	scrapes := statsviz.TimeSeries{
		Name:     "prometheus http scrape",
		Unitfmt:  "%{y:.4s}B",
		GetValue: flightReqgaugeMetric,
	}

	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "scrapes",
		Title:      "prometheus http request Count",
		Type:       statsviz.Scatter,
		InfoText:   "Helps Visualizing Current number of scrapes being served.",
		YAxisTitle: "Scrapes",
		Series:     []statsviz.TimeSeries{scrapes},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

// gauge
func ferretdb_postgresql_metadata_databases_barPlot() statsviz.TimeSeriesPlot {
	databaseCount := statsviz.TimeSeries{
		Name:     "Database count",
		Unitfmt:  "%{y:.4s}",
		GetValue: metadataDbgaugeMetric,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "dbCount",
		Title:      "Postgresql MetaData Database Count",
		Type:       statsviz.Bar,
		YAxisTitle: "Database Count",
		Series:     []statsviz.TimeSeries{databaseCount},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

// gauge
func ferretdb_postgresql_pool_size_barPlot() statsviz.TimeSeriesPlot {
	poolSize := statsviz.TimeSeries{
		Name:     "Postgresql Pool size",
		Unitfmt:  "%{y:.4s}",
		GetValue: poolSizegaugeMetric,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "poolSize",
		Title:      "postgresql pool size",
		Type:       statsviz.Bar,
		YAxisTitle: "Pool Size",
		Series:     []statsviz.TimeSeries{poolSize},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

// counter metric
func promhttp_metric_handler_requests_total_stackPlot() statsviz.TimeSeriesPlot {

	code200 := statsviz.TimeSeries{
		Name:     "Code 200",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code503 := statsviz.TimeSeries{
		Name:     "Code 503",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code500 := statsviz.TimeSeries{
		Name:     "Code 500",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code201 := statsviz.TimeSeries{
		Name:     "Code 201",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code202 := statsviz.TimeSeries{
		Name:     "Code 202",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code203 := statsviz.TimeSeries{
		Name:     "Code 203",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	code204 := statsviz.TimeSeries{
		Name:     "Code 204",
		Unitfmt:  "%{y:.4s}B",
		Type:     statsviz.Bar,
		GetValue: codeCountGen,
	}

	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "prometheus metric handler request count",
		Title:      "Prometheus metric Count(Status Code)",
		Type:       statsviz.Bar,
		BarMode:    statsviz.Stack,
		YAxisTitle: "status codes",
		Series:     []statsviz.TimeSeries{code200, code201, code202, code203, code204, code500, code503},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}

	return plot
}

func codeCountGen() float64 {
	return rand.Float64()
}

func metricRetriever(prometheus_metrics []*io_prometheus_client.MetricFamily, metricName string) *io_prometheus_client.Metric {
	for _, specificMetric := range prometheus_metrics {
		if specificMetric.GetName() == metricName {
			finalMetricSlice := specificMetric.GetMetric()
			for _, x := range finalMetricSlice {
				return x
			}
		}
	}
	return nil
}

func prometheusGather() []*io_prometheus_client.MetricFamily {
	prometheus_metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Fatalf("error in Gathering prometheus metrics: %v", err)
	}
	return prometheus_metrics
}

func poolSizegaugeMetric() float64 {
	str := "ferretdb_postgresql_pool_size"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func metadataDbgaugeMetric() float64 {
	str := "ferretdb_postgresql_metadata_databases"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func flightReqgaugeMetric() float64 {
	str := "promhttp_metric_handler_requests_in_flight"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

// func counterMetric(value string) float64 {
// 	str := "promhttp_metric_handler_requests_total"
// 	p := prometheusGather()
// 	m := metricRetriever(p, str)
// 	x := "code"
// 	for i, label := range m.Label {
// 		if label.Name == &x && label.Value == &value {
// 			return *m.Counter[i].Value
// 		}
// 	}
// 	return *m.Counter.Value
// }

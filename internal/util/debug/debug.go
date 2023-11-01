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
		statsviz.TimeseriesPlot(go_goroutines_scatterPlot()),
		statsviz.TimeseriesPlot(go_memstats_alloc_bytes_barPlot()),
		statsviz.TimeseriesPlot(go_memstats_frees_total_barplot()),
		statsviz.TimeseriesPlot(go_memstats_heap_alloc_bytes_collection()),
		statsviz.TimeseriesPlot(go_memstats_heap_objects()),
		statsviz.TimeseriesPlot(go_memstats_last_gc_time_seconds_scatterPlot()),
		statsviz.TimeseriesPlot(go_memstats_lookups_total_barPlot()),
		statsviz.TimeseriesPlot(go_memstats_mallocs_total_scatterPlot()),
		statsviz.TimeseriesPlot(go_memstats_mcache_bytes_barPlot()),
		statsviz.TimeseriesPlot(go_memstats_mspan_bytes_barPlot()),
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
func go_goroutines_scatterPlot() statsviz.TimeSeriesPlot {
	goroutines := statsviz.TimeSeries{
		Name:     "current goroutines",
		Unitfmt:  "%{y:.4s}B",
		GetValue: getGoroutines,
	}

	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Running Goroutines",
		Title:      "Running Goroutines",
		Type:       statsviz.Scatter,
		InfoText:   "Number of goroutines that currently exist.",
		YAxisTitle: "Goroutines",
		Series:     []statsviz.TimeSeries{goroutines},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot : %v", err)
	}
	return plot
}

//gauge

func go_memstats_alloc_bytes_barPlot() statsviz.TimeSeriesPlot {
	bytesAllocated := statsviz.TimeSeries{
		Name:     "Bytes Allocated and still in use",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_byte_alloc,
	}

	TotalbytesAllocated := statsviz.TimeSeries{
		Name:     "Total Bytes Allocated Even if freed",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_byte_alloc_total,
	}

	BytesAllocBuckHash := statsviz.TimeSeries{
		Name:     "Bytes Used by profiling bucket hash table",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_byte_buck_hash,
	}

	bytesGcSysMetadata := statsviz.TimeSeries{
		Name:     "bytes used for garbage collection system metadata.",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_gc_sys_bytes,
	}

	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "bytes allocated",
		Title:      "Bytes Allocated",
		Type:       statsviz.Bar,
		BarMode:    statsviz.Group,
		YAxisTitle: "Bytes",
		Series:     []statsviz.TimeSeries{bytesAllocated, TotalbytesAllocated, BytesAllocBuckHash, bytesGcSysMetadata},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

func go_memstats_frees_total_barplot() statsviz.TimeSeriesPlot {
	frees := statsviz.TimeSeries{
		Name:     "Number of frees",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_frees_total,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Frees",
		Title:      "Total Number of frees",
		Type:       statsviz.Bar,
		YAxisTitle: "frees",
		Series:     []statsviz.TimeSeries{frees},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

func go_memstats_heap_alloc_bytes_collection() statsviz.TimeSeriesPlot {
	heap_bytesStillUse := statsviz.TimeSeries{
		Name:     "heap bytes allocated and still in use",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_heap_still_use,
	}
	heap_bytesWaiting := statsviz.TimeSeries{
		Name:     "heap bytes waiting to be used",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_heap_waiting,
	}
	heap_bytesInUse := statsviz.TimeSeries{
		Name:     "heap bytes in use",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_heap_in_use,
	}
	heap_bytesReleased := statsviz.TimeSeries{
		Name:     "heap bytes released to OS",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_heap_bytes_released,
	}
	heap_SysBytes := statsviz.TimeSeries{
		Name:     "heap bytes obtained from system",
		Unitfmt:  "%{y:.4s}",
		GetValue: memstats_heap_sys,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Heap Bytes Collection",
		Title:      "Heap Bytes Stats",
		Type:       statsviz.Bar,
		BarMode:    statsviz.Group,
		YAxisTitle: "heap bytes",
		Series:     []statsviz.TimeSeries{heap_bytesStillUse, heap_bytesWaiting, heap_bytesInUse, heap_bytesReleased, heap_SysBytes},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

func go_memstats_heap_objects() statsviz.TimeSeriesPlot {
	objects := statsviz.TimeSeries{
		Name:     "Allocated Objects",
		Unitfmt:  "%{y:.4s}",
		GetValue: GetMemstatsHeapObjects,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Allocated Objects",
		Title:      "Number of Allocated Objects",
		Type:       statsviz.Scatter,
		YAxisTitle: "objects",
		Series:     []statsviz.TimeSeries{objects},
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

	// code201 := statsviz.TimeSeries{
	// 	Name:     "Code 201",
	// 	Unitfmt:  "%{y:.4s}B",
	// 	Type:     statsviz.Bar,
	// 	GetValue: codeCountGen,
	// }

	// code202 := statsviz.TimeSeries{
	// 	Name:     "Code 202",
	// 	Unitfmt:  "%{y:.4s}B",
	// 	Type:     statsviz.Bar,
	// 	GetValue: codeCountGen,
	// }

	// code203 := statsviz.TimeSeries{
	// 	Name:     "Code 203",
	// 	Unitfmt:  "%{y:.4s}B",
	// 	Type:     statsviz.Bar,
	// 	GetValue: codeCountGen,
	// }

	// code204 := statsviz.TimeSeries{
	// 	Name:     "Code 204",
	// 	Unitfmt:  "%{y:.4s}B",
	// 	Type:     statsviz.Bar,
	// 	GetValue: codeCountGen,
	// }

	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "prometheus metric handler request count",
		Title:      "Prometheus metric Count(Status Code)",
		Type:       statsviz.Bar,
		BarMode:    statsviz.Stack,
		YAxisTitle: "status codes",
		Series:     []statsviz.TimeSeries{code200, code500, code503},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}

	return plot
}

func go_memstats_last_gc_time_seconds_scatterPlot() statsviz.TimeSeriesPlot {
	seconds := statsviz.TimeSeries{
		Name:     "seconds since last GC",
		Unitfmt:  "%{y:.4s}",
		GetValue: last_gc_time_seconds,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "seconds since last GC",
		Title:      "Seconds since last GC",
		Type:       statsviz.Scatter,
		YAxisTitle: "Seconds",
		Series:     []statsviz.TimeSeries{seconds},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot: %v", err)
	}
	return plot
}

func go_memstats_lookups_total_barPlot() statsviz.TimeSeriesPlot {
	ptrLookUps := statsviz.TimeSeries{
		Name:     "Number of pointer lookups",
		Unitfmt:  "%{y:.4s}",
		GetValue: ptr_lookups,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Number of pointer lookups",
		Title:      "Pointer Lookups",
		Type:       statsviz.Bar,
		YAxisTitle: "lookups",
		Series:     []statsviz.TimeSeries{ptrLookUps},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot : %v", err)
	}
	return plot
}

func go_memstats_mallocs_total_scatterPlot() statsviz.TimeSeriesPlot {
	mallocs := statsviz.TimeSeries{
		Name:     "Mallocs",
		Unitfmt:  "%{y:.4s}",
		GetValue: mallocs_total,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "Mallocs",
		Title:      "Total number of mallocs",
		Type:       statsviz.Scatter,
		YAxisTitle: "mallocs",
		Series:     []statsviz.TimeSeries{mallocs},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot :%v", err)
	}
	return plot
}

func go_memstats_mcache_bytes_barPlot() statsviz.TimeSeriesPlot {
	bytes := statsviz.TimeSeries{
		Name:     "Bytes in use by mcache struct",
		Unitfmt:  "%{y:.4s}",
		GetValue: mcache_inuse_bytes,
	}
	Sys_bytes := statsviz.TimeSeries{
		Name:     "Bytes in use by mcache struct obtained from system",
		Unitfmt:  "%{y:.4s}",
		GetValue: mcache_inuse_bytes_System,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "bytes being used by mcache",
		Title:      "Bytes in use by mcache structures",
		Type:       statsviz.Bar,
		YAxisTitle: "bytes",
		Series:     []statsviz.TimeSeries{bytes, Sys_bytes},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot :%v", err)
	}
	return plot
}

func go_memstats_mspan_bytes_barPlot() statsviz.TimeSeriesPlot {
	bytes := statsviz.TimeSeries{
		Name:     "Bytes in use by mspan struct",
		Unitfmt:  "%{y:.4s}",
		GetValue: mspan_inuse_bytes,
	}
	Sys_bytes := statsviz.TimeSeries{
		Name:     "Bytes in use by mspan struct obtained from system",
		Unitfmt:  "%{y:.4s}",
		GetValue: mspan_inuse_bytes_System,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       "bytes being used by mspan",
		Title:      "Bytes in use by mspan structures",
		Type:       statsviz.Bar,
		YAxisTitle: "bytes",
		Series:     []statsviz.TimeSeries{bytes, Sys_bytes},
	}.Build()
	if err != nil {
		log.Fatalf("failed to build timeseries plot :%v", err)
	}
	return plot
}

//GetValueFunctions Beyond this point

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

func counterMetric(value string) float64 {
	str := "promhttp_metric_handler_requests_total"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Counter.Value
}

func getGoroutines() float64 {
	str := "go_goroutines"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_byte_alloc() float64 {
	str := "go_memstats_alloc_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_byte_alloc_total() float64 {
	str := "go_memstats_alloc_bytes_total"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Counter.Value
}

func memstats_byte_buck_hash() float64 {
	str := "go_memstats_buck_hash_sys_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_gc_sys_bytes() float64 {
	str := "go_memstats_gc_sys_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_frees_total() float64 {
	str := "go_memstats_frees_total"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Counter.Value
}

func memstats_heap_still_use() float64 {
	str := "go_memstats_heap_alloc_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_heap_waiting() float64 {
	str := "go_memstats_heap_idle_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_heap_in_use() float64 {
	str := "go_memstats_heap_inuse_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_heap_bytes_released() float64 {
	str := "go_memstats_heap_released_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func memstats_heap_sys() float64 {
	str := "go_memstats_heap_sys_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func GetMemstatsHeapObjects() float64 {
	str := "go_memstats_heap_objects"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func last_gc_time_seconds() float64 {
	str := "go_memstats_last_gc_time_seconds"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func ptr_lookups() float64 {
	str := "go_memstats_lookups_total"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Counter.Value
}

func mallocs_total() float64 {
	str := "go_memstats_mallocs_total"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Counter.Value
}

func mcache_inuse_bytes() float64 {
	str := "go_memstats_mcache_inuse_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func mcache_inuse_bytes_System() float64 {
	str := "go_memstats_mcache_sys_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func mspan_inuse_bytes() float64 {
	str := "go_memstats_mspan_inuse_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

func mspan_inuse_bytes_System() float64 {
	str := "go_memstats_mspan_sys_bytes"
	p := prometheusGather()
	m := metricRetriever(p, str)
	return *m.Gauge.Value
}

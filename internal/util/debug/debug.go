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
	"slices"
	"text/template"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

type metric struct {
	metricName string
	metricType int
}

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
	}
	for _, metricFamily := range prometheus_metrics {
		for _, metricSlice := range metricFamily.Metric {
			if metricSlice.Label != nil {
				for _, metricLabel := range metricSlice.Label {
					graphPlot := generateGraph(metricSlice, metricFamily, metricLabel)
					opts = append(opts, statsviz.TimeseriesPlot(graphPlot))
				}
			}
			graphPlot := generateGraphNonLable(metricSlice, metricFamily)
			opts = append(opts, statsviz.TimeseriesPlot(graphPlot))
		}
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

func (obj metric) getValue() float64 {
	x := prometheusGather()
	y := metricRetriever(x, obj.metricName)
	switch obj.metricType {
	case 0:
		return *y.Counter.Value
	case 1:
		return *y.Gauge.Value
	case 2:
		return *y.Summary.SampleSum
	case 3:
		return *y.Untyped.Value
	case 4:
		return *y.Histogram.SampleCountFloat
	}
	return 0
}

func generateGraph(metricSlice *io_prometheus_client.Metric, metricFamily *io_prometheus_client.MetricFamily, metricLabel *io_prometheus_client.LabelPair) statsviz.TimeSeriesPlot {
	key := fmt.Sprintf("%s %s %s", *metricLabel.Name, ":", *metricLabel.Value)

	finalMetricObj := new(metric)
	finalMetricObj.metricName = *metricFamily.Name
	finalMetricObj.metricType = int(*metricFamily.Type)

	GenericPlot := statsviz.TimeSeries{
		Name:     key,
		Unitfmt:  "%{y:.4s}",
		GetValue: finalMetricObj.getValue,
	}
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       *metricLabel.Name + *metricLabel.Value + *metricFamily.Name,
		Title:      *metricFamily.Help,
		Type:       statsviz.Scatter,
		Series:     []statsviz.TimeSeries{GenericPlot},
		YAxisTitle: metricFamily.Type.String(),
	}.Build()
	if err != nil {
		log.Fatalf("Failed to build timeseries plot :%v", err)
	}
	return plot
}

func generateGraphNonLable(metricSlice *io_prometheus_client.Metric, metricFamily *io_prometheus_client.MetricFamily) statsviz.TimeSeriesPlot {

	finalMetricObj := new(metric)
	finalMetricObj.metricName = *metricFamily.Name
	finalMetricObj.metricType = int(*metricFamily.Type)

	GenericPlot := statsviz.TimeSeries{
		Name:     *metricFamily.Name,
		Unitfmt:  "%{y:.4s}",
		GetValue: finalMetricObj.getValue,
	}

	x := generateRandomString(5)
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       *metricFamily.Name + x,
		Title:      *metricFamily.Help,
		Type:       statsviz.Scatter,
		Series:     []statsviz.TimeSeries{GenericPlot},
		YAxisTitle: "Quantity", //Temporay Modification : To be replaced with
	}.Build()
	if err != nil {
		log.Fatalf("Failed to build timeseries plot :%v", err)
	}
	return plot
}

//random string combination generated solves " panic : Duplicate plot name error for metric with different labelPair values"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(randomString)
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

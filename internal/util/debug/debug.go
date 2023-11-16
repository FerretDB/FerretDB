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

	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// struct metric stores respective information about each metricFamily instance
// which is used especially for GetValue attribute of the statsviz.Timeseries
// we need this object construct because GetValue attribute does not allow
// passing arguments for getting specified metrics and as a result
// passing the metricFamily name is not possible by any other way
// other than passing metricName through interface method.
type metric struct {
	metricName       string
	metricType       int
	metricLabel      string
	metricLabelValue string
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
	must.NoError(err)

	http.Handle("/debug/metrics", metricHandler)

	opts := []statsviz.Option{
		statsviz.Root("/debug/graphs"),
	}
	// metricFamily exposes individual metrics as a metricFamily
	// However there can be multiple MetricFamily elements with same Name but with different Label-pair values.
	for _, metricFamily := range prometheus_metrics {
		for _, metricSlice := range metricFamily.Metric {
			graphPlot := generateGraphAll(metricSlice, metricFamily)
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

// getValue repeatedly calls prometheusGatherer as the metrics plotted need to be continuously updated
// the empty quotation check is implemented to return metrics without any label-pair values since
// metrics without label-pair values have only a single instance
// the switch block checks the metric type based on int32 type standard
// mentioned at https://pkg.go.dev/github.com/prometheus/client_model@v0.5.0/go#MetricType
func (obj metric) getValue() float64 {
	prometheus_metrics := prometheusGather()
	if obj.metricLabel == "" {
		Individualmetric := metricRetriever(prometheus_metrics, obj.metricName)
		switch obj.metricType {
		case 0:
			return *Individualmetric.Counter.Value
		case 1:
			return *Individualmetric.Gauge.Value
		case 2:
			return *Individualmetric.Summary.SampleSum
		case 3:
			return *Individualmetric.Untyped.Value
		case 4:
			return *Individualmetric.Histogram.SampleCountFloat
		}
	} else {
		Individualmetric := LabelledMetricRetriever(prometheus_metrics, obj.metricName, obj)
		switch obj.metricType {
		case 0:
			return *Individualmetric.Counter.Value
		case 1:
			return *Individualmetric.Gauge.Value
		case 2:
			return *Individualmetric.Summary.SampleSum
		case 3:
			return *Individualmetric.Untyped.Value
		case 4:
			return *Individualmetric.Histogram.SampleCountFloat
		}
	}
	return 0
}

func generateGraphAll(metricSlice *dto.Metric,
	metricFamily *dto.MetricFamily,
) statsviz.TimeSeriesPlot {
	finalMetricObj := new(metric)
	finalMetricObj.metricName = *metricFamily.Name
	finalMetricObj.metricType = int(*metricFamily.Type)

	GenericPlot := statsviz.TimeSeries{
		Name:     *metricFamily.Name,
		Unitfmt:  "%{y:.4s}",
		GetValue: finalMetricObj.getValue,
	}

	if metricSlice.Label != nil {
		for _, individualLabelPair := range metricSlice.Label {
			finalMetricObj.metricLabel = individualLabelPair.GetName()
			finalMetricObj.metricLabelValue = individualLabelPair.GetValue()
		}
		GenericPlot.Name = finalMetricObj.metricLabel + " : " + finalMetricObj.metricLabelValue
	}

	finalMetricObj.metricLabel = ""
	finalMetricObj.metricLabelValue = ""

	x := generateRandomString(5)
	plot, err := statsviz.TimeSeriesPlotConfig{
		Name:       *metricFamily.Name + x,
		Title:      *metricFamily.Help,
		Type:       statsviz.Scatter,
		Series:     []statsviz.TimeSeries{GenericPlot},
		YAxisTitle: "", // Units Change depending on the data , therefore Unit cannot be generated programatically
	}.Build()
	if err != nil {
		log.Fatalf("Failed to build timeseries plot :%v", err)
	}
	return plot
}

// Random string combination generated solves " panic : Duplicate plot name error for metrics  with different labelPair values".

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(randomString)
}

// metricRetriever returns *io_prometheus_client.Metric whose Name matches provided input.
func metricRetriever(prometheus_metrics []*dto.MetricFamily, metricName string) *dto.Metric {
	for _, specificMetric := range prometheus_metrics {
		if *specificMetric.Name == metricName {
			finalMetricSlice := specificMetric.GetMetric()
			for _, x := range finalMetricSlice {
				return x
			}
		}
	}
	return nil
}

func LabelledMetricRetriever(prometheus_metrics []*dto.MetricFamily,
	metricName string, obj metric,
) *dto.Metric {
	for _, specificMetric := range prometheus_metrics {
		if *specificMetric.Name == metricName {
			for _, x := range specificMetric.Metric {
				for _, y := range x.Label {
					if y.GetValue() == obj.metricLabelValue {
						return x
					}
				}
			}
		}
	}
	return nil
}

func prometheusGather() []*dto.MetricFamily {
	prometheus_metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Fatalf("error in Gathering prometheus metrics: %v", err)
	}
	return prometheus_metrics
}

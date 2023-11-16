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

package debug

import (
	"log"
	"math/rand"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/arl/statsviz"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type plotter struct {
	g prometheus.Gatherer
}

func newPlotter(g prometheus.Gatherer) *plotter {
	return &plotter{
		g: g,
	}
}

func (p *plotter) plots() ([]statsviz.TimeSeriesPlot, error) {
	return nil, nil
}

// struct metric stores respective information about each metricFamily instance
// which is used especially for GetValue attribute of the statsviz.Timeseries
// we need this object construct because GetValue attribute does not allow
// passing arguments for getting specified metrics and as a result
// passing the metricFamily name is not possible by any other way
// other than passing metricName through interface method.
type metric struct {
	p                *plotter
	metricName       string
	metricType       dto.MetricType
	metricLabel      string
	metricLabelValue string
}

// getValue repeatedly calls prometheusGatherer as the metrics plotted need to be continuously updated
// the empty quotation check is implemented to return metrics without any label-pair values since
// metrics without label-pair values have only a single instance
// the switch block checks the metric type based on int32 type standard
// mentioned at https://pkg.go.dev/github.com/prometheus/client_model@v0.5.0/go#MetricType
func (obj metric) getValue() float64 {
	m := must.NotFail(obj.p.g.Gather())

	if obj.metricLabel == "" {
		Individualmetric := metricRetriever(m, obj.metricName)
		switch obj.metricType {
		case dto.MetricType_COUNTER:
			return *Individualmetric.Counter.Value
		case dto.MetricType_GAUGE:
			return *Individualmetric.Gauge.Value
		case dto.MetricType_SUMMARY:
			return *Individualmetric.Summary.SampleSum
		case dto.MetricType_UNTYPED:
			return *Individualmetric.Untyped.Value
		case dto.MetricType_HISTOGRAM:
			return *Individualmetric.Histogram.SampleCountFloat
		}
	} else {
		Individualmetric := labelledMetricRetriever(m, obj.metricName, obj)
		switch obj.metricType {
		case dto.MetricType_COUNTER:
			return *Individualmetric.Counter.Value
		case dto.MetricType_GAUGE:
			return *Individualmetric.Gauge.Value
		case dto.MetricType_SUMMARY:
			return *Individualmetric.Summary.SampleSum
		case dto.MetricType_UNTYPED:
			return *Individualmetric.Untyped.Value
		case dto.MetricType_HISTOGRAM:
			return *Individualmetric.Histogram.SampleCountFloat
		}
	}

	return 0
}

func (p *plotter) generateGraphAll(metricSlice *dto.Metric, metricFamily *dto.MetricFamily) statsviz.TimeSeriesPlot {
	finalMetricObj := new(metric)
	finalMetricObj.p = p
	finalMetricObj.metricName = *metricFamily.Name
	finalMetricObj.metricType = *metricFamily.Type

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
func metricRetriever(m []*dto.MetricFamily, metricName string) *dto.Metric {
	for _, specificMetric := range m {
		if *specificMetric.Name == metricName {
			finalMetricSlice := specificMetric.GetMetric()
			for _, x := range finalMetricSlice {
				return x
			}
		}
	}
	return nil
}

func labelledMetricRetriever(m []*dto.MetricFamily, metricName string, obj metric) *dto.Metric {
	for _, specificMetric := range m {
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

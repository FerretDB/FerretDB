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
	"strings"

	"github.com/arl/statsviz"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type plotter struct {
	g *gatherer
	l *zap.Logger
}

func newPlotter(g *gatherer, l *zap.Logger) *plotter {
	return &plotter{
		g: g,
		l: l,
	}
}

func (p *plotter) plots() ([]statsviz.TimeSeriesPlot, error) {
	var res []statsviz.TimeSeriesPlot

	for name, mf := range p.g.GatherMap() {
		switch *mf.Type {
		case dto.MetricType_COUNTER, dto.MetricType_GAUGE:
			// nothing
		case dto.MetricType_SUMMARY, dto.MetricType_UNTYPED, dto.MetricType_HISTOGRAM, dto.MetricType_GAUGE_HISTOGRAM:
			continue
		default:
			return nil, lazyerrors.Errorf("unexpected metric type %v", *mf.Type)
		}

		for _, m := range mf.Metric {
			help := strings.TrimSuffix(*mf.Help, ".")

			if len(m.Label) == 0 {
				c := statsviz.TimeSeriesPlotConfig{
					Name:     name,
					Title:    help,
					Type:     statsviz.Scatter,
					InfoText: name,
					Series: []statsviz.TimeSeries{{
						Name:     name,
						Unitfmt:  "",
						HoverOn:  "",
						Type:     statsviz.Scatter,
						GetValue: newGraph(p.g, p.l, name, help, "", "").getValue,
					}},
				}

				p, err := c.Build()
				if err != nil {
					return nil, lazyerrors.Error(err)
				}

				res = append(res, p)
				continue
			}

			// make one plot per metric name + label name
			// TODO

			// make one graph per label value
		}
	}

	return res, nil
}

/*
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
*/

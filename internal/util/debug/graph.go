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
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

// graph represents a single graph on a single plot.
type graph struct {
	g          *gatherer
	l          *zap.Logger
	metricName string
	help       string
	labelName  string
	labelValue string
}

// newGraph creates a new graph.
func newGraph(g *gatherer, l *zap.Logger, metricName, help, labelName, labelValue string) *graph {
	return &graph{
		g:          g,
		l:          l,
		metricName: metricName,
		help:       help,
		labelName:  labelName,
		labelValue: labelValue,
	}
}

// getValue returns the current value.
func (g *graph) getValue() float64 {
	mf := g.g.GatherMap()[g.metricName]
	if mf == nil {
		g.l.Warn("Prometheus metric not found", zap.String("metric", g.metricName))
		return 0
	}

	var v float64

	for _, m := range mf.Metric {
		if g.labelName == "" {
			v += metricValue(m)
			continue
		}

		for _, lp := range m.Label {
			if *lp.Name != g.labelName {
				continue
			}

			if *lp.Value == g.labelValue {
				v += metricValue(m)
			}

			break
		}
	}

	return v
}

func metricValue(m *dto.Metric) float64 {
	if m.Counter != nil {
		return *m.Counter.Value
	}

	return *m.Gauge.Value
}

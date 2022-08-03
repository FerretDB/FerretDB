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

package clientconn

import "github.com/prometheus/client_golang/prometheus"

// ConnMetrics represents conn metrics.
type ConnMetrics struct {
	requests          *prometheus.CounterVec
	responses         *prometheus.CounterVec
	aggregationStages *prometheus.CounterVec
}

// newConnMetrics creates new conn metrics.
func newConnMetrics() *ConnMetrics {
	return &ConnMetrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of requests.",
			},
			[]string{"opcode", "command"},
		),
		responses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "responses_total",
				Help:      "Total number of responses.",
			},
			[]string{"opcode", "command", "result"},
		),
		aggregationStages: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "aggregation_stages_total",
				Help:      "Total number of aggregation pipeline stages.",
			},
			[]string{"opcode", "command", "stage"},
		),
	}
}

// Describe implements prometheus.Collector.
func (cm *ConnMetrics) Describe(ch chan<- *prometheus.Desc) {
	cm.requests.Describe(ch)
	cm.responses.Describe(ch)
	cm.aggregationStages.Describe(ch)
}

// Collect implements prometheus.Collector.
func (cm *ConnMetrics) Collect(ch chan<- prometheus.Metric) {
	cm.requests.Collect(ch)
	cm.responses.Collect(ch)
	cm.aggregationStages.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*ConnMetrics)(nil)
)

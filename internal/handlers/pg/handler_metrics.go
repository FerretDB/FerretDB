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

package pg

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "ferretdb"
	subsystem = "handler"
)

// Metrics represents handler metrics.
type Metrics struct {
	requests  *prometheus.CounterVec
	responses *prometheus.CounterVec
}

// NewMetrics creates new handler metrics.
func NewMetrics() *Metrics {
	return &Metrics{
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
	}
}

// Describe implements prometheus.Collector.
func (lm *Metrics) Describe(ch chan<- *prometheus.Desc) {
	lm.requests.Describe(ch)
	lm.responses.Describe(ch)
}

// Collect implements prometheus.Collector.
func (lm *Metrics) Collect(ch chan<- prometheus.Metric) {
	lm.requests.Collect(ch)
	lm.responses.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Metrics)(nil)
)

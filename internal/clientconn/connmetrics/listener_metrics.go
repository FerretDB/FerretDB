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

package connmetrics

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "ferretdb"
	subsystem = "client"
)

// ListenerMetrics represents listener metrics.
type ListenerMetrics struct {
	ConnectedClients prometheus.Gauge
	Accepts          *prometheus.CounterVec
	ConnMetrics      *ConnMetrics
}

// NewListenerMetrics creates new listener metrics.
func NewListenerMetrics() *ListenerMetrics {
	return &ListenerMetrics{
		ConnectedClients: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "connected",
				Help:      "The current number of connected clients.",
			},
		),
		Accepts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "accepts_total",
				Help:      "Total number of accepted client connections.",
			},
			[]string{"error"},
		),
		ConnMetrics: newConnMetrics(),
	}
}

// Describe implements prometheus.Collector.
func (lm *ListenerMetrics) Describe(ch chan<- *prometheus.Desc) {
	lm.ConnectedClients.Describe(ch)
	lm.Accepts.Describe(ch)
	lm.ConnMetrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (lm *ListenerMetrics) Collect(ch chan<- prometheus.Metric) {
	lm.ConnectedClients.Collect(ch)
	lm.Accepts.Collect(ch)
	lm.ConnMetrics.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*ListenerMetrics)(nil)
)

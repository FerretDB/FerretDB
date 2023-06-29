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

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "ferretdb"
	subsystem = "client"
)

// ListenerMetrics represents listener metrics.
type ListenerMetrics struct {
	Accepts     *prometheus.CounterVec
	Durations   *prometheus.HistogramVec
	ConnMetrics *ConnMetrics
}

// NewListenerMetrics creates new listener metrics.
func NewListenerMetrics() *ListenerMetrics {
	return &ListenerMetrics{
		Accepts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "accepts_total",
				Help:      "Total number of accepted client connections.",
			},
			[]string{"error"},
		),
		Durations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "duration_seconds",
				Help:      "Client connection lifetime in seconds.",
				Buckets: []float64{
					1,
					5,
					10,
					30,
					(1 * time.Minute).Seconds(),
					(5 * time.Minute).Seconds(),
					(10 * time.Minute).Seconds(),
					(30 * time.Minute).Seconds(),
				},
			},
			[]string{"error"},
		),

		ConnMetrics: newConnMetrics(),
	}
}

// Describe implements prometheus.Collector.
func (lm *ListenerMetrics) Describe(ch chan<- *prometheus.Desc) {
	lm.Accepts.Describe(ch)
	lm.Durations.Describe(ch)
	lm.ConnMetrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (lm *ListenerMetrics) Collect(ch chan<- prometheus.Metric) {
	lm.Accepts.Collect(ch)
	lm.Durations.Collect(ch)
	lm.ConnMetrics.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*ListenerMetrics)(nil)
)

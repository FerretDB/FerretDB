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

package fsql

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "ferretdb"
	subsystem = "sqldb"
)

// metricsCollector exposes DB's state as Prometheus metrics.
type metricsCollector struct {
	labels prometheus.Labels
	db     DB
}

// NewMetricsCollector creates a new metricsCollector.
func NewMetricsCollector(name string, db DB) *metricsCollector {
	return &metricsCollector{
		db: db,
		labels: prometheus.Labels{
			"name": name,
		},
	}
}

// Describe implements prometheus.Collector.
func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements prometheus.Collector.
func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open"),
			"The number of established connections both in use and idle.",
			nil, c.labels,
		),
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*metricsCollector)(nil)
)

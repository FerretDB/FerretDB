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
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "ferretdb"
	subsystem = "sqldb"
)

// metricsCollector exposes DB's state as Prometheus metrics.
type metricsCollector struct {
	labels prometheus.Labels
	statsF func() sql.DBStats
}

// newMetricsCollector creates a new metricsCollector.
func newMetricsCollector(db string, statsF func() sql.DBStats) *metricsCollector {
	return &metricsCollector{
		statsF: statsF,
		labels: prometheus.Labels{
			"db": db,
		},
	}
}

// Describe implements prometheus.Collector.
func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements prometheus.Collector.
func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.statsF()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open_max"),
			"Maximum number of open connections to the database.",
			nil, c.labels,
		),
		prometheus.GaugeValue,
		float64(stats.MaxOpenConnections),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open"),
			"The number of established connections both in use and idle.",
			nil, c.labels,
		),
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)

	// Add other metrics from stats here.
	// TODO https://github.com/FerretDB/FerretDB/issues/2909

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_lifetime_closed_total"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil, c.labels,
		),
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*metricsCollector)(nil)
)

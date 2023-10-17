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

package state

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/build/version"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = ""
)

// metricsCollector exposes provider's state as Prometheus metrics.
type metricsCollector struct {
	p               *Provider
	addUUIDToMetric bool
}

// newMetricsCollector creates a new metricsCollector.
//
// If addUUIDToMetric is true, then the UUID is added to the Prometheus metric.
func newMetricsCollector(p *Provider, addUUIDToMetric bool) *metricsCollector {
	return &metricsCollector{
		p:               p,
		addUUIDToMetric: addUUIDToMetric,
	}
}

// Describe implements prometheus.Collector.
func (mc *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

// Collect implements prometheus.Collector.
func (mc *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	info := version.Get()
	constLabels := prometheus.Labels{
		"version": info.Version,
		"commit":  info.Commit,
		"branch":  info.Branch,
		"dirty":   strconv.FormatBool(info.Dirty),
		"package": info.Package,
		"debug":   strconv.FormatBool(info.DebugBuild),
	}

	s := mc.p.Get()

	constLabels["telemetry"] = s.TelemetryString()
	constLabels["update_available"] = strconv.FormatBool(s.UpdateAvailable)

	if mc.addUUIDToMetric {
		constLabels["uuid"] = s.UUID
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "up"),
			"FerretDB instance state.",
			[]string{"backend_name", "backend_version"},
			constLabels,
		),
		prometheus.GaugeValue,
		1,
		s.BackendName,
		s.BackendVersion,
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*metricsCollector)(nil)
)

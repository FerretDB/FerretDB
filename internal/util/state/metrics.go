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

	"github.com/FerretDB/FerretDB/v2/build/version"
)

// metricsCollector exposes provider's state as Prometheus metric.
type metricsCollector struct {
	p       *Provider
	addUUID bool
}

// newMetricsCollector creates a new metricsCollector.
//
// If addUUID is true, then the "uuid" label is added.
func newMetricsCollector(p *Provider, addUUID bool) *metricsCollector {
	return &metricsCollector{
		p:       p,
		addUUID: addUUID,
	}
}

// Describe implements [prometheus.Collector].
func (mc *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

// Collect implements [prometheus.Collector].
// It exposes a single metric with various labels.
func (mc *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	info := version.Get()
	constLabels := prometheus.Labels{
		"version": info.Version,
		"commit":  info.Commit,
		"branch":  info.Branch,
		"dirty":   strconv.FormatBool(info.Dirty),
		"package": info.Package,
		"dev":     strconv.FormatBool(info.DevBuild),
	}

	labels := []string{
		"postgresql",
		"documentdb",
		"telemetry",
		"update_available",
	}
	if mc.addUUID {
		labels = append(labels, "uuid")
	}

	s := mc.p.Get()

	values := []string{
		s.PostgreSQLVersion,
		s.DocumentDBVersion,
		s.TelemetryString(),
		strconv.FormatBool(s.UpdateAvailable),
	}
	if mc.addUUID {
		values = append(values, s.UUID)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"ferretdb_up",
			"FerretDB instance state.",
			labels,
			constLabels,
		),
		prometheus.GaugeValue,
		1,
		values...,
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*metricsCollector)(nil)
)

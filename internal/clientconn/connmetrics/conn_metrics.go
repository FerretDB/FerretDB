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

// Package connmetrics provides listener and connection metrics.
package connmetrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ConnMetrics represents conn metrics.
type ConnMetrics struct {
	Requests  *prometheus.CounterVec
	Responses *prometheus.CounterVec
}

// commandMetrics represents command results metrics.
type commandMetrics struct {
	Failures map[string]int // count by error codes; no "ok" there
	Total    int            // both ok and errors
}

// newConnMetrics creates connection metrics.
func newConnMetrics() *ConnMetrics {
	return &ConnMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of requests.",
			},
			[]string{"opcode", "command"},
		),
		Responses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "responses_total",
				Help:      "Total number of responses.",
			},
			[]string{"opcode", "command", "argument", "result"},
		),
	}
}

// Describe implements prometheus.Collector.
func (cm *ConnMetrics) Describe(ch chan<- *prometheus.Desc) {
	cm.Requests.Describe(ch)
	cm.Responses.Describe(ch)
}

// Collect implements prometheus.Collector.
func (cm *ConnMetrics) Collect(ch chan<- prometheus.Metric) {
	cm.Requests.Collect(ch)
	cm.Responses.Collect(ch)
}

// GetResponses returns a map with all response metrics:
//
// opcode (e.g. "OP_MSG", "OP_QUERY") ->
// command (e.g. "update", "aggregate") ->
// argument that caused an error (e.g. "$set", "$count (stage)"; or "unknown") ->
// result (e.g. "NotImplemented", "InternalError"; or "ok") ->
// count.
func (cm *ConnMetrics) GetResponses() map[string]map[string]map[string]commandMetrics {
	metrics := make(chan prometheus.Metric)
	go func() {
		cm.Responses.Collect(metrics)
		close(metrics)
	}()

	res := map[string]map[string]map[string]commandMetrics{}

	for m := range metrics {
		var content dto.Metric
		must.NoError(m.Write(&content))

		var opcode, command, argument, result string
		for _, label := range content.GetLabel() {
			switch label.GetName() {
			case "opcode":
				opcode = label.GetValue()
			case "command":
				command = label.GetValue()
			case "argument":
				argument = label.GetValue()
			case "result":
				result = label.GetValue()
			default:
				panic(fmt.Sprintf("%s is not a valid label. Allowed: [opcode, command, argument, result]", label.GetName()))
			}
		}

		if _, ok := res[opcode]; !ok {
			res[opcode] = map[string]map[string]commandMetrics{}
		}

		if _, ok := res[opcode][command]; !ok {
			res[opcode][command] = map[string]commandMetrics{}
		}

		if _, ok := res[opcode][command][argument]; !ok {
			res[opcode][command][argument] = commandMetrics{}
		}

		m := res[opcode][command][argument]

		v := int(content.GetCounter().GetValue())
		m.Total += v

		if result != "ok" {
			if m.Failures == nil {
				m.Failures = map[string]int{}
			}
			m.Failures[result] += v
		}

		res[opcode][command][argument] = m
	}

	return res
}

// check interfaces
var (
	_ prometheus.Collector = (*ConnMetrics)(nil)
)

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

package middleware

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Parts of Prometheus metric names.
// TODO https://github.com/FerretDB/FerretDB/issues/4965
const (
	namespace = "ferretdb"
	subsystem = "client"
)

// Metrics represents middleware Metrics.
type Metrics struct {
	requests  *prometheus.CounterVec
	responses *prometheus.CounterVec
}

// CommandMetrics represents command results metrics.
type CommandMetrics struct {
	Failures map[string]int // count by result, except "ok"
	Total    int            // both "ok" and failures
}

// NewMetrics creates new metrics.
func NewMetrics() *Metrics {
	// Do we want to use "opcode" as a label?
	// Should we use to track the listener that created the request?
	// Or metric for that should be in the listener itself?
	// TODO https://github.com/FerretDB/FerretDB/issues/4965
	m := &Metrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of requests.",
			},
			[]string{"opcode", "command"},
		),

		// That probably should be a histogram or summary by duration.
		// TODO https://github.com/FerretDB/FerretDB/issues/4965
		responses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "responses_total",
				Help:      "Total number of responses.",
			},
			[]string{"opcode", "command", "argument", "result"},
		),
	}

	m.requests.With(prometheus.Labels{
		"opcode":  "OP_MSG",
		"command": "find",
	})
	m.responses.With(prometheus.Labels{
		"opcode":   "OP_MSG",
		"command":  "find",
		"argument": "unknown",
		"result":   string(resultOK),
	})

	return m
}

// Describe implements [prometheus.Collector].
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.responses.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.responses.Collect(ch)
}

// GetResponses returns a map with all response metrics:
//
// opcode (e.g. "OP_MSG", "OP_QUERY") ->
// command (e.g. "find", "aggregate") ->
// argument that caused an error (e.g. "sort", "$count (stage)"; or "unknown") ->
// result (e.g. "ok", "NotImplemented", "error", or "panic") ->
// count.
func (m *Metrics) GetResponses() map[string]map[string]map[string]CommandMetrics {
	metrics := make(chan prometheus.Metric)
	go func() {
		m.responses.Collect(metrics)
		close(metrics)
	}()

	res := map[string]map[string]map[string]CommandMetrics{}

	for m := range metrics {
		var content dto.Metric
		must.NoError(m.Write(&content))

		var opcode, command, argument, result string

		for _, label := range content.GetLabel() {
			v := label.GetValue()

			switch name := label.GetName(); name {
			case "opcode":
				opcode = v
			case "command":
				command = v
			case "argument":
				argument = v
			case "result":
				result = v
			default:
				panic(fmt.Sprintf("%q is not a valid label. Allowed: [opcode, command, argument, result]", name))
			}
		}

		if _, ok := res[opcode]; !ok {
			res[opcode] = map[string]map[string]CommandMetrics{}
		}

		if _, ok := res[opcode][command]; !ok {
			res[opcode][command] = map[string]CommandMetrics{}
		}

		if _, ok := res[opcode][command][argument]; !ok {
			res[opcode][command][argument] = CommandMetrics{}
		}

		m := res[opcode][command][argument]

		v := int(content.GetCounter().GetValue())
		m.Total += v

		if result != string(resultOK) {
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
	_ prometheus.Collector = (*Metrics)(nil)
)

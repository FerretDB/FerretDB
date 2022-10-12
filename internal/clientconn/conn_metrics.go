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

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ConnMetrics represents conn metrics.
type ConnMetrics struct {
	requests          *prometheus.CounterVec
	responses         *prometheus.CounterVec
	aggregationStages *prometheus.CounterVec

	cmds []string
}

type CommandMetrics interface{}

type BasicCommandMetrics struct {
	Failed int64
	Total  int64
}

type UpdateCommandMetrics struct {
	ArrayFilters int64
	Failed       int64
	Pipeline     int64
	Total        int64
}

// newConnMetrics creates new conn metrics.
//
// The cmds is the list of all expected commands that could be measured.
// After providing them, they will be set with the zero values.
func newConnMetrics(cmds []string) *ConnMetrics {
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
			[]string{"command", "stage"},
		),
		cmds: cmds,
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

func (cm *ConnMetrics) Responses() map[string]CommandMetrics {
	res := make(map[string]CommandMetrics)

	// initialize commands in the map to show zero values in the metrics output
	for _, cmd := range cm.cmds {
		// update related operators have more fields in the output
		switch cmd {
		case "update", "clusterUpdate", "findAndModify":
			res[cmd] = UpdateCommandMetrics{}
		default:
			res[cmd] = BasicCommandMetrics{}
		}
	}

	ch := make(chan prometheus.Metric)
	go func() {
		cm.responses.Collect(ch)
		cm.aggregationStages.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		var content dto.Metric
		must.NoError(m.Write(&content))

		var stage int

		var command, opcode, result string
		for _, label := range content.GetLabel() {
			switch label.GetName() {
			case "command":
				command = label.GetValue()
			case "opcode":
				opcode = label.GetValue()
			case "result":
				result = label.GetValue()
			case "stage":
				stage = must.NotFail(strconv.Atoi(label.GetValue()))
			default:
				panic(fmt.Sprintf("%s is not a valid label. Allowed: [command, opcode, result, stage]", label.GetName()))
			}
		}

		if opcode != "OP_MSG" {
			continue
		}

		value := int64(content.GetCounter().GetValue())

		cm := res[command]
		if cm == nil {
			cm = BasicCommandMetrics{}
		}

		switch cm := cm.(type) {
		case UpdateCommandMetrics:
			cm.Total += value
			if result != "ok" && result != "Unset" {
				cm.Failed += value
			}
			cm.Pipeline += int64(stage)
			// TODO: add metrics for arrayFilters
			res[command] = cm
		case BasicCommandMetrics:
			cm.Total += value
			if result != "ok" && result != "Unset" {
				cm.Failed += value
			}
			res[command] = cm
		default:
			panic(fmt.Sprintf("Invalid command metric type: %T", cm))
		}
	}

	return res
}

// check interfaces
var (
	_ prometheus.Collector = (*ConnMetrics)(nil)
)

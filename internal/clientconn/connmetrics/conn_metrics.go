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
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ConnMetrics represents conn metrics.
type ConnMetrics struct {
	Requests          *prometheus.CounterVec
	Responses         *prometheus.CounterVec
	AggregationStages *prometheus.CounterVec

	cmds []string
}

// CommandMetrics represents metrics for a single command.
type CommandMetrics interface{}

// BasicCommandMetrics contains all metrics fields used in most of commands.
type BasicCommandMetrics struct {
	Failed int64
	Total  int64
}

// UpdateCommandMetrics contains all metrics fields used in update, clusterUpdate and findAndModify commands.
type UpdateCommandMetrics struct {
	ArrayFilters int64
	Failed       int64
	Pipeline     int64
	Total        int64
}

// newConnMetrics creates connection metrics.
//
// The cmds is the list of all expected commands that could be measured.
// Commands provided in the list will be set with zero values.
func newConnMetrics(cmds []string) *ConnMetrics {
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
			[]string{"opcode", "command", "result"},
		),
		AggregationStages: prometheus.NewCounterVec(
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
	cm.Requests.Describe(ch)
	cm.Responses.Describe(ch)
	cm.AggregationStages.Describe(ch)
}

// Collect implements prometheus.Collector.
func (cm *ConnMetrics) Collect(ch chan<- prometheus.Metric) {
	cm.Requests.Collect(ch)
	cm.Responses.Collect(ch)
	cm.AggregationStages.Collect(ch)
}

// GetResponses returns a map with all metrics related to all commands.
// The key in the map is the command name and the value is a struct with
// all related metrics to this command.
func (cm *ConnMetrics) GetResponses() map[string]CommandMetrics {
	res := map[string]CommandMetrics{}

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

	metrics := make(chan prometheus.Metric)

	go func() {
		cm.Responses.Collect(metrics)
		cm.AggregationStages.Collect(metrics)
		close(metrics)
	}()

	for m := range metrics {
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

		cmdMetrics := res[command]
		if cmdMetrics == nil {
			cmdMetrics = BasicCommandMetrics{}
		}

		switch cmdMetrics := cmdMetrics.(type) {
		case UpdateCommandMetrics:
			cmdMetrics.Total += value
			if result != "ok" && result != "Unset" {
				cmdMetrics.Failed += value
			}
			cmdMetrics.Pipeline += int64(stage)
			// TODO: https://github.com/FerretDB/FerretDB/issues/1259
			res[command] = cmdMetrics
		case BasicCommandMetrics:
			cmdMetrics.Total += value
			if result != "ok" && result != "Unset" {
				cmdMetrics.Failed += value
			}
			res[command] = cmdMetrics
		default:
			panic(fmt.Sprintf("Invalid command metric type: %T", cmdMetrics))
		}
	}

	return res
}

// check interfaces
var (
	_ prometheus.Collector = (*ConnMetrics)(nil)
)

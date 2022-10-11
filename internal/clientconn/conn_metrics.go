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
	"log"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ConnMetrics represents conn metrics.
type ConnMetrics struct {
	requests          *prometheus.CounterVec
	responses         *prometheus.CounterVec
	aggregationStages *prometheus.CounterVec
}

type CommandMetrics struct {
	Failed int64
	Total  int64
}

// newConnMetrics creates new conn metrics.
func newConnMetrics() *ConnMetrics {
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

	ch := make(chan prometheus.Metric)
	go func() {
		cm.responses.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		var content dto.Metric
		must.NoError(m.Write(&content))

		// # HELP ferretdb_client_responses_total Total number of responses.
		// # TYPE ferretdb_client_responses_total counter
		// ferretdb_client_responses_total{command="",opcode="OP_REPLY",result="ok"} 5
		// ferretdb_client_responses_total{command="atlasVersion",opcode="OP_MSG",result="CommandNotFound"} 1
		// ferretdb_client_responses_total{command="buildInfo",opcode="OP_MSG",result="ok"} 1
		// ferretdb_client_responses_total{command="getCmdLineOpts",opcode="OP_MSG",result="ok"} 1
		// ferretdb_client_responses_total{command="getFreeMonitoringStatus",opcode="OP_MSG",result="ok"} 1
		// ferretdb_client_responses_total{command="getLog",opcode="OP_MSG",result="ok"} 1
		// ferretdb_client_responses_total{command="getParameter",opcode="OP_MSG",result="Unset"} 1
		// ferretdb_client_responses_total{command="hello",opcode="OP_MSG",result="ok"} 1
		// ferretdb_client_responses_total{command="hello",opcode="OP_MSG",result="SomethingBad"} 1
		// ferretdb_client_responses_total{command="notImplementedCommand",opcode="OP_MSG",result="CommandNotFound"} 1
		// ferretdb_client_responses_total{command="ping",opcode="OP_MSG",result="ok"} 1

		var command, opcode, result string
		for _, label := range content.GetLabel() {
			switch label.GetName() {
			case "command":
				command = label.GetValue()
			case "opcode":
				opcode = label.GetValue()
			case "result":
				result = label.GetValue()
			default:
				panic("oops")
			}
		}

		if opcode != "OP_MSG" {
			continue
		}

		value := int64(content.GetCounter().GetValue())

		cm := res[command]
		cm.Total += value
		if result != "ok" {
			cm.Failed += value
		}
		res[command] = cm

		log.Println(command, result, *content.Counter.Value)

		if result != "ok" {
			failed++
		}

		if len(results) > 10 {
			panic(results)
		}

		//if v, ok := cmdResps[cmd]; ok {
		//	failed = v.Failed + failed
		//}

		cmdResps[command] = CommandMetrics{
			Total:  int64(*content.Counter.Value),
			Failed: failed,
		}

		// TODO check renameMe.Label
		// TODO use renameMe.Counter
	}

	return cmdResps
}

// check interfaces
var (
	_ prometheus.Collector = (*ConnMetrics)(nil)
)

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResponses(t *testing.T) {
	m := newConnMetrics()
	m.Responses.WithLabelValues("OP_MSG", "update", "$set", "ok").Inc()
	m.Responses.WithLabelValues("OP_MSG", "update", "$set", "panic").Inc()
	expected := map[string]map[string]map[string]commandMetrics{
		"OP_MSG": {
			"update": {
				"$set": commandMetrics{
					Failures: map[string]int{
						"panic": 1,
					},
					Total: 2,
				},
			},
		},
	}
	assert.Equal(t, expected, m.GetResponses())
}

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

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCommandsFreeMonitoringGetFreeMonitoringStatus(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	expected := map[string]any{
		"state": "disabled",
		"ok":    float64(1),
	}

	var actual bson.D
	err := s.Collection.Database().RunCommand(s.Ctx, bson.D{{"getFreeMonitoringStatus", 1}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	keys := CollectKeys(t, actual)

	for k, item := range expected {
		assert.Contains(t, keys, k)
		assert.IsType(t, item, m[k])

		if it, ok := item.(primitive.D); ok {
			z := m[k].(primitive.D)
			AssertEqualDocuments(t, it, z)
			continue
		}
		assert.Equal(t, m[k], item)
	}
}

func TestCommandsFreeMonitoringSetFreeMonitoring(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
	}{
		"Enable": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", "enable"}},
			err: &mongo.CommandError{
				Code:    50840,
				Name:    "Location50840",
				Message: `Free Monitoring has been disabled via the command-line and/or config file`,
			},
		},
		"Disable": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", "disable"}},
			err: &mongo.CommandError{
				Code:    50840,
				Name:    "Location50840",
				Message: `Free Monitoring has been disabled via the command-line and/or config file`,
			},
		},
		"Other": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", "foobar"}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Enumeration value 'foobar' for field 'setFreeMonitoring.action' is not a valid value.`,
			},
		},
		"Empty": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", ""}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Enumeration value '' for field 'setFreeMonitoring.action' is not a valid value.`,
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := s.Collection.Database().RunCommand(s.Ctx, tc.command).Decode(&actual)

			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

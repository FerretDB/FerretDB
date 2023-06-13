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

package free_monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCommandsFreeMonitoringGetFreeMonitoringStatus(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	necessaryKeys := []string{"state", "ok"}

	allowedKeys := []string{"state", "message", "url", "userReminder", "ok"}

	var actual bson.D
	err := s.Collection.Database().RunCommand(s.Ctx, bson.D{{"getFreeMonitoringStatus", 1}}).Decode(&actual)
	require.NoError(t, err)

	keys := integration.CollectKeys(t, actual)

	for _, k := range allowedKeys {
		assert.Contains(t, allowedKeys, k)
	}

	for _, k := range necessaryKeys {
		assert.Contains(t, keys, k)
	}
}

func TestCommandsFreeMonitoringSetFreeMonitoring(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	for name, tc := range map[string]struct {
		command        bson.D // required, command to run
		expectedRes    bson.D // optional, expected response
		expectedStatus string // optional, expected status

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Enable": {
			command:        bson.D{{"setFreeMonitoring", 1}, {"action", "enable"}},
			expectedRes:    bson.D{{"ok", float64(1)}},
			expectedStatus: "enabled",
		},
		"Disable": {
			command:        bson.D{{"setFreeMonitoring", 1}, {"action", "disable"}},
			expectedRes:    bson.D{{"ok", float64(1)}},
			expectedStatus: "disabled",
		},
		"Other": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", "foobar"}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Enumeration value 'foobar' for field 'setFreeMonitoring.action' is not a valid value.`,
			},
			altMessage: `Enumeration value 'foobar' for field 'setFreeMonitoring.action' is not a valid value.`,
		},
		"Empty": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", ""}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Enumeration value '' for field 'setFreeMonitoring.action' is not a valid value.`,
			},
		},
		"DocumentCommand": {
			command:        bson.D{{"setFreeMonitoring", bson.D{}}, {"action", "enable"}},
			expectedRes:    bson.D{{"ok", float64(1)}},
			expectedStatus: "enabled",
		},
		"NilCommand": {
			command:        bson.D{{"setFreeMonitoring", nil}, {"action", "enable"}},
			expectedRes:    bson.D{{"ok", float64(1)}},
			expectedStatus: "enabled",
		},
		"ActionMissing": {
			command: bson.D{{"setFreeMonitoring", nil}},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: `BSON field 'setFreeMonitoring.action' is missing but a required field`,
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2704",
		},
		"ActionTypeDocument": {
			command: bson.D{{"setFreeMonitoring", 1}, {"action", bson.D{}}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'setFreeMonitoring.action' is the wrong type 'object', expected type 'string'",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2704",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			// these tests shouldn't be run in parallel, because they work on the same database

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			require.NotNil(t, tc.command, "command must not be nil")

			var res bson.D
			err := s.Collection.Database().RunCommand(s.Ctx, tc.command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			integration.AssertEqualDocuments(t, tc.expectedRes, res)

			if tc.expectedStatus != "" {
				var actual bson.D
				err := s.Collection.Database().RunCommand(s.Ctx, bson.D{{"getFreeMonitoringStatus", 1}}).Decode(&actual)
				require.NoError(t, err)

				actualStatus, ok := actual.Map()["state"]
				require.True(t, ok)

				if actualStatus == "disabled" {
					if v, ok := actual.Map()["debug"]; ok {
						debug, ok := v.(bson.D)
						require.True(t, ok)
						actualStatus, ok = debug.Map()["state"]
						require.True(t, ok)
					}
				}

				assert.Equal(t, tc.expectedStatus, actualStatus)
			}
		})
	}
}

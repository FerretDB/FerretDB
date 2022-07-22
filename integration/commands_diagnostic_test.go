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
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

func TestCommandsDiagnosticGetLog(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	var actual bson.D
	err := s.TargetCollection.Database().RunCommand(s.Ctx, bson.D{{"getLog", "startupWarnings"}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	t.Log(m)

	assert.Equal(t, float64(1), m["ok"])
	assert.Equal(t, []string{"totalLinesWritten", "log", "ok"}, CollectKeys(t, actual))

	assert.IsType(t, int32(0), m["totalLinesWritten"])
}

func TestCommandsDiagnosticHostInfo(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"hostInfo", 42}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	t.Log(m)

	assert.Equal(t, float64(1), m["ok"])
	assert.Equal(t, []string{"system", "os", "extra", "ok"}, CollectKeys(t, actual))

	os := m["os"].(bson.D)
	assert.Equal(t, []string{"type", "name", "version"}, CollectKeys(t, os))

	if runtime.GOOS == "linux" {
		require.NotEmpty(t, os.Map()["name"], "os name should not be empty")
		require.NotEmpty(t, os.Map()["version"], "os version should not be empty")
	}

	system := m["system"].(bson.D)
	keys := CollectKeys(t, system)
	assert.Contains(t, keys, "currentTime")
	assert.Contains(t, keys, "hostname")
	assert.Contains(t, keys, "cpuAddrSize")
	assert.Contains(t, keys, "numCores")
	assert.Contains(t, keys, "cpuArch")
}

func TestCommandsDiagnosticListCommands(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"listCommands", 42}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	t.Log(m)

	assert.Equal(t, float64(1), m["ok"])
	assert.Equal(t, []string{"commands", "ok"}, CollectKeys(t, actual))

	commands := m["commands"].(bson.D)
	listCommands := commands.Map()["listCommands"].(bson.D)
	assert.NotEmpty(t, listCommands.Map()["help"].(string))
}

func TestCommandsDiagnosticConnectionStatus(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"connectionStatus", "*"}}).Decode(&actual)
	require.NoError(t, err)

	ok := actual.Map()["ok"]

	assert.Equal(t, float64(1), ok)
}

func TestCommandsDiagnosticExplain(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)
	dbName := testutil.DatabaseName(t)

	hostname, err := os.Hostname()
	require.NoError(t, err)
	expected := bson.D{
		{"serverInfo", bson.D{
			{"host", hostname}, // w/o port for simplicity
			{"version", version.MongoDBVersion},
			{"gitVersion", version.Get().Commit},
			{"ferretdbVersion", version.Get().Version},
		}},
		{"ok", int32(1)},
	}

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
	}{
		"count": {
			command: bson.D{
				{
					"explain", bson.D{
						{"count", testutil.CollectionName(t)},
						{"query", bson.D{{"value", bson.D{{"$type", "array"}}}}},
					},
				},
				{"verbosity", "queryPlanner"},
			},
		},
		"find": {
			command: bson.D{
				{
					"explain", bson.D{
						{"find", testutil.CollectionName(t)},
						{"query", bson.D{{"value", bson.D{{"$type", "array"}}}}},
					},
				},
				{"verbosity", "queryPlanner"},
			},
			err: &mongo.CommandError{
				Code:    40415, // (Location40415) 9 FailedToParse
				Message: "BSON field 'FindCommandRequest.query' is an unknown field.",
				Name:    "Location40415",
			},
		},
		"findAndModify": {
			command: bson.D{
				{
					"explain", bson.D{
						{"query", bson.D{{
							"$and",
							bson.A{
								bson.D{{"value", bson.D{{"$gt", 0}}}},
								bson.D{{"value", bson.D{{"$lt", 0}}}},
							},
						}}},
						{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
						{"upsert", true},
					},
				},
				{"verbosity", "queryPlanner"},
			},
			err: &mongo.CommandError{
				Code:    59, // (Location40415) 9 FailedToParse
				Message: "Explain failed due to unknown command: query",
				Name:    "CommandNotFound",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err = collection.Database().RunCommand(ctx, tc.command).Decode(&actual)

			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			actualD := ConvertDocument(t, actual)
			expectedD := ConvertDocument(t, expected)
			commandD := ConvertDocument(t, tc.command)

			explainDoc := must.NotFail(commandD.Get("explain")).(*types.Document)
			must.NoError(explainDoc.Set("$db", dbName))
			must.NoError(expectedD.Set("command", explainDoc))

			require.NotEmpty(t, must.NotFail(actualD.Get("queryPlanner")))
			require.NotEmpty(t, must.NotFail(actualD.Get("serverInfo")))

			t.Logf("expected %#v", must.NotFail(expectedD.Get("command")))
			t.Logf("actual %#v", must.NotFail(actualD.Get("command")))
			// t.FailNow()
		})
	}
}

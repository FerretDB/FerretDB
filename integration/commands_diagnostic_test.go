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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCommandsDiagnosticGetLog(t *testing.T) {
	t.Parallel()
	res := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	ctx, collection := res.Ctx, res.Collection

	for name, tc := range map[string]struct {
		command  bson.D
		expected map[string]any
		err      *mongo.CommandError
		alt      string
	}{
		"Asterisk": {
			command: bson.D{{"getLog", "*"}},
			expected: map[string]any{
				"names": bson.A(bson.A{"global", "startupWarnings"}),
				"ok":    float64(1),
			},
		},
		"Global": {
			command: bson.D{{"getLog", "global"}},
			expected: map[string]any{
				"totalLinesWritten": int64(1024),
				"log":               bson.A{},
				"ok":                float64(1),
			},
		},
		"StartupWarnings": {
			command: bson.D{{"getLog", "startupWarnings"}},
			expected: map[string]any{
				"totalLinesWritten": int64(1024),
				"log":               bson.A{},
				"ok":                float64(1),
			},
		},
		"NonExistentName": {
			command: bson.D{{"getLog", "nonExistentName"}},
			err: &mongo.CommandError{
				Code:    96,
				Name:    "OperationFailed",
				Message: `No log named 'nonExistentName'`,
			},
			alt: `no RecentEntries named: nonExistentName`,
		},
		"Nil": {
			command: bson.D{{"getLog", nil}},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: `BSON field 'getLog.getLog' is missing but a required field`,
			},
		},
		"Array": {
			command: bson.D{{"getLog", bson.A{}}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'getLog.getLog' is the wrong type 'array', expected type 'string'`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			if err != nil {
				AssertEqualAltError(t, *tc.err, tc.alt, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			k := CollectKeys(t, actual)

			for key, item := range tc.expected {
				assert.Contains(t, k, key)
				if key != "log" && key != "totalLinesWritten" {
					assert.Equal(t, m[key], item)
				}
			}
		})
	}
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
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		Providers: []shareddata.Provider{shareddata.Int32s},
	})
	ctx, collection := s.Ctx, s.Collection

	isTCP := s.IsTCP(t)

	for name, tc := range map[string]struct {
		query   bson.D
		command bson.D
	}{
		"Count": {
			query:   bson.D{{"count", collection.Name()}},
			command: bson.D{{"count", collection.Name()}, {"$db", collection.Database().Name()}},
		},
		"Find": {
			query: bson.D{
				{"find", collection.Name()},
				{"filter", bson.D{{"v", bson.D{{"$gt", int32(0)}}}}},
			},
			command: bson.D{
				{"find", collection.Name()},
				{"filter", bson.D{{"v", bson.D{{"$gt", int32(0)}}}}},
				{"$db", collection.Database().Name()},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D

			err := collection.Database().RunCommand(ctx, bson.D{{"explain", tc.query}}).Decode(&actual)
			require.NoError(t, err)

			explainResult := actual.Map()

			assert.Equal(t, float64(1), explainResult["ok"])
			assert.Equal(t, "1", explainResult["explainVersion"])
			assert.Equal(t, tc.command, explainResult["command"])

			serverInfo := ConvertDocument(t, explainResult["serverInfo"].(bson.D))
			keys := serverInfo.Keys()
			values := serverInfo.Values()

			var host string
			var port int32
			var gitVersion string
			var version string

			for i, k := range keys {
				switch k {
				case "host":
					host = values[i].(string)
				case "port":
					port = values[i].(int32)
				case "gitVersion":
					gitVersion = values[i].(string)
				case "version":
					version = values[i].(string)
				}
			}

			assert.NotEmpty(t, host)

			// only check port number on TCP, not on Unix socket
			if isTCP {
				assert.NotEmpty(t, port)
			}

			assert.NotEmpty(t, gitVersion)
			assert.Regexp(t, `^6\.0\.`, version)

			assert.NotEmpty(t, explainResult["queryPlanner"])
			assert.IsType(t, bson.D{}, explainResult["queryPlanner"])
		})
	}
}

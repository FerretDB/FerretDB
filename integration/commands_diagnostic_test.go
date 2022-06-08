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

	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// startFields := []zap.Field{
// 	zap.String("version", info.Version),
// 	zap.String("commit", info.Commit),
// 	zap.String("branch", info.Branch),
// 	zap.Bool("dirty", info.Dirty),
// }
// for _, k := range info.BuildEnvironment.Keys() {
// 	v := must.NotFail(info.BuildEnvironment.Get(k))
// 	startFields = append(startFields, zap.Any(k, v))
// }
// logger.Info("Starting FerretDB "+info.Version+"...", startFields...)
/////////////////////////////////
// startFields := zap.Field{
// 	Key:       "strt",
// 	Type:      1,
// 	Integer:   1,
// 	String:    "tst",
// 	Interface: nil,
// }
// logger.Debug("Starting test", startFields)

// collections := types.MakeArray(len(names))
// 	for _, n := range names {
// 		d := must.NotFail(types.NewDocument(
// 			"name", n,
// 			"type", "collection",
// 		))
// 		if err = collections.Append(d); err != nil {
// 			return nil, lazyerrors.Error(err)
// 		}
// 	}

// 	var reply wire.OpMsg
// 	must.NoError(reply.SetSections(wire.OpMsgSection{
// 		Documents: []*types.Document{must.NotFail(types.NewDocument(
// 			"cursor", must.NotFail(types.NewDocument(
// 				"id", int64(0),
// 				"ns", db+".$cmd.listCollections",
// 				"firstBatch", collections,
// 			)),
// 			"ok", float64(1),
// 		))},
// 	}))

func TestCommandsDiagnosticGetLog(t *testing.T) {

	// bson.D{{"getLog", "*"}}
	// bson.A(bson.A{"global", "startupWarnings"})
	//
	// bson.D{{"getLog", "global"}
	// {  	"totalLinesWritten" : 18307.0,
	// 	"log" : [],
	// 	"ok" : 1.0
	// }
	//
	// bson.D{getLog:"non-existent name"}
	// { 	"ok" : 0.0,
	// 	"errmsg" : "no RamLog named: non-existent name"
	// }
	//
	// bson.D{getLog:NaN}
	// { 		"ok" : 0.0,
	// 	"errmsg" : "Argument to getLog must be of type String; found nan.0 of type double",
	// 	"code" : 14.0,
	// 	"codeName" : "TypeMismatch"
	// }

	//	t.Parallel()
	logging.Setup(zap.DebugLevel)
	logger := zap.L()

	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})
	logger.Info("TST")

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
				"totalLinesWritten": 0,
				"log":               bson.A{},
				"ok":                float64(1),
			},
		},

		"NonExistentName": {
			command: bson.D{{"getLog", "nonExistentName"}},
			err: &mongo.CommandError{
				Code:    0,
				Message: `no RamLog named: nonExistentName`,
			},
			alt: "no RamLog named: nonExistentName",
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
	ctx, collection := setup(t)

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
	ctx, collection := setup(t)

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
	ctx, collection := setup(t)

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"connectionStatus", "*"}}).Decode(&actual)
	require.NoError(t, err)

	ok := actual.Map()["ok"]

	assert.Equal(t, float64(1), ok)
}

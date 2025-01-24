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
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil/teststress"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestConnectionStatusCommand(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"connectionStatus", "*"}}).Decode(&actual)
	require.NoError(t, err)

	ok := actual.Map()["ok"]

	assert.Equal(t, float64(1), ok)
}

func TestExplainCommand(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		Providers: []shareddata.Provider{shareddata.Int32s},
	})
	ctx, collection := s.Ctx, s.Collection

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D

			err := collection.Database().RunCommand(ctx, bson.D{{"explain", tc.query}}).Decode(&actual)
			require.NoError(t, err)

			var actualComparable bson.D

			for _, elem := range actual {
				switch elem.Key {
				case "explainVersion":
					assert.True(t, elem.Value == "1" || elem.Value == "2") // explainVersion 1 and 2 are in use for different methods on Mongo 7
					actualComparable = append(actualComparable, bson.E{"explainVersion", "1"})

				case "serverInfo":
					var elemComparable bson.D

					for _, subElem := range elem.Value.(bson.D) {
						switch subElem.Key {
						case "host":
							assert.IsType(t, "", subElem.Value)
							assert.NotEmpty(t, subElem.Value)
							elemComparable = append(elemComparable, bson.E{"host", ""})

						case "port":
							assert.IsType(t, int32(0), subElem.Value)
							elemComparable = append(elemComparable, bson.E{"port", int32(0)})

						case "version":
							assert.IsType(t, "", subElem.Value)
							assert.Regexp(t, `^7\.0\.`, subElem.Value)
							elemComparable = append(elemComparable, bson.E{"version", "7.0.0"})

						case "gitVersion":
							assert.IsType(t, "", subElem.Value)
							assert.NotEmpty(t, subElem.Value)
							elemComparable = append(elemComparable, bson.E{"gitVersion", ""})

						case "ferretdb":
							// ignore

						default:
							elemComparable = append(elemComparable, subElem)
						}
					}

					actualComparable = append(actualComparable, bson.E{"serverInfo", elemComparable})

				case "queryPlanner":
					assert.NotEmpty(t, elem.Value)
					assert.IsType(t, bson.D{}, elem.Value)
					actualComparable = append(actualComparable, bson.E{elem.Key, bson.D{}})

				case "serverParameters", "executionStats":
					// ignore

				default:
					actualComparable = append(actualComparable, elem)
				}
			}

			expected := bson.D{
				{"explainVersion", "1"},
				{"queryPlanner", bson.D{}},
				{"command", tc.command},
				{"serverInfo", bson.D{
					{"host", ""},
					{"port", int32(0)},
					{"version", "7.0.0"},
					{"gitVersion", ""},
				}},
				{"ok", float64(1)},
			}

			AssertEqualDocuments(t, expected, actualComparable)
		})
	}
}

func TestGetLogCommand(t *testing.T) {
	t.Parallel()
	res := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	ctx, collection := res.Ctx, res.Collection

	for name, tc := range map[string]struct {
		command bson.D // required, command to run

		expectedComparable bson.D
		err                *mongo.CommandError // optional, expected error from MongoDB
		altMessage         string              // optional, alternative error message for FerretDB, ignored if empty
	}{
		"Asterisk": {
			command: bson.D{{"getLog", "*"}},
			expectedComparable: bson.D{
				{"names", bson.A{"global", "startupWarnings"}},
				{"ok", float64(1)},
			},
		},
		"Global": {
			command: bson.D{{"getLog", "global"}},
			expectedComparable: bson.D{
				{Key: "totalLinesWritten"},
				{Key: "log"},
				{"ok", float64(1)},
			},
		},
		"StartupWarnings": {
			command: bson.D{{"getLog", "startupWarnings"}},
			expectedComparable: bson.D{
				{Key: "totalLinesWritten"},
				{Key: "log"},
				{"ok", float64(1)},
			},
		},
		"NonExistentName": {
			command: bson.D{{"getLog", "nonExistentName"}},
			err: &mongo.CommandError{
				Code:    96,
				Name:    "OperationFailed",
				Message: `No log named 'nonExistentName'`,
			},
			altMessage: `no RecentEntries named: nonExistentName`,
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			var resComparable bson.D

			for _, elem := range res {
				switch elem.Key {
				case "log":
					resComparable = append(resComparable, bson.E{Key: elem.Key})
					log, ok := elem.Value.(bson.A)
					assert.True(t, ok)
					assert.Positive(t, len(log))
				case "totalLinesWritten":
					assert.IsType(t, int32(0), elem.Value)
					assert.Positive(t, elem.Value)
					resComparable = append(resComparable, bson.E{Key: elem.Key})
				default:
					resComparable = append(resComparable, elem)
				}
			}

			AssertEqualDocuments(t, tc.expectedComparable, resComparable)
		})
	}
}

func TestHostInfoCommand(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var a bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"hostInfo", 42}}).Decode(&a)
	require.NoError(t, err)

	var actualComparable bson.D

	for _, elem := range a {
		switch elem.Key {
		case "system":
			var elemForComparsion bson.D

			for _, subElem := range elem.Value.(bson.D) {
				switch subElem.Key {
				case "currentTime":
					assert.IsType(t, primitive.DateTime(0), subElem.Value)
					elemForComparsion = append(elemForComparsion, bson.E{"currentTime", primitive.DateTime(0)})

				case "hostname", "cpuArch":
					assert.IsType(t, "", subElem.Value)
					elemForComparsion = append(elemForComparsion, bson.E{subElem.Key, ""})

				case "cpuAddrSize", "numCores":
					assert.IsType(t, int32(0), subElem.Value)
					elemForComparsion = append(elemForComparsion, bson.E{subElem.Key, int32(0)})

				case "numPhysicalCores", "numCpuSockets", "numNumaNodes", "numaEnabled", "memSizeMB", "memLimitMB":
					// not implemented in FerretDB, do nothing
					// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/587

				default:
					elemForComparsion = append(elemForComparsion, subElem)
				}
			}

			actualComparable = append(actualComparable, bson.E{"system", elemForComparsion})

		case "os":
			var elemForComparsion bson.D

			for _, subElem := range elem.Value.(bson.D) {
				switch subElem.Key {
				case "type", "name", "version":
					assert.IsType(t, "", subElem.Value)
					elemForComparsion = append(elemForComparsion, bson.E{subElem.Key, ""})

				default:
					elemForComparsion = append(elemForComparsion, subElem)
				}
			}

			actualComparable = append(actualComparable, bson.E{"os", elemForComparsion})

		case "extra":
			assert.IsType(t, bson.D{}, elem.Value)
			actualComparable = append(actualComparable, bson.E{"extra", bson.D{}})

		default:
			actualComparable = append(actualComparable, elem)
		}
	}

	expected := bson.D{
		{"system", bson.D{
			{"currentTime", primitive.DateTime(0)},
			{"hostname", ""},
			{"cpuAddrSize", int32(0)},
			{"numCores", int32(0)},
			{"cpuArch", ""},
		}},
		{"os", bson.D{
			{"type", ""},
			{"name", ""},
			{"version", ""},
		}},
		{"extra", bson.D{}},
		{"ok", float64(1)},
	}

	AssertEqualDocuments(t, expected, actualComparable)
}

func TestListCommandsCommand(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var res bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"listCommands", 42}}).Decode(&res)
	require.NoError(t, err)

	var actualForComparsion bson.D

	for _, v := range res {
		switch v.Key {
		case "commands":
			var commandsComparable bson.D

			for _, command := range v.Value.(bson.D) {
				var commandComparable bson.D

				switch command.Key {
				case "listCommands":
					for _, subV := range command.Value.(bson.D) {
						switch subV.Key {
						case "help":
							assert.IsType(t, "", subV.Value)
							commandComparable = append(commandComparable, bson.E{"help", bson.D{}})

						case "requiresAuth", "secondaryOk", "adminOnly", "apiVersions", "deprecatedApiVersions":
							// not implemented in FerretDB, do nothing
							// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/588

						default:
							commandComparable = append(commandComparable, subV)
						}
					}

					commandsComparable = append(commandsComparable, bson.E{command.Key, commandComparable})

				default:
					// do nothing, we only check "listCommands" command for now
				}
			}

			actualForComparsion = append(actualForComparsion, bson.E{"commands", commandsComparable})

		default:
			actualForComparsion = append(actualForComparsion, v)
		}
	}

	expected := bson.D{
		{"commands", bson.D{{"listCommands", bson.D{{"help", bson.D{}}}}}},
		{"ok", float64(1)},
	}

	AssertEqualDocuments(t, expected, actualForComparsion)
}

func TestValidateCommand(t *testing.T) {
	t.Parallel()

	t.Run("Basic", func(tt *testing.T) {
		tt.Parallel()

		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1015")

		ctx, collection := setup.Setup(tt, shareddata.Doubles)

		var res bson.D
		command := bson.D{{"validate", collection.Name()}}
		err := collection.Database().RunCommand(ctx, command).Decode(&res)
		require.NoError(t, err)

		var uuid primitive.Binary

		for _, elem := range res {
			if elem.Key != "uuid" {
				continue
			}

			uuid = elem.Value.(primitive.Binary)
			require.Equal(t, bson.TypeBinaryUUID, uuid.Subtype)
			require.Equal(t, 16, len(uuid.Data))

			break
		}

		expected := bson.D{
			{"ns", "TestValidateCommand-Basic.TestValidateCommand-Basic"},
			{"uuid", uuid},
			{"nInvalidDocuments", int32(0)},
			{"nNonCompliantDocuments", int32(0)},
			{"nrecords", int32(25)},
			{"nIndexes", int32(1)},
			{"keysPerIndex", bson.D{{"_id_", int32(25)}}},
			{"indexDetails", bson.D{{"_id_", bson.D{{"valid", true}}}}},
			{"valid", true},
			{"repaired", false},
			{"warnings", bson.A{}},
			{"errors", bson.A{}},
			{"extraIndexEntries", bson.A{}},
			{"missingIndexEntries", bson.A{}},
			{"corruptRecords", bson.A{}},
			{"ok", float64(1)},
		}

		AssertEqualDocuments(t, expected, res)
	})

	t.Run("TwoIndexes", func(tt *testing.T) {
		tt.Parallel()

		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1016")

		ctx, collection := setup.Setup(tt, shareddata.Doubles)

		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"a", 1}}})
		require.NoError(t, err)

		var res bson.D
		command := bson.D{{"validate", collection.Name()}}
		err = collection.Database().RunCommand(ctx, command).Decode(&res)
		require.NoError(t, err)

		var uuid primitive.Binary

		for _, elem := range res {
			if elem.Key != "uuid" {
				continue
			}

			uuid = elem.Value.(primitive.Binary)
			require.Equal(t, bson.TypeBinaryUUID, uuid.Subtype)
			require.Equal(t, 16, len(uuid.Data))

			break
		}

		expected := bson.D{
			{"ns", "TestValidateCommand-TwoIndexes.TestValidateCommand-TwoIndexes"},
			{"uuid", uuid},
			{"nInvalidDocuments", int32(0)},
			{"nNonCompliantDocuments", int32(0)},
			{"nrecords", int32(25)},
			{"nIndexes", int32(2)},
			{"keysPerIndex", bson.D{{"_id_", int32(25)}, {"a_1", int32(25)}}},
			{"indexDetails", bson.D{
				{"_id_", bson.D{{"valid", true}}},
				{"a_1", bson.D{{"valid", true}}},
			}},
			{"valid", true},
			{"repaired", false},
			{"warnings", bson.A{}},
			{"errors", bson.A{}},
			{"extraIndexEntries", bson.A{}},
			{"missingIndexEntries", bson.A{}},
			{"corruptRecords", bson.A{}},
			{"ok", float64(1)},
		}

		AssertEqualDocuments(t, expected, res)
	})
}

func TestValidateCommandError(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		command bson.D

		err              *mongo.CommandError
		failsForFerretDB string
	}{
		"InvalidTypeDocument": {
			command: bson.D{{"validate", bson.D{}}},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type object",
			},
		},
		"NonExistentCollection": {
			command: bson.D{{"validate", "nonExistentCollection"}},
			err: &mongo.CommandError{
				Code:    26,
				Name:    "NamespaceNotFound",
				Message: "Collection 'TestValidateCommandError-NonExistentCollection.nonExistentCollection' does not exist to validate.",
			},
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			ctx, collection := setup.Setup(t, shareddata.Doubles)

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(res)

			assert.Nil(t, res)
			AssertEqualCommandError(t, *tc.err, err)
		})
	}
}

func TestWhatsMyURICommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	collection1 := s.Collection
	databaseName := s.Collection.Database().Name()
	collectionName := s.Collection.Name()

	// only check port number on TCP connection, no need to check on Unix domain socket
	isUnix := s.IsUnixSocket(t)

	// setup second client connection to check that `whatsmyuri` returns different ports
	client2, err := mongo.Connect(s.Ctx, options.Client().ApplyURI(s.MongoDBURI))
	require.NoError(t, err)

	defer client2.Disconnect(s.Ctx)

	collection2 := client2.Database(databaseName).Collection(collectionName)

	var ports []string

	for _, collection := range []*mongo.Collection{collection1, collection2} {
		var res bson.D
		command := bson.D{{"whatsmyuri", int32(1)}}
		err = collection.Database().RunCommand(s.Ctx, command).Decode(&res)
		require.NoError(t, err)

		var actualComparable bson.D

		for _, field := range res {
			switch field.Key {
			case "you":
				you := field.Value.(string)

				if !isUnix {
					// record ports to compare that they are not equal for two different clients.
					var port string
					_, port, err = net.SplitHostPort(you)
					require.NoError(t, err)
					assert.NotEmpty(t, port)
					ports = append(ports, port)
				}

				actualComparable = append(actualComparable, bson.E{"you", ""})

			default:
				actualComparable = append(actualComparable, field)
			}
		}

		expected := bson.D{
			{"you", ""},
			{"ok", float64(1)},
		}

		AssertEqualDocuments(t, expected, actualComparable)
	}

	if !isUnix {
		require.Equal(t, 2, len(ports))
		assert.NotEqual(t, ports[0], ports[1])
	}
}

// TestWhatsMyURICommandSingleConn tests that SingleConn behaves like advertised.
func TestWhatsMyURICommandSingleConn(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{SingleConn: true})

	collection1 := s.Collection
	databaseName := s.Collection.Database().Name()
	collectionName := s.Collection.Name()

	t.Run("SameClientStress", func(t *testing.T) {
		t.Parallel()

		ports := make(chan string, 10)

		teststress.StressN(t, len(ports), func(ready chan<- struct{}, start <-chan struct{}) {
			ready <- struct{}{}
			<-start

			var res bson.D
			err := collection1.Database().RunCommand(s.Ctx, bson.D{{"whatsmyuri", int32(1)}}).Decode(&res)
			require.NoError(t, err)

			var actualComparable bson.D

			for _, field := range res {
				switch field.Key {
				case "you":
					var port string
					_, port, err = net.SplitHostPort(field.Value.(string))
					require.NoError(t, err)
					assert.NotEmpty(t, port)
					ports <- port

					actualComparable = append(actualComparable, bson.E{"you", ""})

				default:
					actualComparable = append(actualComparable, field)
				}
			}

			expected := bson.D{
				{"you", ""},
				{"ok", float64(1)},
			}

			AssertEqualDocuments(t, expected, actualComparable)
		})

		close(ports)

		firstPort := <-ports
		for port := range ports {
			require.Equal(t, firstPort, port, "expected same client to use the same port")
		}
	})

	t.Run("DifferentClient", func(t *testing.T) {
		t.Parallel()

		u, err := url.Parse(s.MongoDBURI)
		require.NoError(t, err)

		client2, err := mongo.Connect(s.Ctx, options.Client().ApplyURI(u.String()))
		require.NoError(t, err)

		defer client2.Disconnect(s.Ctx)

		collection2 := client2.Database(databaseName).Collection(collectionName)

		var ports []string

		for _, collection := range []*mongo.Collection{collection1, collection2} {
			var res bson.D
			err := collection.Database().RunCommand(s.Ctx, bson.D{{"whatsmyuri", int32(1)}}).Decode(&res)
			require.NoError(t, err)

			var actualComparable bson.D

			for _, field := range res {
				switch field.Key {
				case "you":
					var port string
					_, port, err = net.SplitHostPort(field.Value.(string))
					require.NoError(t, err)
					assert.NotEmpty(t, port)
					ports = append(ports, port)

					actualComparable = append(actualComparable, bson.E{"you", ""})

				default:
					actualComparable = append(actualComparable, field)
				}
			}

			expected := bson.D{
				{"you", ""},
				{"ok", float64(1)},
			}

			AssertEqualDocuments(t, expected, actualComparable)
		}

		require.Equal(t, 2, len(ports))
		assert.NotEqual(t, ports[0], ports[1])
	})
}

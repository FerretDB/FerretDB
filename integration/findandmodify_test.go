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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestFindAndModifySimple(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command  bson.D
		response bson.D
	}{
		"EmptyQueryRemove": {
			command: bson.D{
				{"query", bson.D{}},
				{"remove", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}}},
				{"value", bson.D{{"_id", "array"}, {"v", bson.A{int32(42)}}}},
				{"ok", float64(1)},
			},
		},
		"NewDoubleNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-smallest"}}},
				{"update", bson.D{{"_id", "double-smallest"}, {"v", int32(43)}}},
				{"new", float64(42)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "double-smallest"}, {"v", int32(43)}}},
				{"ok", float64(1)},
			},
		},
		"NewDoubleZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-zero"}}},
				{"update", bson.D{{"_id", "double-zero"}, {"v", 43.0}}},
				{"new", float64(0)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "double-zero"}, {"v", 0.0}}},
				{"ok", float64(1)},
			},
		},
		"NewDoubleNaN": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-zero"}}},
				{"update", bson.D{{"_id", "double-zero"}, {"v", 43.0}}},
				{"new", math.NaN()},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "double-zero"}, {"v", float64(43)}}},
				{"ok", float64(1)},
			},
		},
		"NewIntNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", int32(11)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"ok", float64(1)},
			},
		},
		"NewIntZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32-zero"}}},
				{"update", bson.D{{"_id", "int32-zero"}, {"v", int32(43)}}},
				{"new", int32(0)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int32-zero"}, {"v", int32(0)}}},
				{"ok", float64(1)},
			},
		},
		"NewLongNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
				{"new", int64(11)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
				{"ok", float64(1)},
			},
		},
		"NewLongZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64-zero"}}},
				{"update", bson.D{{"_id", "int64-zero"}, {"v", int64(43)}}},
				{"new", int64(0)},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64-zero"}, {"v", int64(0)}}},
				{"ok", float64(1)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			command := bson.D{{"findAndModify", collection.Name()}}
			command = append(command, tc.command...)
			if command.Map()["sort"] == nil {
				command = append(command, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
			}

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.response, actual)
		})
	}
}

func TestFindAndModifyEmptyCollectionName(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		err        *mongo.CommandError
		altMessage string
	}{
		"EmptyCollectionName": {
			err: &mongo.CommandError{
				Code:    73,
				Message: "Invalid namespace specified 'testfindandmodifyemptycollectionname_emptycollectionname.'",
				Name:    "InvalidNamespace",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, bson.D{{"findAndModify", ""}}).Decode(&actual)

			AssertEqualError(t, *tc.err, err)
		})
	}
}

func TestFindAndModifyErrors(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command    bson.D
		err        *mongo.CommandError
		altMessage string
	}{
		"NotEnoughParameters": {
			err: &mongo.CommandError{
				Code:    9,
				Message: "Either an update or remove=true must be specified",
				Name:    "FailedToParse",
			},
		},
		"UpdateAndRemove": {
			command: bson.D{
				{"update", bson.D{}},
				{"remove", true},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Cannot specify both an update and remove=true",
			},
		},
		"UpsertAndRemove": {
			command: bson.D{
				{"upsert", true},
				{"remove", true},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Cannot specify both upsert=true and remove=true ",
			},
			altMessage: "Cannot specify both upsert=true and remove=true",
		},
		"NewAndRemove": {
			command: bson.D{
				{"new", true},
				{"remove", true},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Cannot specify both new=true and remove=true; 'remove' always returns the deleted document",
			},
		},
		"BadSortType": {
			command: bson.D{
				{"update", bson.D{}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.sort' is the wrong type 'string', expected type 'object'",
			},
			altMessage: "BSON field 'sort' is the wrong type 'string', expected type 'object'",
		},
		"BadRemoveType": {
			command: bson.D{
				{"query", bson.D{}},
				{"remove", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.remove' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'remove' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		},
		"BadUpdateType": {
			command: bson.D{
				{"query", bson.D{}},
				{"update", "123"},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Update argument must be either an object or an array",
			},
		},
		"BadNewType": {
			command: bson.D{
				{"query", bson.D{}},
				{"new", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.new' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'new' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		},
		"BadUpsertType": {
			command: bson.D{
				{"query", bson.D{}},
				{"upsert", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.upsert' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'upsert' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			command := bson.D{{"findAndModify", collection.Name()}}
			command = append(command, tc.command...)
			if command.Map()["sort"] == nil {
				command = append(command, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
			}

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)

			AssertEqualAltError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestFindAndModifyUpdate(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command           bson.D
		query             any
		update            bson.D
		response          bson.D
		err               *mongo.CommandError
		skipUpdateCheck   bool
		returnNewDocument bool
	}{
		"Replace": {
			query: bson.D{{"_id", "int64"}},
			command: bson.D{
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
			},
			update: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64"}, {"v", int64(42)}}},
				{"ok", float64(1)},
			},
		},
		"ReplaceWithoutID": {
			query: bson.D{{"_id", "int64"}},
			command: bson.D{
				{"update", bson.D{{"v", int64(43)}}},
			},
			update: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64"}, {"v", int64(42)}}},
				{"ok", float64(1)},
			},
		},
		"ReplaceReturnNew": {
			query: bson.D{{"_id", "int32"}},
			command: bson.D{
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", true},
			},
			update: bson.D{{"_id", "int32"}, {"v", int32(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"ok", float64(1)},
			},
		},
		"UpdateNotExistedIdInQuery": {
			query: bson.D{{"_id", "no-such-id"}},
			command: bson.D{
				{"update", bson.D{{"v", int32(43)}}},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(0)}, {"updatedExisting", false}}},
				{"ok", float64(1)},
			},
		},
		"UpdateNotExistedIdNotInQuery": {
			query: bson.D{{"$and", bson.A{
				bson.D{{"v", bson.D{{"$gt", 0}}}},
				bson.D{{"v", bson.D{{"$lt", 0}}}},
			}}},
			command: bson.D{
				{"update", bson.D{{"v", int32(43)}}},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(0)}, {"updatedExisting", false}}},
				{"ok", float64(1)},
			},
		},
		"UpdateOperatorSet": {
			query: bson.D{{"_id", "int64"}},
			command: bson.D{
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
			},
			update: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64"}, {"v", int64(42)}}},
				{"ok", float64(1)},
			},
		},
		"UpdateOperatorSetReturnNew": {
			query: bson.D{{"_id", "int64"}},
			command: bson.D{
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
				{"new", true},
			},
			update: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
				{"ok", float64(1)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			command := bson.D{{"findAndModify", collection.Name()}, {"query", tc.query}}
			command = append(command, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, float64(1), m["ok"])

			AssertEqualDocuments(t, tc.response, actual)

			if tc.update != nil {
				err = collection.FindOne(ctx, tc.query).Decode(&actual)
				if tc.err != nil {
					AssertEqualError(t, *tc.err, err)
					return
				}
				require.NoError(t, err)

				AssertEqualDocuments(t, tc.update, actual)
			}
		})
	}
}

func TestFindAndModifyUpsert(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command  bson.D
		response bson.D
	}{
		"Upsert": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", true},
				}},
				{"value", bson.D{{"_id", "double"}, {"v", 42.13}}},
				{"ok", float64(1)},
			},
		},
		"UpsertNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", true},
				}},
				{"value", bson.D{{"_id", "double"}, {"v", 43.13}}},
				{"ok", float64(1)},
			},
		},
		"UpsertNoSuchDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", false},
					{"upserted", "no-such-doc"},
				}},
				{"value", bson.D{{"_id", "no-such-doc"}, {"v", 43.13}}},
				{"ok", float64(1)},
			},
		},
		"UpsertNoSuchReplaceDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", false},
					{"upserted", "no-such-doc"},
				}},
				{"value", bson.D{{"_id", "no-such-doc"}, {"v", 43.13}}},
				{"ok", float64(1)},
			},
		},
		"UpsertReplace": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", true},
				}},
				{"value", bson.D{{"_id", "double"}, {"v", 42.13}}},
				{"ok", float64(1)},
			},
		},
		"UpsertReplaceReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", true},
				}},
				{"value", bson.D{{"_id", "double"}, {"v", 43.13}}},
				{"ok", float64(1)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			command := append(bson.D{{"findAndModify", collection.Name()}}, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, float64(1), m["ok"])

			AssertEqualDocuments(t, tc.response, actual)
		})
	}
}

func TestFindAndModifyUpsertComplex(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command         bson.D
		lastErrorObject bson.D
	}{
		"UpsertNoSuchDocumentNoIdInQuery": {
			command: bson.D{
				{"query", bson.D{{
					"$and",
					bson.A{
						bson.D{{"v", bson.D{{"$gt", 0}}}},
						bson.D{{"v", bson.D{{"$lt", 0}}}},
					},
				}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			command := append(bson.D{{"findAndModify", collection.Name()}}, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, float64(1), m["ok"])

			leb, ok := m["lastErrorObject"].(bson.D)
			if !ok {
				t.Fatal(actual)
			}

			for _, v := range leb {
				if v.Key == "upserted" {
					continue
				}
				assert.Equal(t, tc.lastErrorObject.Map()[v.Key], v.Value)
			}
		})
	}
}

func TestFindAndModifyRemove(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		command  bson.D
		response bson.D
	}{
		"Remove": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"remove", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
				}},
				{"value", bson.D{{"_id", "double"}, {"v", 42.13}}},
				{"ok", float64(1)},
			},
		},
		"RemoveEmptyQueryResult": {
			command: bson.D{
				{
					"query",
					bson.D{{
						"$and",
						bson.A{
							bson.D{{"v", bson.D{{"$gt", 0}}}},
							bson.D{{"v", bson.D{{"$lt", 0}}}},
						},
					}},
				},
				{"remove", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(0)}}},
				{"ok", float64(1)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars)

			command := append(bson.D{{"findAndModify", collection.Name()}}, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, float64(1), m["ok"])

			AssertEqualDocuments(t, tc.response, actual)
		})
	}
}

func TestFindAndModifyBadMaxTimeMSType(t *testing.T) {
	setup.SkipForTigris(t) // FindAndModify is not implemented for Tigris yet

	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		command    bson.D
		err        *mongo.CommandError
		altMessage string
	}{
		"BadMaxTimeMSTypeStringFindAndModify": {
			command: bson.D{
				{"findAndModify", collection.Name()},
				{"maxTimeMS", "string"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualAltError(t, *tc.err, tc.altMessage, err)
		})
	}
}

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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestFindAndModifySimple(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any
		response bson.D
		err      *mongo.CommandError
	}{
		"EmptyCollectionName": {
			command: bson.D{{"findAndModify", ""}},
			err: &mongo.CommandError{
				Code:    73,
				Message: "Invalid namespace specified 'testfindandmodifysimple.'",
				Name:    "InvalidNamespace",
			},
		},
		"EmptyQueryRemove": {
			command: bson.D{
				{"findAndModify", collection.Name()},
				{"query", bson.D{}},
				{"remove", true},
			},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}}},
				{
					"value",
					bson.D{
						{"_id", "binary"},
						{"value", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}},
					},
				},
				{"ok", 1.0},
			},
		},
		"NotEnoughParameters": {
			command: bson.D{{"findAndModify", collection.Name()}},
			err: &mongo.CommandError{
				Code:    9,
				Message: "Either an update or remove=true must be specified",
				Name:    "FailedToParse",
			},
		},
		"BadSortType": {
			command: bson.D{
				{"findAndModify", collection.Name()},
				{"update", bson.D{}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.sort' is the wrong type 'string', expected type 'object'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.response, actual)
		})
	}
}

func TestFindAndModifyUpdate(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		query             bson.D
		update            bson.D
		response          bson.D
		err               *mongo.CommandError
		skipUpdateCheck   bool
		returnNewDocument bool
	}{
		"Replace": {
			query:  bson.D{{"_id", "int32"}},
			update: bson.D{{"_id", "int32"}, {"value", int32(43)}},
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int32"}, {"value", int32(42)}}},
				{"ok", 1.0},
			},
		},
		"ReplaceReturnNew": {
			query:             bson.D{{"_id", "int32"}},
			update:            bson.D{{"_id", "int32"}, {"value", int32(43)}},
			returnNewDocument: true,
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(1)}, {"updatedExisting", true}}},
				{"value", bson.D{{"_id", "int32"}, {"value", int32(43)}}},
				{"ok", 1.0},
			},
		},
		"UpdateNotExisted": {
			query:           bson.D{{"_id", "no-such-id"}},
			update:          bson.D{{"_id", "int32"}, {"value", int32(43)}},
			skipUpdateCheck: true,
			response: bson.D{
				{"lastErrorObject", bson.D{{"n", int32(0)}, {"updatedExisting", false}}},
				{"ok", 1.0},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx,
				bson.D{
					{"findAndModify", collection.Name()},
					{"query", tc.query},
					{"update", tc.update},
					{"new", tc.returnNewDocument},
				},
			).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, 1.0, m["ok"])

			AssertEqualDocuments(t, tc.response, actual)

			if tc.skipUpdateCheck {
				return
			}

			err = collection.FindOne(ctx, tc.query).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.update, actual)
		})
	}
}

func TestFindAndModifyUpsert(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		query             bson.D
		update            bson.D
		response          bson.D
		err               *mongo.CommandError
		skipUpdateCheck   bool
		returnNewDocument bool
	}{
		"Upsert": {
			query:             bson.D{{"_id", "int33"}},
			update:            bson.D{{"_id", "int33"}, {"value", int32(43)}},
			returnNewDocument: true,
			response: bson.D{
				{"lastErrorObject", bson.D{
					{"n", int32(1)},
					{"updatedExisting", false},
					{"upserted", "int33"},
				}},
				{"value", bson.D{{"_id", "int33"}, {"value", int32(43)}}},
				{"ok", 1.0},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx,
				bson.D{
					{"findAndModify", collection.Name()},
					{"query", tc.query},
					{"update", tc.update},
					{"new", tc.returnNewDocument},
					{"upsert", true},
				},
			).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, 1.0, m["ok"])

			AssertEqualDocuments(t, tc.response, actual)

			if tc.skipUpdateCheck {
				return
			}

			err = collection.FindOne(ctx, tc.query).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.update, actual)
		})
	}
}

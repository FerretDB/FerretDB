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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestUpdateFieldSet(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		id       string
		update   bson.D
		expected bson.D
		err      *mongo.WriteError
		stat     *mongo.UpdateResult
		alt      string
	}{
		"ArrayNil": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			expected: bson.D{{"_id", "string"}, {"v", bson.A{nil}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"SetSameValueInt": {
			id:       "int32",
			update:   bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			expected: bson.D{{"_id", "int32"}, {"v", int32(42)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}

func TestUpdateFieldErrors(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		id         string
		update     bson.D
		err        *mongo.WriteError
		altMessage string
	}{
		"SetUnsuitableValueType": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.foo", "foo"}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "cannot use the part (v of v.foo) to traverse the element " +
					"({v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]})",
			},
			altMessage: "cannot use path 'v.foo' to traverse the document",
		},
		"RenameEmptyFieldName": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"", "v"}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: "An empty update path is not valid.",
			},
		},
		"RenameEmptyPath": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.", "v"}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: "The update path 'v.' contains an empty field name, which is not allowed.",
			},
		},
		"RenameArrayInvalidIndex": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.-1", "f"}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "cannot use the part (v of v.-1) to traverse the element " +
					"({v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]})",
			},
			altMessage: "cannot use path 'v.-1' to traverse the document",
		},
		"RenameUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.0.foo.0.bar.z", "f"}}}},
			err: &mongo.WriteError{
				Code:    28,
				Message: "cannot use the part (bar of v.0.foo.0.bar.z) to traverse the element ({bar: \"hello\"})",
			},
			altMessage: "types.getByPath: can't access string by path \"z\"",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.ArrayDocuments)

			_, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			AssertEqualAltWriteError(t, *tc.err, tc.altMessage, err)
		})
	}
}

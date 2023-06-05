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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestDistinctErrors(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any                // required
		collName any                // optional
		filter   any                // required
		err      mongo.CommandError // required
	}{
		"EmptyFilter": {
			command: "a",
			filter:  nil,
		},
		"StringFilter": {
			command: "a",
			filter:  "a",
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.query' is the wrong type 'string', expected type 'object'",
			},
		},
		"EmptyCollection": {
			command:  "a",
			filter:   bson.D{},
			collName: "",
			err: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestDistinctErrors.'",
			},
		},
		"CollectionTypeObject": {
			command:  "a",
			filter:   bson.D{},
			collName: bson.D{},
			err: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type object",
			},
		},
		"WrongTypeObject": {
			command: bson.D{},
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'object', expected type 'string'",
			},
		},
		"WrongTypeArray": {
			command: bson.A{},
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'array', expected type 'string'",
			},
		},
		"WrongTypeNumber": {
			command: int32(1),
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'int', expected type 'string'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var collName any = coll.Name()
			if tc.collName != nil {
				collName = tc.collName
			}

			command := bson.D{{"distinct", collName}, {"key", tc.command}, {"query", tc.filter}}

			res := coll.Database().RunCommand(ctx, command)
			if res.Err() != nil {
				AssertEqualCommandError(t, tc.err, res.Err())

				return
			}

			require.NoError(t, res.Err(), "expected no error")
		})
	}
}

func TestDistinctDuplicates(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)

	for name, tc := range map[string]struct {
		docs     []bson.D
		key      string
		expected []any
	}{
		"IntFirst": {
			docs: []bson.D{
				{{"v", int64(42)}},
				{{"v", float64(42)}},
				{{"v", "42"}},
			},
			key:      "v",
			expected: []any{int64(42), "42"},
		},
		"Fff": {
			docs: []bson.D{
				{{"v", int64(42)}},
				{{"v", float64(42)}},
				{{"v", int32(43)}},
				{{"v", "42"}},
			},
			key:      "v",
			expected: []any{int64(42), "42"},
		},
		"FloatFirst": {
			docs: []bson.D{
				{{"v", int64(42)}},
				{{"v", int32(42)}},
				{{"v", "42"}},
			},
			key:      "v",
			expected: []any{int64(42), "42"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var docs = make([]any, len(tc.docs))

			for i, doc := range tc.docs {
				docs[i] = doc
			}

			_, err := coll.InsertMany(ctx, docs)
			require.NoError(t, err)

			distinct, err := coll.Distinct(ctx, tc.key, bson.D{})
			require.NoError(t, err)

			assert.Equal(t, len(tc.expected), len(distinct), distinct)

			for i, value := range distinct {
				expectedValue := tc.expected[i]

				if value != expectedValue {
					switch value.(type) {
					case int64, int32, float64:
						assert.EqualValues(t, expectedValue, value)
					default:
						require.Equal(t, tc.expected, distinct)
					}
				}
			}
		})
	}
}

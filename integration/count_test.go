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
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestQueryCountErrors(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, nil)

	ctx, collection := s.Ctx, s.Collection

	for name, tc := range map[string]struct {
		value    any
		expected mongo.CommandError
	}{
		"CollectionDocument": {
			value: bson.D{
				{"count", bson.D{}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type object",
			},
		},
		"CollectionArray": {
			value: bson.D{
				{"count", primitive.A{}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type array",
			},
		},
		"CollectionDouble": {
			value: bson.D{
				{"count", 3.14},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type double",
			},
		},
		"CollectionBinary": {
			value: bson.D{
				{"count", primitive.Binary{}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type binData",
			},
		},
		"CollectionObjectID": {
			value: bson.D{
				{"count", primitive.ObjectID{}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type objectId",
			},
		},
		"CollectionBool": {
			value: bson.D{
				{"count", true},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type bool",
			},
		},
		"CollectionDate": {
			value: bson.D{
				{"count", time.Now()},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type date",
			},
		},
		"CollectionNull": {
			value: bson.D{
				{"count", nil},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type null",
			},
		},
		"CollectionRegex": {
			value: bson.D{
				{"count", primitive.Regex{Pattern: "/foo/"}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type regex",
			},
		},
		"CollectionInt": {
			value: bson.D{
				{"count", int32(42)},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type int",
			},
		},
		"CollectionTimestamp": {
			value: bson.D{
				{"count", primitive.Timestamp{}},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type timestamp",
			},
		},
		"CollectionLong": {
			value: bson.D{
				{"count", int64(42)},
				{"query", bson.D{}},
			},
			expected: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type long",
			},
		},
		"QueryInt": {
			value: bson.D{
				{"count", "collection"},
				{"query", int32(42)},
			},
			expected: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'count.query' is the wrong type 'int', expected types 'object'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.value).Decode(&actual)
			require.Error(t, err)

			AssertEqualCommandError(t, tc.expected, err)
		})
	}
}

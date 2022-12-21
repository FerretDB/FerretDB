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
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

func TestQueryBadCountType(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
	}{
		"Document": {
			command: bson.D{
				{"count", bson.D{}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type object",
			},
		},
		"Array": {
			command: bson.D{
				{"count", primitive.A{}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type array",
			},
		},
		"Double": {
			command: bson.D{
				{"count", 3.14},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"DoubleWhole": {
			command: bson.D{
				{"count", 42.0},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"Binary": {
			command: bson.D{
				{"count", primitive.Binary{}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type binData",
			},
		},
		"ObjectID": {
			command: bson.D{
				{"count", primitive.ObjectID{}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type objectId",
			},
		},
		"Bool": {
			command: bson.D{
				{"count", true},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type bool",
			},
		},
		"Date": {
			command: bson.D{
				{"count", time.Now()},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type date",
			},
		},
		"Null": {
			command: bson.D{
				{"count", nil},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
		},
		"Regex": {
			command: bson.D{
				{"count", primitive.Regex{Pattern: "/foo/"}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type regex",
			},
		},
		"Int": {
			command: bson.D{
				{"count", int32(42)},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
		},
		"Timestamp": {
			command: bson.D{
				{"count", primitive.Timestamp{}},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type timestamp",
			},
		},
		"Long": {
			command: bson.D{
				{"count", int64(42)},
				{"query", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type long",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualError(t, *tc.err, err)
		})
	}
}

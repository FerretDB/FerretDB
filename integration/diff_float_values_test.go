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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDiffFloatValues(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			command bson.D
			err     mongo.CommandError
		}{
			"NaN": {
				command: bson.D{{"_id", "1"}, {"foo", math.NaN()}},
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { insert: "TestDiffFloatValues", ordered: true, ` +
						`$db: "TestDiffFloatValues", documents: [ { _id: "1", foo: nan.0 } ] } with: NaN is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.command)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				AssertEqualCommandError(t, tc.err, err)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			filter bson.D
			update bson.D
			opts   *options.UpdateOptions
			err    mongo.CommandError
		}{
			"NaN": {
				filter: bson.D{{"_id", "2"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "TestDiffFloatValues", ` +
						`ordered: true, $db: "TestDiffFloatValues", updates: [ { q: { _id: "2" }, ` +
						`u: { $set: { foo: nan.0 } } } ] } with: NaN is not supported`,
				},
			},
			"NaNWithUpsert": {
				filter: bson.D{{"_id", "3"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				opts:   options.Update().SetUpsert(true),
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { update: "TestDiffFloatValues", ` +
						`ordered: true, $db: "TestDiffFloatValues", updates: [ { q: { _id: "3" }, ` +
						`u: { $set: { foo: nan.0 } }, upsert: true } ] } with: NaN is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.filter)
				if err != nil {
					t.Fatal(err)
				}

				_, err = collection.UpdateOne(ctx, tc.filter, tc.update, tc.opts)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				AssertEqualCommandError(t, tc.err, err)
			})
		}
	})

	t.Run("FindAndModify", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			filter bson.D
			update bson.D
			err    mongo.CommandError
		}{
			"NaN": {
				filter: bson.D{{"_id", "4"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { findAndModify: "TestDiffFloatValues", ` +
						`query: { _id: "4" }, update: { $set: { foo: nan.0 } }, $db: "TestDiffFloatValues" } with: NaN ` +
						`is not supported`,
				},
			},
		} {
			name, tc := name, tc

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.filter)
				if err != nil {
					t.Fatal(err)
				}

				err = collection.FindOneAndUpdate(ctx, tc.filter, tc.update).Err()

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				AssertEqualCommandError(t, tc.err, err)
			})
		}
	})
}

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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestDiffDocumentValidation(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct { //nolint:vet // use only for testing
			doc bson.D

			err error
		}{
			"DollarSign": {
				doc: bson.D{{"$foo", "bar"}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "$foo" (key must not start with '$' sign)`,
				}}},
			},
			"DotSign": {
				doc: bson.D{{"foo.bar", "baz"}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "foo.bar" (key must not contain '.' sign)`,
				}}},
			},
			"Infinity": {
				doc: bson.D{{"foo", math.Inf(1)}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": +Inf } (infinity values are not allowed)`,
				}}},
			},
			"NegativeInfinity": {
				doc: bson.D{{"foo", math.Inf(-1)}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": -Inf } (infinity values are not allowed)`,
				}}},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.doc)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				assert.Equal(t, tc.err, UnsetRaw(t, err))
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct { //nolint:vet // used only for testing
			filter bson.D
			update bson.D
			opts   *options.UpdateOptions

			werr *mongo.WriteError
		}{
			"DotSign": {
				filter: bson.D{},
				update: bson.D{{"$set", bson.D{{"foo", bson.D{{"bar.baz", "qaz"}}}}}},
				werr: &mongo.WriteError{
					Code:    2,
					Message: `invalid key: "bar.baz" (key must not contain '.' sign)`,
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.UpdateOne(ctx, tc.filter, tc.update, tc.opts)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				AssertEqualWriteError(t, *tc.werr, err)
			})
		}
	})
}

func TestDiffDocumentValidationNaN(t *testing.T) {
	t.Parallel()

	t.Run("InsertNaN", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t, shareddata.Scalars)

		_, err := collection.InsertOne(ctx, bson.D{{"_id", "nan"}, {"foo", math.NaN()}})

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		require.ErrorContains(t, err, "socket was unexpectedly closed")
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct { //nolint:vet // used only for testing
			filter bson.D
			update bson.D
			opts   *options.UpdateOptions
		}{
			"NaN": {
				filter: bson.D{{"_id", "2"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
			},
			"NaNWithUpsert": {
				filter: bson.D{{"_id", "3"}},
				update: bson.D{{"$set", bson.D{{"foo", math.NaN()}}}},
				opts:   options.Update().SetUpsert(true),
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx, collection := setup.Setup(t, shareddata.Scalars)

				_, err := collection.UpdateOne(ctx, tc.filter, tc.update, tc.opts)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				require.ErrorContains(t, err, "socket was unexpectedly closed")
			})
		}
	})

	t.Run("FindAndModifyNaN", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t, shareddata.Scalars)

		filter := bson.D{{"_id", "4"}}

		_, err := collection.InsertOne(ctx, filter)
		if err != nil {
			t.Fatal(err)
		}

		err = collection.FindOneAndUpdate(ctx, filter, bson.D{{"$set", bson.D{{"foo", math.NaN()}}}}).Err()

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		require.ErrorContains(t, err, "socket was unexpectedly closed")
	})
}

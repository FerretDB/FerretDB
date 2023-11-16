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

func TestDiffNestedArrays(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		_, err := collection.InsertOne(ctx, bson.D{{"foo", bson.A{bson.A{"bar"}}}})

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		expected := mongo.WriteError{
			Index:   0,
			Code:    2,
			Message: `invalid value: { "foo": [ [ "bar" ] ] } (nested arrays are not supported)`,
		}
		AssertEqualWriteError(t, expected, err)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		_, err := collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"foo", bson.A{bson.A{"bar"}}}}}})

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		expected := mongo.WriteError{
			Code:    2,
			Message: `invalid value: { "foo": [ [ "bar" ] ] } (nested arrays are not supported)`,
		}
		AssertEqualWriteError(t, expected, err)
	})
}

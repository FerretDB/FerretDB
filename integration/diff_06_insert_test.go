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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDiffInsertDuplicateKeys(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	doc := bson.D{{"_id", "duplicate_keys"}, {"foo", "bar"}, {"foo", "baz"}}
	_, err := collection.InsertOne(ctx, doc)

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.WriteError{
		Index:   0,
		Code:    2,
		Message: `invalid key: "foo" (duplicate keys are not allowed)`,
	}
	AssertEqualWriteError(t, expected, err)
}

func TestDiffInsertObjectIDHexString(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	hex := "000102030405060708091011"

	objID, err := primitive.ObjectIDFromHex(hex)
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", objID},
	})
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", hex},
	})

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.WriteError{
		Index:   0,
		Code:    11000,
		Message: `E11000 duplicate key error collection: TestDiffInsertObjectIDHexString.TestDiffInsertObjectIDHexString`,
	}
	AssertEqualWriteError(t, expected, err)
}

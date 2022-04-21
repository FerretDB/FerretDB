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
)

func TestCommandsAdministrationCreateDropList(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	db := collection.Database()
	name := collection.Name()

	names, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.Contains(t, names, name)

	err = collection.Drop(ctx)
	require.NoError(t, err)

	// error is consumed by the driver
	err = collection.Drop(ctx)
	require.NoError(t, err)
	err = db.Collection(name).Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var actual bson.D
	err = db.RunCommand(ctx, bson.D{{"drop", name}}).Decode(&actual)
	expectedErr := mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: `ns not found`,
	}
	AssertEqualError(t, expectedErr, err)

	names, err = db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.NotContains(t, names, name)

	err = db.CreateCollection(ctx, name)
	require.NoError(t, err)

	err = db.CreateCollection(ctx, name)
	expectedErr = mongo.CommandError{
		Code: 48,
		Name: "NamespaceExists",
		Message: `Collection already exists. ` +
			`NS: testcommandsadministrationcreatedroplist.testcommandsadministrationcreatedroplist`,
	}
	AssertEqualError(t, expectedErr, err)

	names, err = db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.Contains(t, names, name)
}

func TestCommandsAdministrationGetParameter(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"getParameter", "*"}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	t.Log(m)

	assert.Equal(t, 1.0, m["ok"])

	keys := CollectKeys(t, actual)
	assert.Contains(t, keys, "quiet")
	assert.Equal(t, false, m["quiet"])
}

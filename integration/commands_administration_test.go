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

// assertDatabases checks that expected and actual listDatabases results are similar.
func assertDatabases(t *testing.T, expected, actual mongo.ListDatabasesResult) {
	t.Helper()

	var sizeSum int64
	assert.NotZero(t, actual.TotalSize)

	require.Len(t, actual.Databases, len(expected.Databases), "%+v", actual)
	for i, a := range actual.Databases {
		e := expected.Databases[i]
		require.Equal(t, e.Name, a.Name)

		assert.Zero(t, e.SizeOnDisk)

		if a.Empty {
			assert.Zero(t, a.SizeOnDisk, "%+v", a)
			continue
		}

		// to make comparison easier
		assert.NotZero(t, a.SizeOnDisk, "%+v", a)
		sizeSum += a.SizeOnDisk
		actual.Databases[i].SizeOnDisk = 0
	}

	// That's not true for PostgreSQL, where a sum of `pg_total_relation_size` result for all schemas
	// is not equal to `pg_database_size` for the whole database.
	// assert.Equal(t, sizeSum, actual.TotalSize)
	actual.TotalSize = sizeSum

	expected.TotalSize = sizeSum
	assert.Equal(t, expected, actual)
}

//nolint:paralleltest // we test a global list of databases
func TestCommandsAdministrationCreateDropListDatabases(t *testing.T) {
	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})
	client := collection.Database().Client()
	name := collection.Name()

	// drop remnants of the previous failed run
	_ = client.Database(name).Drop(ctx)

	filter := bson.D{{
		"name", bson.D{{
			"$in", bson.A{"monila", "values", "admin", name},
		}},
	}}
	names, err := client.ListDatabaseNames(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, []string{"admin", "monila", "values"}, names)

	actual, err := client.ListDatabases(ctx, filter)
	require.NoError(t, err)

	expectedBefore := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name: "admin",
		}, {
			Name: "monila",
		}, {
			Name: "values",
		}},
	}
	assertDatabases(t, expectedBefore, actual)

	// there is no explicit command to create database, so create collection instead
	err = client.Database(name).CreateCollection(ctx, name)
	require.NoError(t, err)

	actual, err = client.ListDatabases(ctx, filter)
	require.NoError(t, err)

	expectedAfter := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name: "admin",
		}, {
			Name: "monila",
		}, {
			Name: name,
		}, {
			Name: "values",
		}},
	}
	assertDatabases(t, expectedAfter, actual)

	err = client.Database(name).Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var res bson.D
	err = client.Database(name).RunCommand(ctx, bson.D{{"dropDatabase", 1}}).Decode(&res)
	require.NoError(t, err)

	actual, err = client.ListDatabases(ctx, filter)
	require.NoError(t, err)
	assertDatabases(t, expectedBefore, actual)
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

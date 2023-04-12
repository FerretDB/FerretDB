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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// TestCreateCompat tests collection creation compatibility for the cases that are not covered by tests setup.
func TestCreateCompat(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{}, // collections are not needed for this test
		AddNonExistentCollection: true,
	})

	// We expect to have only one (non-existent) collection as the result of setup.
	require.Len(t, s.TargetCollections, 1)
	require.Len(t, s.CompatCollections, 1)

	targetDB := s.TargetCollections[0].Database()
	compatDB := s.CompatCollections[0].Database()

	// Test collection creation in non-existent database.
	err := targetDB.Drop(s.Ctx)
	require.NoError(t, err)

	err = compatDB.Drop(s.Ctx)
	require.NoError(t, err)

	collName := "in-non-existent-db"

	// schema in case of Tigris.
	schema := fmt.Sprintf(`{
				"title": "%s",
				"description": "Create Collection In Non-Existent Database",
				"primary_key": ["_id"],
				"properties": {
					"_id": {"type": "string"}
				}
			}`, collName,
	)

	// Set $tigrisSchemaString for tigris only.
	opts := options.CreateCollection()
	if setup.IsTigris(t) {
		opts.SetValidator(bson.D{{"$tigrisSchemaString", schema}})
	}

	targetErr := targetDB.CreateCollection(s.Ctx, collName, opts)
	compatErr := compatDB.CreateCollection(s.Ctx, collName)
	require.Equal(t, targetErr, compatErr)

	targetNames, targetErr := targetDB.ListCollectionNames(s.Ctx, bson.D{})
	compatNames, compatErr := compatDB.ListCollectionNames(s.Ctx, bson.D{})
	require.NoError(t, compatErr)
	require.Equal(t, targetErr, compatErr)
	require.Equal(t, targetNames, compatNames)
}

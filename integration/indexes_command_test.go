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

// TestListIndexesNonExistentNS tests that the listIndexes command returns a particular error
// when the namespace (either database or collection) does not exist.
func TestListIndexesNonExistentNS(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		Providers: []shareddata.Provider{shareddata.Composites},
	})
	ctx, collection := s.Ctx, s.Collection

	// Calling driver's method collection.Database().Collection("nonexistent").Indexes().List(ctx)
	// doesn't return an error for non-existent namespaces.
	// So that we should use RunCommand to check the behaviour.
	res := collection.Database().RunCommand(ctx, bson.D{{"listIndexes", "nonexistentColl"}})
	err := res.Err()

	AssertEqualCommandError(t, mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: "ns does not exist: TestListIndexesNonExistentNS.nonexistentColl",
	}, err)

	// Drop database and check that the error is correct.
	require.NoError(t, collection.Database().Drop(ctx))
	res = collection.Database().RunCommand(ctx, bson.D{{"listIndexes", collection.Name()}})
	err = res.Err()

	AssertEqualCommandError(t, mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: "ns does not exist: TestListIndexesNonExistentNS." + collection.Name(),
	}, err)
}

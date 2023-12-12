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

package users

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	// https://www.mongodb.com/docs/manual/reference/command/createUser/

	_ = collection.Database().RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	})

	var res bson.D
	err := collection.Database().RunCommand(ctx, bson.D{
		{"createUser", "testuser"},
		{"roles", bson.A{}},
		{"pwd", "password"},
	}).Decode(&res)

	require.NoError(t, err)

	actual := integration.ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := integration.ConvertDocument(t, bson.D{{"ok", float64(1)}})

	testutil.AssertEqual(t, expected, actual)

	var rec bson.D
	err = collection.Database().Collection("system.users").FindOne(ctx, bson.D{{"user", "testuser"}}).Decode(&rec)
	require.NoError(t, err)

	actualRecorded := integration.ConvertDocument(t, rec)
	expectedRec := integration.ConvertDocument(t, bson.D{{"ok", float64(1)}})

	testutil.AssertEqual(t, expectedRec, actualRecorded)

}

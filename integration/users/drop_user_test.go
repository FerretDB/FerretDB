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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDropUser(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{SetupUser: true})
	ctx, db := s.Ctx, s.Collection.Database()

	err := db.RunCommand(ctx, bson.D{
		{"createUser", "a_user"},
		{"roles", bson.A{}},
		{"pwd", "password"},
	}).Err()
	require.NoError(t, err)

	testCases := map[string]struct { //nolint:vet // for readability
		payload    bson.D
		err        *mongo.CommandError
		altMessage string
		expected   bson.D
	}{
		"NotFound": {
			payload: bson.D{
				{"dropUser", "not_found_user"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: "User 'not_found_user@TestDropUser' not found",
			},
		},
		"Success": {
			payload: bson.D{
				{"dropUser", "a_user"},
			},
			expected: bson.D{
				{"ok", float64(1)},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res bson.D
			err := db.RunCommand(ctx, tc.payload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			actual := integration.ConvertDocument(t, res)
			actual.Remove("$clusterTime")
			actual.Remove("operationTime")

			expected := integration.ConvertDocument(t, tc.expected)
			testutil.AssertEqual(t, expected, actual)

			payload := integration.ConvertDocument(t, tc.payload)
			username := must.NotFail(payload.Get("dropUser")).(string)
			assertUserNotFound(ctx, t, db, username)
		})
	}
}

// assertUserNotFound checks it the user doesn't exist in the admin.system.users collection.
func assertUserNotFound(ctx context.Context, t testing.TB, db *mongo.Database, username string) {
	t.Helper()

	var res bson.D
	err := db.RunCommand(ctx, bson.D{{"usersInfo", bson.A{
		bson.D{
			{"user", username},
			{"db", db.Name()},
		},
	}}}).Decode(&res)

	require.NoError(t, err)

	actual := integration.ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := integration.ConvertDocument(t, bson.D{
		{"users", bson.A{}},
		{"ok", float64(1)},
	})
	testutil.AssertEqual(t, expected, actual)
}

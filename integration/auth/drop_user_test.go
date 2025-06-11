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

package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestDropUserCommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", "a_user"}})

	errcmd := db.RunCommand(ctx, bson.D{ // avoid data race with and shadowing of err in parallel subtests below
		{"createUser", "a_user"},
		{"roles", bson.A{}},
		{"pwd", "password"},
	}).Err()
	require.NoError(t, errcmd)

	testCases := map[string]struct { //nolint:vet // for readability
		username string

		expected         bson.D
		err              *mongo.CommandError
		altMessage       string
		failsForFerretDB string
	}{
		"NotFound": {
			username: "not_found_user",
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: fmt.Sprintf("User 'not_found_user@%s' not found", db.Name()),
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/942",
		},
		"Success": {
			username: "a_user",
			expected: bson.D{
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(t, tc.failsForFerretDB)
			}

			tt.Parallel()

			var res bson.D
			err := db.RunCommand(ctx, bson.D{{"dropUser", tc.username}}).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)
			integration.AssertEqualDocuments(t, tc.expected, res)

			err = db.RunCommand(ctx, bson.D{{"usersInfo", bson.A{
				bson.D{
					{"user", tc.username},
					{"db", db.Name()},
				},
			}}}).Decode(&res)
			require.NoError(t, err)

			expected := bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			}
			integration.AssertEqualDocuments(t, expected, res)
		})
	}
}

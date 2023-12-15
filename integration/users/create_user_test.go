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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_ = collection.Database().RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	})

	db := collection.Database()

	err := db.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: "should_already_exist"},
		{Key: "roles", Value: bson.A{}},
		{Key: "pwd", Value: "password"},
	}).Err()
	assert.NoError(t, err)

	testCases := map[string]struct {
		payload    bson.D
		err        *mongo.CommandError
		altMessage string
		expected   bson.D
		skip       string
	}{
		"AlreadyExists": {
			skip: "TODO", // FIXME

			payload: bson.D{
				{Key: "createUser", Value: "should_already_exist"},
				{Key: "roles", Value: bson.A{}},
				{Key: "pwd", Value: "password"},
			},
			err: &mongo.CommandError{
				Code:    51003,
				Name:    "Location51003",
				Message: "User \"should_already_exist@TestCreateUser\" already exists",
			},
		},
		"MissingPwdOrExternal": {
			skip: "TODO", // FIXME

			payload: bson.D{
				{Key: "createUser", Value: "mising_pwd_or_external"},
				{Key: "roles", Value: bson.A{}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db",
			},
		},
		"Success": {
			payload: bson.D{
				{Key: "createUser", Value: "success_user"},
				{Key: "roles", Value: bson.A{}},
				{Key: "pwd", Value: "password"},
			},
			expected: bson.D{
				{
					Key: "ok", Value: float64(1),
				},
			},
		},
		"WithComment": {
			skip: "TODO", // FIXME

			payload: bson.D{
				{Key: "createUser", Value: "with_comment_user"},
				{Key: "roles", Value: bson.A{}},
				{Key: "pwd", Value: "password"},
				{Key: "comment", Value: "test string comment"},
			},
			expected: bson.D{
				{
					Key: "ok", Value: float64(1),
				},
			},
		},
		"WithCommentComposite": {
			skip: "TODO", // FIXME

			payload: bson.D{
				{Key: "createUser", Value: "with_comment_composite"},
				{Key: "roles", Value: bson.A{}},
				{Key: "pwd", Value: "password"},
				{
					Key: "comment",
					Value: bson.D{
						{Key: "example", Value: "blah"},
						{
							Key: "complex",
							Value: bson.A{
								bson.D{{Key: "x", Value: "y"}},
							},
						},
					},
				},
			},
			expected: bson.D{
				{
					Key: "ok", Value: float64(1),
				},
			},
		},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

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
			assertUserExists(ctx, t, db, payload)
		})
	}
}

func assertUserExists(ctx context.Context, t testing.TB, db *mongo.Database, payload *types.Document) {
	t.Helper()

	var rec bson.D
	err := db.Collection("system.users").FindOne(ctx, bson.D{{"user", must.NotFail(payload.Get("createUser"))}}).Decode(&rec)
	require.NoError(t, err)

	actualRecorded := integration.ConvertDocument(t, rec)
	expectedRec := integration.ConvertDocument(t, bson.D{
		{"user", must.NotFail(payload.Get("createUser"))},
	}) // FIXME

	testutil.AssertEqual(t, expectedRec, actualRecorded)
	// TODO compare other data
}

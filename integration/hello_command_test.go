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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestHello(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	var res bson.D

	require.NoError(t, db.RunCommand(ctx, bson.D{
		{"hello", "1"},
	}).Decode(&res))

	actual := ConvertDocument(t, res)

	assert.Equal(t, must.NotFail(actual.Get("isWritablePrimary")), true)
	assert.Equal(t, must.NotFail(actual.Get("maxBsonObjectSize")), int32(16777216))
	assert.Equal(t, must.NotFail(actual.Get("maxMessageSizeBytes")), int32(48000000))
	assert.Equal(t, must.NotFail(actual.Get("maxMessageSizeBytes")), int32(48000000))
	assert.Equal(t, must.NotFail(actual.Get("maxWriteBatchSize")), int32(100000))
	assert.IsType(t, must.NotFail(actual.Get("localTime")), time.Time{})
	assert.IsType(t, must.NotFail(actual.Get("connectionId")), int32(1))
	assert.Equal(t, must.NotFail(actual.Get("minWireVersion")), int32(0))
	assert.Equal(t, must.NotFail(actual.Get("maxWireVersion")), int32(21))
	assert.Equal(t, must.NotFail(actual.Get("readOnly")), false)
	assert.Equal(t, must.NotFail(actual.Get("ok")), float64(1))
}

func TestHelloWithSupportedMechs(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	usersPayload := []bson.D{
		{
			{"createUser", "hello_user"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
		},
		{
			{"createUser", "hello_user_scram1"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"SCRAM-SHA-1"}},
		},
		{
			{"createUser", "hello_user_scram256"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"SCRAM-SHA-256"}},
		},
	}

	if !setup.IsMongoDB(t) {
		usersPayload = append(usersPayload, primitive.D{
			{"createUser", "hello_user_plain"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"PLAIN"}},
		})
	}

	for _, u := range usersPayload {
		require.NoError(t, db.RunCommand(ctx, u).Err())
	}

	testCases := []struct { //nolint:vet // used for test only
		username string
		db       string
		mechs    *types.Array
		err      bool
	}{
		{
			username: "not_found",
			db:       db.Name(),
		},
		{
			username: "another_db",
			db:       db.Name() + "_not_found",
		},
		{
			username: "hello_user",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256")),
		},
		{
			username: "hello_user_plain",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("PLAIN")),
		},
		{
			username: "hello_user_scram1",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-1")),
		},
		{
			username: "hello_user_scram256",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-256")),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.username, func(t *testing.T) {
			t.Parallel()

			var res bson.D

			if tc.mechs != nil && tc.mechs.Contains("PLAIN") {
				setup.SkipForMongoDB(t, "PLAIN authentication mechanism is not support by MongoDB")
			}

			err := db.RunCommand(ctx, bson.D{
				{"hello", "1"},
				{"saslSupportedMechs", tc.db + "." + tc.username},
			}).Decode(&res)

			if tc.err {
				require.Error(t, err)
			}

			actual := ConvertDocument(t, res)

			if tc.mechs != nil {
				mechanisms := must.NotFail(actual.Get("saslSupportedMechs"))
				assert.ElementsMatch(t, tc.mechs, mechanisms)
			} else {
				assert.False(t, actual.Has("saslSupportedMechs"))
			}

			assert.Equal(t, must.NotFail(actual.Get("isWritablePrimary")), true)
			assert.Equal(t, must.NotFail(actual.Get("maxBsonObjectSize")), int32(16777216))
			assert.Equal(t, must.NotFail(actual.Get("maxMessageSizeBytes")), int32(48000000))
			assert.Equal(t, must.NotFail(actual.Get("maxMessageSizeBytes")), int32(48000000))
			assert.Equal(t, must.NotFail(actual.Get("maxWriteBatchSize")), int32(100000))
			assert.IsType(t, must.NotFail(actual.Get("localTime")), time.Time{})
			assert.IsType(t, must.NotFail(actual.Get("connectionId")), int32(1))
			assert.Equal(t, must.NotFail(actual.Get("minWireVersion")), int32(0))
			assert.Equal(t, must.NotFail(actual.Get("maxWireVersion")), int32(21))
			assert.Equal(t, must.NotFail(actual.Get("readOnly")), false)
			assert.Equal(t, must.NotFail(actual.Get("ok")), float64(1))
		})
	}
}

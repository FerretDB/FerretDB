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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
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

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		SetupUser: true,
		Providers: []shareddata.Provider{shareddata.Scalars, shareddata.Composites},
	})
	ctx, db := s.Ctx, s.Collection.Database()

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

	testCases := map[string]struct { //nolint:vet // used for test only
		user  string
		mechs *types.Array

		err             *mongo.CommandError
		failsForMongoDB string
	}{
		"NotFound": {
			user: db.Name() + ".not_found",
		},
		"AnotherDB": {
			user: db.Name() + "_not_found.another_db",
		},
		"HelloUser": {
			user:  db.Name() + ".hello_user",
			mechs: must.NotFail(types.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256")),
		},
		"HelloUserPlain": {
			user:            db.Name() + ".hello_user_plain",
			mechs:           must.NotFail(types.NewArray("PLAIN")),
			failsForMongoDB: "PLAIN authentication mechanism is not support by MongoDB",
		},
		"HelloUserSCRAM1": {
			user:  db.Name() + ".hello_user_scram1",
			mechs: must.NotFail(types.NewArray("SCRAM-SHA-1")),
		},
		"HelloUserSCRAM256": {
			user:  db.Name() + ".hello_user_scram256",
			mechs: must.NotFail(types.NewArray("SCRAM-SHA-256")),
		},
		"EmptyUsername": {
			user:  db.Name() + ".",
			mechs: nil,
		},
		"MissingSeparator": {
			user: db.Name(),
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "UserName must contain a '.' separated database.user pair",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testtb.TB = tt

			if tc.failsForMongoDB != "" {
				t = setup.FailsForMongoDB(t, tc.failsForMongoDB)
			}

			var res bson.D

			err := db.RunCommand(ctx, bson.D{
				{"hello", "1"},
				{"saslSupportedMechs", tc.user},
			}).Decode(&res)

			if tc.err != nil {
				AssertEqualCommandError(t, *tc.err, err)
				return
			}

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

			if tc.mechs == nil {
				assert.False(t, actual.Has("saslSupportedMechs"))
				return
			}

			mechanisms, err := actual.Get("saslSupportedMechs")
			require.NoError(t, err)
			assert.True(t, mechanisms.(*types.Array).ContainsAll(tc.mechs))
		})
	}
}

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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestHello(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955")

	tt.Parallel()

	ctx, collection := setup.Setup(tt, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	var actual bson.D

	require.NoError(t, db.RunCommand(ctx, bson.D{
		{"hello", "1"},
	}).Decode(&actual))

	var actualComparable, actualFieldNames bson.D

	for _, field := range actual {
		switch field.Key {
		case "hosts", "setName", "topologyVersion", "setVersion", "secondary", "primary", "me", "electionId", "lastWrite":
			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
			continue
		case "connectionId":
			assert.IsType(t, int32(0), field.Value)
		case "localTime":
			assert.IsType(t, primitive.DateTime(0), field.Value)
		default:
			actualComparable = append(actualComparable, field)
		}

		actualFieldNames = append(actualFieldNames, bson.E{Key: field.Key})
	}

	expectedComparable := bson.D{
		{"isWritablePrimary", true},
		{"maxBsonObjectSize", int32(16777216)},
		{"maxMessageSizeBytes", int32(48000000)},
		{"maxWriteBatchSize", int32(100000)},
		{"logicalSessionTimeoutMinutes", int32(30)},
		{"minWireVersion", int32(0)},
		{"maxWireVersion", int32(21)},
		{"readOnly", false},
		{"ok", float64(1)},
	}
	AssertEqualDocuments(t, expectedComparable, actualComparable)

	expectedFieldNames := bson.D{
		{Key: "isWritablePrimary"},
		{Key: "maxBsonObjectSize"},
		{Key: "maxMessageSizeBytes"},
		{Key: "maxWriteBatchSize"},
		{Key: "localTime"},
		{Key: "logicalSessionTimeoutMinutes"},
		{Key: "connectionId"},
		{Key: "minWireVersion"},
		{Key: "maxWireVersion"},
		{Key: "readOnly"},
		{Key: "ok"},
	}
	AssertEqualDocuments(t, expectedFieldNames, actualFieldNames)
}

func TestHelloWithSupportedMechs(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		Providers: []shareddata.Provider{shareddata.Scalars, shareddata.Composites},
	})
	ctx, db := s.Ctx, s.Collection.Database()

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", "hello_user_scram256"}})

	require.NoError(t, db.RunCommand(ctx, bson.D{
		{"createUser", "hello_user_scram256"},
		{"roles", bson.A{}},
		{"pwd", "hello_password"},
		{"mechanisms", bson.A{"SCRAM-SHA-256"}},
	}).Err())

	testCases := map[string]struct { //nolint:vet // used for test only
		user  string
		mechs bson.A

		err              *mongo.CommandError
		failsForFerretDB string
	}{
		"NotFound": {
			user:             db.Name() + ".not_found",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955",
		},
		"AnotherDB": {
			user:             db.Name() + "_not_found.another_db",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955",
		},
		"HelloUserSCRAM256": {
			user:  db.Name() + ".hello_user_scram256",
			mechs: bson.A{"SCRAM-SHA-256"},
		},
		"EmptyUsername": {
			user:             db.Name() + ".",
			mechs:            nil,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955",
		},
		"MissingSeparator": {
			user: db.Name(),
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "UserName must contain a '.' separated database.user pair",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955",
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
			err := db.RunCommand(ctx, bson.D{
				{"hello", "1"},
				{"saslSupportedMechs", tc.user},
			}).Decode(&res)

			if tc.err != nil {
				AssertEqualCommandError(t, *tc.err, err)
				return
			}

			var actualComparable, actualFieldNames bson.D

			for _, field := range res {
				switch field.Key {
				case "hosts", "setName", "topologyVersion", "setVersion", "secondary", "primary", "me", "electionId", "lastWrite":
					// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
					continue
				case "connectionId":
					assert.IsType(t, int32(0), field.Value)
				case "localTime":
					assert.IsType(t, primitive.DateTime(0), field.Value)
				case "saslSupportedMechs":
					// the order of mechanisms is not guaranteed
					assert.ElementsMatch(t, tc.mechs, field.Value)
				default:
					actualComparable = append(actualComparable, field)
				}

				actualFieldNames = append(actualFieldNames, bson.E{Key: field.Key})
			}

			expected := bson.D{
				{"isWritablePrimary", true},
				{"maxBsonObjectSize", int32(16777216)},
				{"maxMessageSizeBytes", int32(48000000)},
				{"maxWriteBatchSize", int32(100000)},
				{"logicalSessionTimeoutMinutes", int32(30)},
				{"minWireVersion", int32(0)},
				{"maxWireVersion", int32(21)},
				{"readOnly", false},
				{"ok", float64(1)},
			}
			AssertEqualDocuments(t, expected, actualComparable)

			fieldNames := bson.D{
				{Key: "isWritablePrimary"},
				{Key: "maxBsonObjectSize"},
				{Key: "maxMessageSizeBytes"},
				{Key: "maxWriteBatchSize"},
				{Key: "localTime"},
				{Key: "logicalSessionTimeoutMinutes"},
				{Key: "connectionId"},
				{Key: "minWireVersion"},
				{Key: "maxWireVersion"},
				{Key: "readOnly"},
			}

			if tc.mechs != nil {
				fieldNames = append(fieldNames, bson.E{Key: "saslSupportedMechs"})
			}
			fieldNames = append(fieldNames, bson.E{Key: "ok"})

			AssertEqualDocuments(t, fieldNames, actualFieldNames)
		})
	}
}

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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestRolesCommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database().Client().Database("admin")

	testCases := map[string]struct { //nolint:vet // for readability
		user         string
		roles        bson.A
		authzCommand bson.D

		expected         bson.D
		err              *mongo.CommandError
		altMessage       string
		failsForFerretDB string
	}{
		"Unauthorized": {
			user: "userA",
			roles: bson.A{bson.D{
				{"role", "readAnyDatabase"},
				{"db", db.Name()},
			}},
			authzCommand: bson.D{{"updateUser", "userX"}, {"pwd", "password"}},
			err: &mongo.CommandError{
				Code: 13,
				Name: "Unauthorized",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(t, tc.failsForFerretDB)
			}

			t.Cleanup(func() {
				err := db.RunCommand(ctx, bson.D{{"dropUser", tc.user}}).Err()
				assert.NoError(t, err)
			})

			var res bson.D
			err := db.RunCommand(ctx, bson.D{
				{"createUser", tc.user},
				{"roles", tc.roles},
				{"pwd", "password"},
			}).Decode(&res)
			require.NoError(t, err)

			auth := options.Credential{
				Username: tc.user,
				Password: "password",
			}
			conn, err := mongo.Connect(ctx, options.Client().ApplyURI(s.MongoDBURI).SetAuth(auth))
			require.NoError(t, err)

			err = conn.Database(db.Name()).RunCommand(ctx, tc.authzCommand).Decode(&res)
			if tc.err != nil {
				// only compare error codes and names, not the message as it contains generated values
				//
				// (Unauthorized) not authorized on admin to execute command { createUser: "userX",
				// roles: [], pwd: "xxx", lsid: { id: UUID("85c334b9-84a5-478e-a65c-c7ff93e070b9") },
				// $clusterTime: { clusterTime: Timestamp(1754452120, 1),
				// signature: { hash: BinData(0, 2C87500872D1A4E4909A0CD6E37B60D93E40E0EC),
				// keyId: 7534627867146059783 } }, $db: "admin" }
				integration.AssertMatchesCommandError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			integration.AssertEqualDocuments(t, tc.expected, res)
		})
	}
}

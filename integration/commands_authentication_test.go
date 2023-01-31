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
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCommandsAuthenticationSASLStart(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "postgreSQL authentication test")

	t.Parallel()

	testcases := map[string]struct {
		skip                 string
		dbErr                string
		listenerTLS          bool
		clientTLS            bool
		invalidUsername      bool
		invalidPassword      bool
		invalidAuthMechanism bool
	}{
		"TLS": {
			listenerTLS: true,
			clientTLS:   true,
		},
		"TCP": {
			listenerTLS: false,
			clientTLS:   false,
		},
		"TLSWrongUsername": {
			listenerTLS:     true,
			clientTLS:       true,
			invalidUsername: true,
			dbErr:           "auth error",
		},
		"TLSWrongPassword": {
			listenerTLS:     true,
			clientTLS:       true,
			invalidPassword: true,
			dbErr:           "auth error",
			skip:            "https://github.com/FerretDB/FerretDB/pull/1857",
		},
		"NotPlainTLS": {
			listenerTLS:          true,
			clientTLS:            true,
			invalidAuthMechanism: true,
			dbErr:                "auth error",
		},
	}

	for name, tc := range testcases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				setup.SkipForPostgresWithReason(t, tc.skip)
			}

			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				Flags: map[string]any{
					"target-tls":         tc.listenerTLS,
					"target-port":        0,
					"target-unix-socket": false,
					"postgresql-url":     "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
				},
			})

			ctx := s.Ctx

			// s.MongoDBURI looks like:
			// `mongodb://username:password@127.0.0.1:35697/?authMechanism=PLAIN`.
			clientURI, err := url.Parse(s.MongoDBURI)
			require.NoError(t, err)

			user := clientURI.User.Username()
			password, _ := clientURI.User.Password()
			authMechanism := "PLAIN"

			if tc.invalidUsername {
				user = "invalid"
			}

			if tc.invalidPassword {
				password = "invalid"
			}

			if tc.invalidAuthMechanism {
				authMechanism = ""
			}

			auth := options.Credential{
				Username:      user,
				Password:      password,
				AuthMechanism: authMechanism,
			}

			clientOpts := options.Client().ApplyURI(clientURI.String())

			if tc.clientTLS {
				clientOpts.SetAuth(auth).SetTLSConfig(setup.GetClientTLSConfig(t))
			}

			// upon Connect clientURI.String() value of auth is overridden with SetAuth.
			client, err := mongo.Connect(ctx, clientOpts)
			require.NoError(t, err)

			t.Cleanup(func() {
				err = client.Disconnect(ctx)
				require.NoError(t, err)
			})

			// client calls any query to check authentication success or error.
			dbs, err := client.ListDatabaseNames(ctx, bson.D{})
			if tc.dbErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.dbErr)
				return
			}

			require.NoError(t, err)

			// expects to find databases such as "admin" and "public".
			require.NotEmpty(t, dbs)
		})
	}
}

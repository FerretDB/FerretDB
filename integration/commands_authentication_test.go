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
	"strings"
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
		skip            string
		dbErr           string
		listenerTLS     bool
		clientTLS       bool
		invalidUsername bool
		invalidPassword bool
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
			dbErr:           "role \"invalid\" does not exist",
		},
		"TLSWrongPassword": {
			listenerTLS:     true,
			clientTLS:       true,
			invalidPassword: true,
			skip:            "TODO: Expects error but connection is established successfully",
		},
		"TLSWithNonTLSClient": {
			listenerTLS: true,
			clientTLS:   false,
			skip:        "MongoDB driver attempts to reconnect until time out",
		},
	}

	for name, tc := range testcases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				setup.SkipForPostgresWithReason(t, tc.skip)
			}

			postgreSQLURL := "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1"
			port := 0
			unix := false
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				Flags: map[string]any{
					"target-tls":         tc.listenerTLS,
					"target-port":        port,
					"target-unix-socket": unix,
					"postgresql-url":     postgreSQLURL,
				},
			})

			ctx := s.Ctx

			// s.MongoDBURI looks like `mongodb://username:password@127.0.0.1:35697/?authMechanism=PLAIN`.
			// For testing invalid parameter, we replace each part of it.
			clientURI := s.MongoDBURI

			if tc.invalidUsername {
				clientURI = strings.Replace(clientURI, "username", "invalid", 1)
			}

			if tc.invalidPassword {
				clientURI = strings.Replace(clientURI, "password", "invalid", 1)
			}

			clientOpts := options.Client().ApplyURI(clientURI)

			if tc.clientTLS {
				clientOpts.SetTLSConfig(setup.GetClientTLSConfig(t))
			}

			client, err := mongo.Connect(ctx, clientOpts)
			require.NoError(t, err)

			t.Cleanup(func() {
				err = client.Disconnect(ctx)
				require.NoError(t, err)
			})

			_, err = client.ListDatabases(ctx, bson.D{})
			if tc.dbErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.dbErr)
				return
			}
		})
	}
}

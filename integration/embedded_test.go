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
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/ferretdb"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestEmbedded(t *testing.T) {
	t.Skip("https://github.com/FerretDB/FerretDB/issues/2062")

	t.Parallel()

	for name, tc := range map[string]struct {
		config *ferretdb.Config
	}{
		"TCP": {
			config: &ferretdb.Config{
				Listener: ferretdb.ListenerConfig{
					TCP: "127.0.0.1:37027",
				},
				Handler:       "pg",
				PostgreSQLURL: testutil.PostgreSQLURL(t, nil),
			},
		},
		"TLS": {
			config: &ferretdb.Config{
				Listener: ferretdb.ListenerConfig{
					TLS:         "127.0.0.1:37028",
					TLSCertFile: filepath.Join(setup.CertsRoot, "server-cert.pem"),
					TLSKeyFile:  filepath.Join(setup.CertsRoot, "server-key.pem"),
					TLSCAFile:   filepath.Join(setup.CertsRoot, "rootCA-cert.pem"),
				},
				Handler:       "pg",
				PostgreSQLURL: testutil.PostgreSQLURL(t, nil),
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := ferretdb.New(tc.config)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(testutil.Ctx(t))
			defer cancel()

			// check that Run exits on context cancel
			done := make(chan struct{})
			go func() {
				err = f.Run(ctx)
				t.Logf("Run exited with %v.", err) // result is undefined for now
				cancel()
				close(done)
			}()

			client, err := mongo.Connect(ctx, options.Client().ApplyURI(f.MongoDBURI()))
			require.NoError(t, err)

			filter := bson.D{{
				"name",
				bson.D{{
					"$not",
					bson.D{{
						"$regex",
						primitive.Regex{Pattern: "test.*"},
					}},
				}},
			}}

			var names []string
			names, err = client.ListDatabaseNames(ctx, filter)
			require.NoError(t, err)
			assert.Equal(t, []string{"admin", "public"}, names)

			require.NoError(t, client.Disconnect(ctx))

			cancel()
			<-done
		})
	}
}

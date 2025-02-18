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

package ferretdb_test

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeps(t *testing.T) {
	t.Parallel()

	var res struct {
		Deps []string `json:"Deps"`
	}
	b, err := exec.Command("go", "list", "-json").Output()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &res))

	assert.NotContains(t, res.Deps, "testing", `package "testing" should not be imported by non-testing code`)
}

/*
func Example_tcp() {
	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TCP: "127.0.0.1:17027",
		},
		Logger:        slog.With("component", "ferretdb"),
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() {
		done <- f.Run(ctx)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.
	// For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	// import "go.mongodb.org/mongo-driver/mongo/options"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri))

	cancel()

	err = <-done
	if err != nil {
		log.Fatal(err)
	}

	// Output: mongodb://127.0.0.1:17027/
}

func Example_unix() {
	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			Unix: "/tmp/ferretdb.sock",
		},
		Logger:        slog.With("component", "ferretdb"),
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() {
		done <- f.Run(ctx)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.
	// For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	// import "go.mongodb.org/mongo-driver/mongo/options"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri))

	cancel()

	err = <-done
	if err != nil {
		log.Fatal(err)
	}

	// Output: mongodb://%2Ftmp%2Fferretdb.sock/
}

func Example_tls() {
	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TLS:         "127.0.0.1:17028",
			TLSCertFile: filepath.Join(testutil.BuildCertsDir, "server-cert.pem"),
			TLSKeyFile:  filepath.Join(testutil.BuildCertsDir, "server-key.pem"),
			TLSCAFile:   filepath.Join(testutil.BuildCertsDir, "rootCA-cert.pem"),
		},
		Logger:        slog.With("component", "ferretdb"),
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() {
		done <- f.Run(ctx)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.
	// To connect to TLS listener, set TLS config.
	// For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	// import "go.mongodb.org/mongo-driver/mongo/options"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri))

	cancel()

	err = <-done
	if err != nil {
		log.Fatal(err)
	}

	// Output: mongodb://127.0.0.1:17028/?tls=true
}

func TestEmbedded(t *testing.T) {
	ctx, cancel := context.WithCancel(testutil.Ctx(t))

	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TCP: "127.0.0.1:0",
		},
		Logger:        testutil.Logger(t),
		Handler:       "postgresql",
		PostgreSQLURL: testutil.TestPostgreSQLURI(t, ctx, ""),
	})
	require.NoError(t, err)

	done := make(chan struct{})

	go func() {
		require.NoError(t, f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)

	//nolint:forbidigo // bson is required to use the driver
	_, err = client.Database(dbName).Collection(collName).InsertOne(ctx, bson.M{"foo": "bar"})
	require.NoError(t, err)

	cancel()
	<-done
}
*/

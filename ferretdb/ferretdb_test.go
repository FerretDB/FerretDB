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
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/ferretdb"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func Example_tcp() {
	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TCP: "127.0.0.1:17027",
		},
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		log.Print(f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual. For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri))

	cancel()
	<-done

	// Output: mongodb://127.0.0.1:17027/
}

func Example_unix() {
	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			Unix: "/tmp/ferretdb.sock",
		},
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		log.Print(f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.

	cancel()
	<-done

	// Output: mongodb://%2Ftmp%2Fferretdb.sock/
}

func Example_tls() {
	certPath := filepath.Join("..", "build", "certs", "server-cert.pem")
	keyPath := filepath.Join("..", "build", "certs", "server-key.pem")
	caPath := filepath.Join("..", "build", "certs", "rootCA-cert.pem")

	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TLS:         "127.0.0.1:17028",
			TLSCertFile: certPath,
			TLSKeyFile:  keyPath,
			TLSCAFile:   caPath,
		},
		Handler:       "postgresql",
		PostgreSQLURL: "postgres://127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		log.Print(f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual. To connect to TLS listener, set TLS config.
	// For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	// import "go.mongodb.org/mongo-driver/mongo/options"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri))

	cancel()
	<-done

	// Output: mongodb://127.0.0.1:17028/?tls=true
}

func TestEmbedded(t *testing.T) {
	ctx, cancel := context.WithCancel(testutil.Ctx(t))

	f, err := ferretdb.New(&ferretdb.Config{
		Listener: ferretdb.ListenerConfig{
			TCP: "127.0.0.1:0",
		},
		Handler:       "postgresql",
		PostgreSQLURL: testutil.TestPostgreSQLURI(t, ctx, "postgres://username@127.0.0.1:5432/ferretdb?search_path="),
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

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
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/FerretDB/FerretDB/v2/ferretdb"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
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

func Example() {
	f, err := ferretdb.New(&ferretdb.Config{
		PostgreSQLURL: "postgres://username:password@127.0.0.1:5432/postgres",
		ListenAddr:    "127.0.0.1:17027",
		StateDir:      ".",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		f.Run(ctx)
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.
	// For example:
	//
	// import "go.mongodb.org/mongo-driver/v2/mongo"
	// import "go.mongodb.org/mongo-driver/v2/mongo/options"
	//
	// [...]
	//
	// mongo.Connect(options.Client().ApplyURI(uri))

	cancel()
	<-done

	// Output: mongodb://127.0.0.1:17027/
}

func TestFerretDBWithCustomLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, new(slog.HandlerOptions)))

	f, err := ferretdb.New(&ferretdb.Config{
		PostgreSQLURL: testutil.PostgreSQLURL(t),
		ListenAddr:    "127.0.0.1:0",
		StateDir:      t.TempDir(),
		LogLevel:      slog.LevelWarn,
		Logger:        logger,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	done := make(chan struct{})

	go func() {
		f.Run(ctx)
		close(done)
	}()

	uri := f.MongoDBURI()
	require.NotEmpty(t, uri)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	// check that non-existent command is logged
	client.Database("admin").RunCommand(ctx, bson.D{{Key: "nonExistentCommand", Value: 1}})

	res := client.Database("admin").RunCommand(ctx, bson.D{{Key: "getLog", Value: "global"}})

	var actual bson.D
	err = res.Decode(&actual)
	require.NoError(t, err)

	require.Len(t, actual, 3)
	require.Equal(t, "log", actual[0].Key)
	require.Contains(t, actual[0].Value, "no such command: 'nonExistentCommand'")
	require.Equal(t, "ok", actual[2].Key)
	require.Equal(t, 1.0, actual[2].Value)

	err = client.Disconnect(ctx)
	require.NoError(t, err)

	cancel()
	<-done
}

func TestFerretDB(t *testing.T) {
	f, err := ferretdb.New(&ferretdb.Config{
		PostgreSQLURL: testutil.PostgreSQLURL(t),
		ListenAddr:    "127.0.0.1:0",
		StateDir:      t.TempDir(),
		LogLevel:      slog.LevelDebug,
		LogOutput:     testutil.NewLogWriter(t),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	done := make(chan struct{})

	go func() {
		f.Run(ctx)
		close(done)
	}()

	uri := f.MongoDBURI()
	require.NotEmpty(t, uri)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	res := client.Database("admin").RunCommand(ctx, bson.D{{Key: "getFreeMonitoringStatus", Value: 1}})

	var actual bson.D
	err = res.Decode(&actual)
	require.NoError(t, err)

	expected := bson.D{
		{Key: "state", Value: "undecided"},
		{Key: "message", Value: "monitoring is undecided"},
		{Key: "ok", Value: 1.0},
	}
	assert.Equal(t, expected, actual)

	err = client.Disconnect(ctx)
	require.NoError(t, err)

	cancel()
	<-done
}

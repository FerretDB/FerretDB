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
	"flag"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	startupPort = flag.String("port", "ferretdb", "port to use")

	startupOnce sync.Once
)

// setup returns test-specific context (that is cancelled when the test ends) and database client.
func setup(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()

	startupOnce.Do(func() { startup(t) })

	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	port, err := strconv.Atoi(*startupPort)
	if err != nil {
		t.Fatal(err)
	}

	// start in-process FerretDB if port is not set
	if port == 0 {
		pgPool := testutil.Pool(ctx, t, nil, logger)

		l := clientconn.NewListener(&clientconn.NewListenerOpts{
			ListenAddr: "127.0.0.1:0",
			Mode:       clientconn.NormalMode,
			PgPool:     pgPool,
			Logger:     logger.Named("listener"),
		})

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := l.Run(ctx)
			if err == nil || err == context.Canceled {
				logger.Info("Listener stopped")
			} else {
				logger.Error("Listener stopped", zap.Error(err))
			}
		}()

		port = l.Addr().(*net.TCPAddr).Port
	}

	uri := fmt.Sprintf("mongodb://127.0.0.1:%d", port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	databaseName := databaseName(t)
	db := client.Database(databaseName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		if t.Failed() {
			t.Logf("Keeping database %q for debugging.", databaseName)
		} else {
			client.Database(databaseName)
			err = db.Drop(context.Background())
			require.NoError(t, err)
		}

		err = client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	err = db.Drop(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	return ctx, db
}

// startup initializes things that should be initialized only once.
func startup(t *testing.T) {
	t.Helper()

	*startupPort = strings.ToLower(*startupPort)
	switch *startupPort {
	case "ferretdb":
		*startupPort = "0"
	case "default":
		*startupPort = "27017"
	case "mongodb":
		*startupPort = "37017"
	}

	if _, err := strconv.Atoi(*startupPort); err != nil {
		t.Fatal(err)
	}

	logging.Setup(zap.DebugLevel)

	ctx := context.Background()

	go debug.RunHandler(ctx, "127.0.0.1:0", zap.L().Named("debug"))
}

func scalarsData() map[string]any {
	return map[string]any{
		"double":                   42.13,
		"double-zero":              0.0,
		"double-negative-zero":     math.Copysign(0, -1),
		"double-max":               math.MaxFloat64,
		"double-smallest":          math.SmallestNonzeroFloat64,
		"double-positive-infinity": math.Inf(+1),
		"double-negative-infinity": math.Inf(-1),
		"double-nan":               math.NaN(),

		"string":       "foo",
		"string-empty": "",

		// no Document
		// no Array

		"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
		"binary-empty": primitive.Binary{},

		// no Undefined

		"bool-false": false,
		"bool-true":  true,

		"datetime":          time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
		"datetime-epoch":    time.Unix(0, 0),
		"datetime-year-min": time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
		"datetime-year-max": time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC),

		"null": nil,

		"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
		"regex-empty": primitive.Regex{},

		// no DBPointer
		// no JavaScript code
		// no Symbol
		// no JavaScript code w/ scope

		"int32":      int32(42),
		"int32-zero": int32(0),
		"int32-max":  int32(math.MaxInt32),
		"int32-min":  int32(math.MinInt32),

		"timestamp":   primitive.Timestamp{T: 42, I: 13},
		"timestamp-i": primitive.Timestamp{I: 1},

		"int64":      int64(42),
		"int64-zero": int64(0),
		"int64-max":  int64(math.MaxInt64),
		"int64-min":  int64(math.MinInt64),

		// no 128-bit decimal floating point (yet)

		// no Min key
		// no Max key
	}
}

func insertScalars(ctx context.Context, t *testing.T, db *mongo.Database) {
	t.Helper()

	collection := db.Collection(collectionName(t))

	for id, v := range scalarsData() {
		_, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"value", v}})
		require.NoError(t, err)
	}
}

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

// Package setup provides integration tests setup helpers.
package setup

import (
	"context"
	"flag"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/xiter"

	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

// Flags.
var (
	targetURLF     = flag.String("target-url", "", "target system's URL; if empty, in-process FerretDB is used")
	targetBackendF = flag.String("target-backend", "", "target system's backend: "+strings.Join(allBackends, ", "))

	postgreSQLURLF    = flag.String("postgresql-url", "", "in-process FerretDB: PostgreSQL URL")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "in-process FerretDB: use Unix domain socket")
	targetProxyAddrF  = flag.String("target-proxy-addr", "", "in-process FerretDB: use given proxy")

	compatURLF = flag.String("compat-url", "", "compat system's (MongoDB) URL for compatibility tests; if empty, they are skipped")

	otelTracesURLF = flag.String("otel-traces-url", "http://127.0.0.1:4318/v1/traces", "OpenTelemetry OTLP/HTTP traces endpoint URL")

	noXFailF = flag.Bool("no-xfail", false, "Disallow expected failures")

	benchDocsF = flag.Int("bench-docs", 1000, "benchmarks: number of documents to generate per iteration")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = flag.String("log-level", slog.LevelDebug.String(), "log level for tests")
)

// Other globals.
var (
	allBackends = []string{"ferretdb", "ferretdb-yugabytedb", "mongodb"}
)

// WireConn defines if and how wire client connection is established.
type WireConn int

const (
	// WireConnNoConn does not create wire client connection.
	WireConnNoConn WireConn = iota

	// WireConnNoAuth creates non-authenticated wire client connection.
	WireConnNoAuth

	// WireConnAuth creates authenticated wire client connection.
	WireConnAuth
)

// SetupOpts represents setup options.
//
// Add option to use read-only user.
// TODO https://github.com/FerretDB/FerretDB/issues/1025
type SetupOpts struct {
	ListenerOpts *ListenerOpts

	// Database to use. If empty, temporary test-specific database is created and dropped after test.
	DatabaseName string

	// Data providers. If empty, collection is not created.
	Providers []shareddata.Provider

	// Benchmark data provider. If empty, collection is not created.
	BenchmarkProvider shareddata.BenchmarkProvider

	// PoolSize ensures that MongoDB driver uses exactly this number of connections for operations
	// (not counting extra connections for monitoring that are mostly idle).
	// Zero value disables explicit pool configuration.
	PoolSize int

	// DisableOtel disable OpenTelemetry monitoring for MongoDB driver.
	DisableOtel bool

	// WireConn defines if and how wire client connection is established.
	WireConn WireConn

	// extraOptions adds or replaces query parameters in the MongoDB URI.
	// It is used by both driver and wire client connections.
	//
	// Note that wire client connection does not support many options
	// and returns an error if it encounters an unknown one.
	//
	// This field is unexported because that general API wasn't actually used (see PoolSize).
	// It might be exported again if needed.
	extraOptions url.Values
}

// SetupResult represents setup results.
type SetupResult struct {
	Ctx        context.Context
	Collection *mongo.Collection
	WireConn   *wireclient.Conn
	MongoDBURI string // without database name
}

// IsUnixSocket returns true if MongoDB URI is a Unix domain socket.
func (s *SetupResult) IsUnixSocket(tb testing.TB) bool {
	tb.Helper()

	// we can't use a regular url.Parse because
	// MongoDB really wants Unix domain socket path in the host part of the URI
	opts := options.Client().ApplyURI(s.MongoDBURI)
	res := slices.ContainsFunc(opts.Hosts, func(host string) bool {
		return strings.Contains(host, "/")
	})

	tb.Logf("IsUnixSocket: %q - %v", s.MongoDBURI, res)

	return res
}

// SetupWithOpts setups the test according to given options.
func SetupWithOpts(tb testing.TB, opts *SetupOpts) *SetupResult {
	tb.Helper()

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	setupCtx, span := otel.Tracer("").Start(ctx, "setup.SetupWithOpts")
	defer span.End()

	if opts == nil {
		opts = new(SetupOpts)
	}

	if opts.extraOptions == nil {
		opts.extraOptions = make(url.Values)
	}

	if opts.PoolSize > 0 {
		opts.extraOptions.Set("maxConnecting", strconv.Itoa(opts.PoolSize))
		opts.extraOptions.Set("maxPoolSize", strconv.Itoa(opts.PoolSize))
		opts.extraOptions.Set("minPoolSize", strconv.Itoa(opts.PoolSize))
	}

	var levelVar slog.LevelVar
	levelVar.Set(slog.LevelError)
	if *debugSetupF {
		levelVar.Set(slog.LevelDebug)
	}

	logger := testutil.LevelLogger(tb, &levelVar)

	uri := *targetURLF
	if uri == "" {
		uri = setupListener(tb, setupCtx, opts.ListenerOpts, logger)
	}

	if len(opts.extraOptions) > 0 {
		u, err := url.Parse(uri)
		require.NoError(tb, err)

		q := u.Query()

		for k, vs := range opts.extraOptions {
			for _, v := range vs {
				q.Set(k, v)
			}
		}

		u.RawQuery = q.Encode()
		uri = u.String()
		tb.Logf("URI with extra options: %s", uri)
	}

	client := setupClient(tb, setupCtx, uri, opts.DisableOtel)

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	collection := setupCollection(tb, setupCtx, client, opts)

	var conn *wireclient.Conn

	if opts.WireConn != WireConnNoConn {
		clearUri, creds, _, authMechanism, err := wireclient.Credentials(uri)
		require.NoError(tb, err)

		conn = setupWireConn(tb, setupCtx, clearUri, testutil.Logger(tb))

		if opts.WireConn == WireConnAuth {
			u, err := url.Parse(uri)
			require.NoError(tb, err)

			user := u.User.Username()
			require.NotEmpty(tb, user)

			pass, _ := u.User.Password()
			require.NotEmpty(tb, pass)

			require.NoError(tb, conn.Login(ctx, creds, "admin", authMechanism))
		}
	}

	err := levelVar.UnmarshalText([]byte(*logLevelF))
	require.NoError(tb, err)

	return &SetupResult{
		Ctx:        ctx,
		Collection: collection,
		WireConn:   conn,
		MongoDBURI: uri,
	}
}

// Setup setups a single collection for all providers, if they are present.
func Setup(tb testing.TB, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
	})
	return s.Ctx, s.Collection
}

// setupCollection setups a single collection for all providers, if they are present.
func setupCollection(tb testing.TB, ctx context.Context, client *mongo.Client, opts *SetupOpts) *mongo.Collection {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setup.setupCollection")
	defer span.End()

	var ownDatabase bool
	databaseName := opts.DatabaseName
	if databaseName == "" {
		databaseName = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	collectionName := testutil.CollectionName(tb)

	database := client.Database(databaseName)
	collection := database.Collection(collectionName)

	// drop remnants of the previous failed run
	if ownDatabase {
		cleanupDatabase(tb, ctx, database)
	} else {
		cleanupCollection(tb, ctx, collection)
	}

	// drop collection and (possibly) database unless test failed
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s.%s for debugging", databaseName, collectionName)
			return
		}

		if ownDatabase {
			cleanupDatabase(tb, ctx, database)

			return
		}

		cleanupCollection(tb, ctx, collection)
	})

	var inserted bool

	switch {
	case len(opts.Providers) > 0:
		require.Nil(tb, opts.BenchmarkProvider, "Both Providers and BenchmarkProvider were set")
		inserted = InsertProviders(tb, ctx, collection, opts.Providers...)
	case opts.BenchmarkProvider != nil:
		inserted = insertBenchmarkProvider(tb, ctx, collection, opts.BenchmarkProvider)
	}

	if len(opts.Providers) == 0 && opts.BenchmarkProvider == nil {
		tb.Logf("Collection %s.%s wasn't created because no providers were set", databaseName, collectionName)
	} else {
		require.True(tb, inserted)
	}

	return collection
}

// InsertProviders inserts documents from specified Providers into collection. It returns true if any document was inserted.
func InsertProviders(tb testing.TB, ctx context.Context, collection *mongo.Collection, providers ...shareddata.Provider) (inserted bool) {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setup.InsertProviders")
	defer span.End()

	for _, provider := range providers {
		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "provider %q", provider.Name())
		require.Len(tb, res.InsertedIDs, len(docs))
		inserted = true
	}

	return
}

// insertBenchmarkProvider inserts documents from specified BenchmarkProvider into collection.
// It returns true if any document was inserted.
//
// The function calculates the checksum of all inserted documents and compare them with provider's hash.
func insertBenchmarkProvider(tb testing.TB, ctx context.Context, collection *mongo.Collection, provider shareddata.BenchmarkProvider) (inserted bool) {
	tb.Helper()

	for docs := range xiter.Chunk(provider.Docs(), 100) {
		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err)
		require.Len(tb, res.InsertedIDs, len(docs))

		inserted = true
	}

	return
}

// cleanupCollection deletes all documents in the collection for ferretdb-yugabytedb backend
// or drops the collection for other backends.
func cleanupCollection(tb testing.TB, ctx context.Context, collection *mongo.Collection) {
	tb.Helper()

	if IsYugabyteDB(tb) {
		// dropping collection or database fails for ferretdb-yugabytedb
		// TODO https://github.com/yugabyte/yugabyte-db/issues/27698
		_, err := collection.DeleteMany(ctx, bson.D{})
		require.NoError(tb, err)

		return
	}

	err := collection.Drop(ctx)
	require.NoError(tb, err)
}

// cleanupDatabase drops all users, then deletes all collections in the database
// for ferretdb-yugabytedb backend or drops the database for other backends.
func cleanupDatabase(tb testing.TB, ctx context.Context, database *mongo.Database) {
	tb.Helper()

	err := database.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}}).Err()
	require.NoError(tb, err)

	if IsYugabyteDB(tb) {
		// dropping collection or database fails for ferretdb-yugabytedb
		// TODO https://github.com/yugabyte/yugabyte-db/issues/27698
		var collections []string
		collections, err = database.ListCollectionNames(ctx, bson.D{})
		require.NoError(tb, err)

		for _, collection := range collections {
			cleanupCollection(tb, ctx, database.Collection(collection))
		}

		return
	}

	err = database.Drop(ctx)
	require.NoError(tb, err)
}

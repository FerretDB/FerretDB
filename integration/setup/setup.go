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
	"strings"
	"time"

	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/driver"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// Flags.
var (
	targetURLF     = flag.String("target-url", "", "target system's URL; if empty, in-process FerretDB is used")
	targetBackendF = flag.String("target-backend", "", "target system's backend: '%s'"+strings.Join(allBackends, "', '"))

	targetProxyAddrF  = flag.String("target-proxy-addr", "", "in-process FerretDB: use given proxy")
	targetTLSF        = flag.Bool("target-tls", false, "in-process FerretDB: use TLS")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "in-process FerretDB: use Unix domain socket")

	postgreSQLURLF = flag.String("postgresql-url", "", "in-process FerretDB: PostgreSQL URL for 'postgresql' handler.")
	sqliteURLF     = flag.String("sqlite-url", "", "in-process FerretDB: SQLite URI for 'sqlite' handler.")
	mysqlURLF      = flag.String("mysql-url", "", "in-process FerretDB: MySQL URL for 'mysql' handler.")
	hanaURLF       = flag.String("hana-url", "", "in-process FerretDB: Hana URL for 'hana' handler.")

	batchSizeF = flag.Int("batch-size", 100, "maximum insertion batch size")

	compatURLF = flag.String("compat-url", "", "compat system's (MongoDB) URL for compatibility tests; if empty, they are skipped")

	benchDocsF = flag.Int("bench-docs", 1000, "benchmarks: number of documents to generate per iteration")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = flag.String("log-level", slog.LevelDebug.String(), "log level for tests")

	disablePushdownF = flag.Bool("disable-pushdown", false, "disable pushdown")
)

// Other globals.
var (
	allBackends = []string{"ferretdb-postgresql", "ferretdb-sqlite", "ferretdb-mysql", "ferretdb-hana", "mongodb"}
)

// SetupOpts represents setup options.
//
// Add option to use read-only user.
// TODO https://github.com/FerretDB/FerretDB/issues/1025
type SetupOpts struct {
	// Database to use. If empty, temporary test-specific database is created and dropped after test.
	DatabaseName string

	// Data providers. If empty, collection is not created.
	Providers []shareddata.Provider

	// Benchmark data provider. If empty, collection is not created.
	BenchmarkProvider shareddata.BenchmarkProvider

	// ExtraOptions sets the options in MongoDB URI, when the option exists it overwrites that option.
	ExtraOptions url.Values

	// UseDriver enables low-level driver connection creation.
	UseDriver bool

	// DriverNoAuth disables automatic authentication of the low-level driver connection.
	DriverNoAuth bool

	// Options to override default backend configuration.
	BackendOptions *BackendOpts

	// DisableOtel disable OpenTelemetry monitoring for MongoDB client.
	DisableOtel bool
}

// BackendOpts represents backend configuration used for test setup.
type BackendOpts struct {
	// Capped collections cleanup interval.
	CappedCleanupInterval time.Duration

	// Percentage of documents to cleanup for capped collections. If not set, defaults to 20.
	CappedCleanupPercentage uint8

	// MaxBsonObjectSizeBytes is the maximum allowed size of a document, if not set FerretDB sets the default.
	MaxBsonObjectSizeBytes int

	// DisableNewAuth true uses the old backend authentication.
	DisableNewAuth bool
}

// SetupResult represents setup results.
type SetupResult struct {
	Ctx        context.Context
	Collection *mongo.Collection
	DriverConn *wireclient.Conn
	MongoDBURI string // without database name
}

// IsUnixSocket returns true if MongoDB URI is a Unix domain socket.
func (s *SetupResult) IsUnixSocket(tb testtb.TB) bool {
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
func SetupWithOpts(tb testtb.TB, opts *SetupOpts) *SetupResult {
	tb.Helper()

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	setupCtx, span := otel.Tracer("").Start(ctx, "SetupWithOpts")
	defer span.End()

	if opts == nil {
		opts = new(SetupOpts)
	}

	var levelVar slog.LevelVar
	levelVar.Set(slog.LevelError)
	if *debugSetupF {
		levelVar.Set(slog.LevelDebug)
	}

	logger := testutil.LevelLogger(tb, &levelVar)

	uri := *targetURLF
	if uri == "" {
		uri = setupListener(tb, setupCtx, logger, opts.BackendOptions)
	}

	if opts.ExtraOptions != nil {
		u, err := url.Parse(uri)
		require.NoError(tb, err)

		q := u.Query()

		for k, vs := range opts.ExtraOptions {
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

	if opts.UseDriver {
		u, err := url.Parse(uri)
		require.NoError(tb, err)

		query := u.Query()
		u.RawQuery = ""

		conn = setupClientDriver(tb, setupCtx, u.String(), testutil.Logger(tb))

		if u.User != nil && !opts.DriverNoAuth {
			user := u.User.Username()
			pass, _ := u.User.Password()

			mech := query.Get("authMechanism")
			if mech == "" {
				mech = "SCRAM-SHA-1"
			}

			authDB := query.Get("authSource")

			require.NoError(tb, driver.Authenticate(ctx, conn, user, password.WrapPassword(pass), mech, authDB))
		}
	}

	err := levelVar.UnmarshalText([]byte(*logLevelF))
	require.NoError(tb, err)

	return &SetupResult{
		Ctx:        ctx,
		Collection: collection,
		DriverConn: conn,
		MongoDBURI: uri,
	}
}

// Setup setups a single collection for all providers, if they are present.
func Setup(tb testtb.TB, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
	})
	return s.Ctx, s.Collection
}

// SetupDriver setups a single collection for all providers, if they are present,
// and returns authenticated low-level driver connection.
func SetupDriver(tb testtb.TB, providers ...shareddata.Provider) (context.Context, *wireclient.Conn) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
		UseDriver: true,
		ExtraOptions: url.Values{
			"authSource": []string{"admin"},
		},
	})

	return s.Ctx, s.DriverConn
}

// setupCollection setups a single collection for all providers, if they are present.
func setupCollection(tb testtb.TB, ctx context.Context, client *mongo.Client, opts *SetupOpts) *mongo.Collection {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupCollection")
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
	_ = collection.Drop(ctx)
	if ownDatabase {
		cleanupDatabase(ctx, tb, database, opts.BackendOptions)
	}

	var inserted bool

	switch {
	case len(opts.Providers) > 0:
		require.Nil(tb, opts.BenchmarkProvider, "Both Providers and BenchmarkProvider were set")
		inserted = InsertProviders(tb, ctx, collection, opts.Providers...)
	case opts.BenchmarkProvider != nil:
		inserted = insertBenchmarkProvider(tb, ctx, collection, opts.BenchmarkProvider)
	}

	if len(opts.Providers) == 0 && opts.BenchmarkProvider == nil {
		tb.Logf("Collection %s.%s wasn't created because no providers were set.", databaseName, collectionName)
	} else {
		require.True(tb, inserted)
	}

	// delete collection and (possibly) database unless test failed
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s.%s for debugging.", databaseName, collectionName)
			return
		}

		err := collection.Drop(ctx)
		require.NoError(tb, err)

		if ownDatabase {
			cleanupDatabase(ctx, tb, database, opts.BackendOptions)
		}
	})

	return collection
}

// InsertProviders inserts documents from specified Providers into collection. It returns true if any document was inserted.
func InsertProviders(tb testtb.TB, ctx context.Context, collection *mongo.Collection, providers ...shareddata.Provider) (inserted bool) {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "insertProviders")
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
func insertBenchmarkProvider(tb testtb.TB, ctx context.Context, collection *mongo.Collection, provider shareddata.BenchmarkProvider) (inserted bool) {
	tb.Helper()

	iter := provider.NewIterator()
	defer iter.Close()

	for {
		docs, err := iterator.ConsumeValuesN(iter, 100)
		require.NoError(tb, err)

		if len(docs) == 0 {
			break
		}

		insertDocs := make([]any, len(docs))
		for i, doc := range docs {
			insertDocs[i] = doc
		}

		res, err := collection.InsertMany(ctx, insertDocs)
		require.NoError(tb, err)
		require.Len(tb, res.InsertedIDs, len(docs))

		inserted = true
	}

	return
}

// cleanupUser removes users for the given database if new authentication is enabled and drops that database.
func cleanupDatabase(ctx context.Context, tb testtb.TB, database *mongo.Database, opts *BackendOpts) {
	ctx, span := otel.Tracer("").Start(ctx, "cleanupDatabase")
	defer span.End()

	if opts == nil || !opts.DisableNewAuth {
		err := database.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}}).Err()
		require.NoError(tb, err)
	}

	require.NoError(tb, database.Drop(ctx))
}

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
	"fmt"
	"net/url"
	"path/filepath"
	"runtime/trace"
	"strings"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Flags.
var (
	targetURLF     = flag.String("target-url", "", "target system's URL; if empty, in-process FerretDB is used")
	targetBackendF = flag.String("target-backend", "", "target system's backend: '%s'"+strings.Join(allBackends, "', '"))

	targetProxyAddrF  = flag.String("target-proxy-addr", "", "in-process FerretDB: use given proxy")
	targetTLSF        = flag.Bool("target-tls", false, "in-process FerretDB: use TLS")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "in-process FerretDB: use Unix socket")

	postgreSQLURLF = flag.String("postgresql-url", "", "in-process FerretDB: PostgreSQL URL for 'pg' handler.")
	sqliteURLF     = flag.String("sqlite-url", "", "in-process FerretDB: SQLite URI for 'sqlite' handler.")
	hanaURLF       = flag.String("hana-url", "", "in-process FerretDB: Hana URL for 'hana' handler.")

	compatURLF = flag.String("compat-url", "", "compat system's (MongoDB) URL for compatibility tests; if empty, they are skipped")

	benchDocsF = flag.Int("bench-docs", 1000, "benchmarks: number of documents to generate per iteration")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	disableFilterPushdownF = flag.Bool("disable-filter-pushdown", false, "disable filter pushdown")
	enableSortPushdownF    = flag.Bool("enable-sort-pushdown", false, "enable sort pushdown")
	enableOplogF           = flag.Bool("enable-oplog", false, "enable OpLog")

	useNewPGF = flag.Bool("use-new-pg", false, "use new PostgreSQL backend")
)

// Other globals.
var (
	allBackends = []string{"ferretdb-pg", "ferretdb-sqlite", "ferretdb-hana", "mongodb"}

	CertsRoot = filepath.Join("..", "build", "certs") // relative to `integration` directory
)

// SetupOpts represents setup options.
//
// Add option to use read-only user.
// TODO https://github.com/FerretDB/FerretDB/issues/1025
type SetupOpts struct {
	// Database to use. If empty, temporary test-specific database is created and dropped after test.
	DatabaseName string

	// Collection to use. If empty, temporary test-specific collection is created and dropped after test.
	// Most tests should keep this empty.
	CollectionName string

	// Data providers. If empty, collection is not created.
	Providers []shareddata.Provider

	// Benchmark data provider. If empty, collection is not created.
	BenchmarkProvider shareddata.BenchmarkProvider

	// ExtraOptions sets the options in MongoDB URI, when the option exists it overwrites that option.
	ExtraOptions url.Values
}

// SetupResult represents setup results.
type SetupResult struct {
	Ctx        context.Context
	Collection *mongo.Collection
	MongoDBURI string
}

// IsUnixSocket returns true if MongoDB URI is a Unix socket.
func (s *SetupResult) IsUnixSocket(tb testtb.TB) bool {
	tb.Helper()

	// we can't use a regular url.Parse because
	// MongoDB really wants Unix socket path in the host part of the URI
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

	defer trace.StartRegion(setupCtx, "SetupWithOpts").End()

	if opts == nil {
		opts = new(SetupOpts)
	}

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := testutil.LevelLogger(tb, level)

	uri := *targetURLF
	if uri == "" {
		uri = setupListener(tb, setupCtx, logger)
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

	client := setupClient(tb, setupCtx, uri)

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	collection := setupCollection(tb, setupCtx, client, opts)

	level.SetLevel(*logLevelF)

	return &SetupResult{
		Ctx:        ctx,
		Collection: collection,
		MongoDBURI: uri,
	}
}

// Setup setups a single collection for all providers, if the are present.
func Setup(tb testtb.TB, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
	})
	return s.Ctx, s.Collection
}

// setupCollection setups a single collection for all providers, if they are present.
func setupCollection(tb testtb.TB, ctx context.Context, client *mongo.Client, opts *SetupOpts) *mongo.Collection {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupCollection")
	defer span.End()

	defer observability.FuncCall(ctx)()

	var ownDatabase bool
	databaseName := opts.DatabaseName
	if databaseName == "" {
		databaseName = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	var ownCollection bool
	collectionName := opts.CollectionName
	if collectionName == "" {
		collectionName = testutil.CollectionName(tb)
		ownCollection = true
	}

	database := client.Database(databaseName)
	collection := database.Collection(collectionName)

	// drop remnants of the previous failed run
	_ = collection.Drop(ctx)
	if ownDatabase {
		_ = database.Drop(ctx)
	}

	var inserted bool

	switch {
	case len(opts.Providers) > 0:
		require.Nil(tb, opts.BenchmarkProvider, "Both Providers and BenchmarkProvider were set")
		inserted = insertProviders(tb, ctx, collection, opts.Providers...)
	case opts.BenchmarkProvider != nil:
		inserted = insertBenchmarkProvider(tb, ctx, collection, opts.BenchmarkProvider)
	}

	if len(opts.Providers) == 0 && opts.BenchmarkProvider == nil {
		tb.Logf("Collection %s.%s wasn't created because no providers were set.", databaseName, collectionName)
	} else {
		require.True(tb, inserted)
	}

	if ownCollection {
		// delete collection and (possibly) database unless test failed
		tb.Cleanup(func() {
			if tb.Failed() {
				tb.Logf("Keeping %s.%s for debugging.", databaseName, collectionName)
				return
			}

			err := collection.Drop(ctx)
			require.NoError(tb, err)

			if ownDatabase {
				err = database.Drop(ctx)
				require.NoError(tb, err)
			}
		})
	}

	return collection
}

// insertProviders inserts documents from specified Providers into collection. It returns true if any document was inserted.
func insertProviders(tb testtb.TB, ctx context.Context, collection *mongo.Collection, providers ...shareddata.Provider) (inserted bool) {
	tb.Helper()

	collectionName := collection.Name()

	for _, provider := range providers {
		spanName := fmt.Sprintf("insertProviders/%s/%s", collectionName, provider.Name())
		provCtx, span := otel.Tracer("").Start(ctx, spanName)
		region := trace.StartRegion(provCtx, spanName)

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(provCtx, docs)
		require.NoError(tb, err, "provider %q", provider.Name())
		require.Len(tb, res.InsertedIDs, len(docs))
		inserted = true

		region.End()
		span.End()
	}

	return
}

// insertBenchmarkProvider inserts documents from specified BenchmarkProvider into collection.
// It returns true if any document was inserted.
//
// The function calculates the checksum of all inserted documents and compare them with provider's hash.
func insertBenchmarkProvider(tb testtb.TB, ctx context.Context, collection *mongo.Collection, provider shareddata.BenchmarkProvider) (inserted bool) {
	tb.Helper()

	collectionName := collection.Name()

	spanName := fmt.Sprintf("insertBenchmarkProvider/%s/%s", collectionName, provider.Name())
	provCtx, span := otel.Tracer("").Start(ctx, spanName)
	region := trace.StartRegion(provCtx, spanName)

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

		res, err := collection.InsertMany(provCtx, insertDocs)
		require.NoError(tb, err)
		require.Len(tb, res.InsertedIDs, len(docs))

		inserted = true
	}

	region.End()
	span.End()

	return
}

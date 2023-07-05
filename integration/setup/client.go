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

package setup

import (
	"context"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// makeClient returns new client for the given working MongoDB URI.
func makeClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)

	clientOpts.SetMonitor(otelmongo.NewMonitor())

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// make sure that FerretDB-backend connection works
	_, err = client.ListDatabases(ctx, bson.D{})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return client, nil
}

// setupClient returns test-specific client for the given MongoDB URI.
//
// It disconnects automatically when test ends.
//
// If the connection can't be established, it panics,
// as it doesn't make sense to proceed with other tests if we couldn't connect in one of them.
func setupClient(tb testutil.TB, ctx context.Context, uri string) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	defer observability.FuncCall(ctx)()

	client, err := makeClient(ctx, uri)
	if err != nil {
		tb.Error(err)
		panic("setupClient: " + err.Error())
	}

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	return client
}

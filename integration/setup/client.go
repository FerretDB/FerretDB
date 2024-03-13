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
	"net/url"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// makeClient returns new client for the given working MongoDB URI.
func makeClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)

	clientOpts.SetMonitor(otelmongo.NewMonitor())

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// When too many connections are open, PostgreSQL returns an error,
	// but this error is hanging (panic in setupClient doesn't immediately stop the test).
	// TODO https://github.com/FerretDB/FerretDB/issues/3535
	// if err = client.Ping(ctx, nil); err != nil {
	// 	return nil, lazyerrors.Error(err)
	// }

	return client, nil
}

// setupClient returns test-specific client for the given MongoDB URI.
//
// It disconnects automatically when test ends.
//
// If the connection can't be established, it panics,
// as it doesn't make sense to proceed with other tests if we couldn't connect in one of them.
func setupClient(tb testtb.TB, ctx context.Context, uri string) *mongo.Client {
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

// toAbsolutePathUri replaces tlsCertificateKeyFile and tlsCaFile path to an absolute path.
// If the test is run from subdirectory such as `integration/user/`, the absolute path
// is found by looking for the file in the parent directory.
func toAbsolutePathUri(tb testtb.TB, uri string) string {
	u, err := url.Parse(uri)
	require.NoError(tb, err)

	values := url.Values{}

	for k, v := range u.Query() {
		require.Len(tb, v, 1)

		switch k {
		case "tlsCertificateKeyFile", "tlsCaFile":
			file := v[0]

			if !filepath.IsAbs(file) {
				file = filepath.Join(Dir(tb), v[0])
			}

			_, err := os.Stat(file)
			if os.IsNotExist(err) {
				file = filepath.Join(Dir(tb), "..", v[0])
			}

			values[k] = []string{file}
		default:
			values[k] = v
		}
	}

	u.RawQuery = values.Encode()

	return u.String()
}

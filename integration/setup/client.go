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
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// setClientPaths replaces file names in query parameters with absolute paths.
func setClientPaths(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	for k, vs := range q {
		switch k {
		case "tlsCertificateKeyFile", "tlsCaFile":
			for i, v := range vs {
				if strings.Contains(v, "/") {
					return "", lazyerrors.Errorf("%q: %q should contain only a file name, got %q", uri, k, v)
				}

				vs[i] = filepath.Join(testutil.BuildCertsDir, v)
			}
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// makeClient returns new client for the given working MongoDB URI.
func makeClient(ctx context.Context, uri string, disableOtel bool) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)

	if !disableOtel {
		clientOpts.SetMonitor(otelmongo.NewMonitor(otelmongo.WithCommandAttributeDisabled(false)))
	}

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
func setupClient(tb testtb.TB, ctx context.Context, uri string, disableOtel bool) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	client, err := makeClient(ctx, uri, disableOtel)
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

// setupClient returns test-specific non-authenticated low-level driver connection for the given MongoDB URI.
//
// It disconnects automatically when test ends.
//
// If the connection can't be established, it panics,
// as it doesn't make sense to proceed with other tests if we couldn't connect in one of them.
func setupClientDriver(tb testtb.TB, ctx context.Context, uri string, l *slog.Logger) *wireclient.Conn {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClientDriver")
	defer span.End()

	conn, err := wireclient.Connect(ctx, uri, l)
	if err != nil {
		tb.Error(err)
		panic("setupClientDriver: " + err.Error())
	}

	tb.Cleanup(func() {
		err = conn.Close()
		require.NoError(tb, err)
	})

	return conn
}

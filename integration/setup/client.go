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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// mongoDBURIOpts represents mongoDBURI's options.
type mongoDBURIOpts struct {
	hostPort       string // for TCP and TLS
	unixSocketPath string
	tlsAndAuth     bool
}

// mongoDBURI builds MongoDB URI with given options.
func mongoDBURI(tb testing.TB, opts *mongoDBURIOpts) string {
	tb.Helper()

	var host string

	if opts.hostPort != "" {
		require.Empty(tb, opts.unixSocketPath, "both hostPort and unixSocketPath are set")
		host = opts.hostPort
	} else {
		host = opts.unixSocketPath
	}

	var user *url.Userinfo
	q := make(url.Values)

	if opts.tlsAndAuth {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath cannot be used with TLS")

		// we don't separate TLS and auth just for simplicity of our test configurations
		q.Set("tls", "true")
		q.Set("tlsCertificateKeyFile", filepath.Join(CertsRoot, "client.pem"))
		q.Set("tlsCaFile", filepath.Join(CertsRoot, "rootCA-cert.pem"))
		q.Set("authMechanism", "PLAIN")
		user = url.UserPassword("username", "password")
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme:   "mongodb",
		Host:     host,
		Path:     "/",
		User:     user,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// makeClient returns new client for the given MongoDB URI and extra options.
// The extraOpts that comes later overwrites value for the query key.
func makeClient(ctx context.Context, uri string, extraOpts url.Values) (*mongo.Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	for k, vs := range extraOpts {
		for _, v := range vs {
			q.Set(k, v)
		}
	}

	u.RawQuery = q.Encode()
	clientOpts := options.Client().ApplyURI(u.String())
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
func setupClient(tb testing.TB, ctx context.Context, uri string, extraOpts url.Values) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	defer observability.FuncCall(ctx)()

	client, err := makeClient(ctx, uri, extraOpts)
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

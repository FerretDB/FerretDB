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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	targetPortF = flag.Int("target-port", 0, "target system's port for tests; if 0, in-process FerretDB is used")
	proxyAddrF  = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")
	handlerF    = flag.String("handler", "pg", "handler to use for in-process FerretDB") // TODO
	compatPortF = flag.Int("compat-port", 37017, "second system's port for compatibility tests; if 0, they are skipped")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	startupOnce sync.Once
)

// setupClient returns MongoDB client for database on 127.0.0.1:port.
func setupClient(tb testing.TB, ctx context.Context, port int) *mongo.Client {
	tb.Helper()

	require.Greater(tb, port, 0)
	require.Less(tb, port, 65536)

	// those options should not affect anything except tests speed
	v := url.Values{
		// TODO: Test fails occured on some platforms due to i/o timeout.
		// Needs more investigation.
		//
		//"connectTimeoutMS":         []string{"5000"},
		//"serverSelectionTimeoutMS": []string{"5000"},
		//"socketTimeoutMS":          []string{"5000"},
		//"heartbeatFrequencyMS":     []string{"30000"},

		//"minPoolSize":   []string{"1"},
		//"maxPoolSize":   []string{"1"},
		//"maxConnecting": []string{"1"},
		//"maxIdleTimeMS": []string{"0"},

		//"directConnection": []string{"true"},
		//"appName":          []string{tb.Name()},
	}

	u := url.URL{
		Scheme:   "mongodb",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     "/",
		RawQuery: v.Encode(),
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	require.NoError(tb, err)

	err = client.Ping(ctx, nil)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	return client
}

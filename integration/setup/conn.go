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

	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// setupWireConn returns test-specific non-authenticated low-level wire connection for the given MongoDB URI.
//
// It disconnects automatically when test ends.
//
// If the connection can't be established, it panics,
// as it doesn't make sense to proceed with other tests if we couldn't connect in one of them.
func setupWireConn(tb testtb.TB, ctx context.Context, uri string, l *slog.Logger) *wireclient.Conn {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupWireConn")
	defer span.End()

	conn, err := wireclient.Connect(ctx, uri, l)
	if err != nil {
		tb.Error(err)
		panic("setupWireConn: " + err.Error())
	}

	tb.Cleanup(func() {
		err = conn.Close()
		require.NoError(tb, err)
	})

	return conn
}

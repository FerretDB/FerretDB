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

package handlers

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

type setupOpts struct {
	noLogging bool
	poolOpts  *testutil.PoolOpts
}

// setup creates shared objects for testing.
//
// Using shared objects helps us spot concurrency bugs.
// If some test is failing and the log output is confusing, and you are tempted to move setup call to subtest,
// instead run that single test with `go test -run test/name`.
func setup(t testing.TB, opts *setupOpts) (context.Context, common.Handler, *pgdb.Pool) {
	t.Helper()

	if opts == nil {
		opts = new(setupOpts)
	}

	var l *zap.Logger
	if opts.noLogging {
		l = zap.NewNop()
	} else {
		l = zaptest.NewLogger(t)
	}

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, opts.poolOpts, l)
	handler := pg.New(&pg.NewOpts{
		PgPool:   pool,
		L:        l,
		PeerAddr: "127.0.0.1:12345",
	})

	return ctx, handler, pool
}

// addToSeedCorpus adds given header and message body to handler's fuzzing seed corpus.
//
//nolint:deadcode // TODO https://github.com/FerretDB/FerretDB/issues/438
func addToSeedCorpus(tb testing.TB, header *wire.MsgHeader, msg wire.MsgBody) {
	tb.Skip("TODO https://github.com/FerretDB/FerretDB/issues/438")

	tb.Helper()

	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)
	err := wire.WriteMessage(bufw, header, msg)
	require.NoError(tb, err)
	err = bufw.Flush()
	require.NoError(tb, err)

	testutil.WriteSeedCorpusFile(tb, "FuzzHandler", buf.Bytes())
}

func FuzzHandler(f *testing.F) {
	f.Skip("TODO https://github.com/FerretDB/FerretDB/issues/438")

	ctx, handler, _ := setup(f, &setupOpts{
		// to avoid panic: testing: f.Logf was called inside the fuzz target, use t.Logf instead
		noLogging: true,
		poolOpts: &testutil.PoolOpts{
			ReadOnly: true,
		},
	})

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		br := bytes.NewReader(b)
		bufr := bufio.NewReader(br)

		reqHeader, reqBody, err := wire.ReadMessage(bufr)
		if err != nil {
			t.Skip()
		}

		// check only panics for now
		common.Route(handler, ctx, reqHeader, reqBody)
	})
}

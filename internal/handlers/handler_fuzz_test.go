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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

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
		handler.Handle(ctx, reqHeader, reqBody)
	})
}

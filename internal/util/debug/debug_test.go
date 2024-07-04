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

// Package debug provides debug facilities.
package debug

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func assertProbe(t *testing.T, u string, expected int) {
	t.Helper()

	res, err := http.Get(u)
	require.NoError(t, err)
	assert.Equal(t, expected, res.StatusCode)
}

func TestProbes(t *testing.T) {
	t.Parallel()

	var livez, readyz atomic.Bool

	h, err := Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       testutil.Logger(t),
		R:       prometheus.NewRegistry(),
		Livez:   livez.Load,
		Readyz:  readyz.Load,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	done := make(chan struct{})

	go func() {
		h.Serve(ctx)
		close(done)
	}()

	live := "http://" + h.lis.Addr().String() + "/debug/livez"
	ready := "http://" + h.lis.Addr().String() + "/debug/readyz"

	assertProbe(t, live, http.StatusInternalServerError)
	assertProbe(t, ready, http.StatusInternalServerError)

	readyz.Store(true)

	assertProbe(t, live, http.StatusInternalServerError)
	assertProbe(t, ready, http.StatusInternalServerError)

	livez.Store(true)

	assertProbe(t, live, http.StatusOK)
	assertProbe(t, ready, http.StatusOK)

	cancel()
	<-done // prevent panic on logging after test ends
}

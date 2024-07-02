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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDebugHandlerStartupProbe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	var started atomic.Bool

	var wg sync.WaitGroup
	wg.Add(1)

	h, err := Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       testutil.Logger(t),
		R:       prometheus.NewRegistry(),
		Started: &started,
	})
	require.NoError(t, err)

	addr := h.lis.Addr()

	go func() {
		h.Serve(ctx)
		wg.Done()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr.String()+"/debug/started", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	started.Store(true)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Cancel the context to stop the handler.
	// The WaitGroup is needed to make sure that all logs were printed before the test finished.
	cancel()
	wg.Wait()
}

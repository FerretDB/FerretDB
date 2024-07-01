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
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestRunHandlerStartupProbe(t *testing.T) {
	t.Parallel()

	// create and close TCP socket, to obtain a free port
	l, err := net.ListenTCP("tcp", must.NotFail(net.ResolveTCPAddr("tcp", "localhost:0")))
	require.NoError(t, err)

	require.NoError(t, l.Close())

	ctx, cancel := context.WithCancel(testutil.Ctx(t))

	addr := l.Addr().(*net.TCPAddr)
	started := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	h, err := Listen(&ListenOpts{
		TCPAddr:         addr.String(), //TODO
		L:               testutil.Logger(t),
		R:               prometheus.NewRegistry(),
		FerretdbStarted: started,
	})
	require.NoError(t, err)

	go func() {
		h.Serve(ctx)
		wg.Done()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr.String()+"/debug/started", nil)
	require.NoError(t, err)

	// Wait for handler
	for range 5 {
		_, err = http.DefaultClient.Do(req)
		if err == nil {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	close(started)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Cancel the context to stop RunHandler.
	// The WaitGroup is needed to make sure that all logs were printed before the test finished.
	cancel()
	wg.Wait()
}

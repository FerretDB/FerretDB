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

package dataapi

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestGracefulShutdown(t *testing.T) {
	t.Parallel()

	addr, _ := setupDataAPI(t, false)

	// Test that the server starts and can serve requests
	t.Run("ServerResponds", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get("http://" + addr + "/openapi.json")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test basic cancellation behavior (graceful shutdown is tested indirectly through setupDataAPI cleanup)
	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(testutil.Ctx(t))
		defer cancel()

		l := testutil.Logger(t)

		// Create a listener that we can control
		apiLis, err := Listen(&ListenOpts{
			TCPAddr: "127.0.0.1:0",
			L:       l,
			Handler: nil, // This will cause a panic if accessed, but we won't access it
		})
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)

		finished := make(chan struct{})

		go func() {
			defer wg.Done()
			defer close(finished)
			// This should exit when ctx is canceled
			apiLis.Run(ctx)
		}()

		// Give the server a moment to start
		time.Sleep(10 * time.Millisecond)

		// Cancel the context
		cancel()

		// Wait for Run to finish with timeout
		select {
		case <-finished:
			// Success - Run exited
		case <-time.After(5 * time.Second):
			t.Fatal("Run did not exit within timeout after context cancellation")
		}

		wg.Wait()
	})
}
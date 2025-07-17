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
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestDataAPIEndpoints(t *testing.T) {
	t.Parallel()

	// Create a simple listener without database dependency
	l := testutil.Logger(t)
	apiLis, err := Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       l,
		Handler: nil, // nil handler is OK for OpenAPI endpoint
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	defer cancel()

	// Start the server
	done := make(chan struct{})
	go func() {
		defer close(done)
		apiLis.Run(ctx)
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	addr := apiLis.lis.Addr().String()

	t.Run("OpenAPIEndpoint", func(t *testing.T) {
		resp, err := http.Get("http://" + addr + "/openapi.json")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// Validate that it's valid JSON
		var spec map[string]interface{}
		err = json.Unmarshal(body, &spec)
		require.NoError(t, err)

		// Check that it has the expected OpenAPI structure
		assert.Equal(t, "3.0.0", spec["openapi"])
		assert.Contains(t, spec, "info")
		assert.Contains(t, spec, "paths")
		assert.Contains(t, spec, "components")

		// Check that it's the FerretDB Data API spec
		info := spec["info"].(map[string]interface{})
		assert.Equal(t, "FerretDB Data API", info["title"])
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/openapi.json", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	// Test graceful shutdown
	t.Run("GracefulShutdown", func(t *testing.T) {
		// Cancel the context to trigger shutdown
		cancel()

		// Wait for server to stop gracefully
		select {
		case <-done:
			// Success - server stopped gracefully
		case <-time.After(5 * time.Second):
			t.Fatal("Server did not stop gracefully within timeout")
		}
	})
}
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

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestOpenAPISpec(t *testing.T) {
	t.Parallel()

	server := &Server{
		l: testutil.Logger(t),
	}

	t.Run("GetOpenAPISpec", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		w := httptest.NewRecorder()

		server.OpenAPISpec(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))

		// Validate that it's valid JSON
		var spec map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&spec)
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
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/openapi.json", nil)
		w := httptest.NewRecorder()

		server.OpenAPISpec(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
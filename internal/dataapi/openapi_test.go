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
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPI(t *testing.T) {
	t.Parallel()

	addr, _ := setupDataAPI(t, false)

	t.Run("Spec", func(t *testing.T) {
		resp, err := http.Get("http://" + addr + "/openapi.json")
		require.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, resp.Body.Close())
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", resp.Header.Get("Cache-Control"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var spec map[string]any
		err = json.Unmarshal(body, &spec)
		require.NoError(t, err)

		assert.Equal(t, "3.0.0", spec["openapi"])
		assert.Contains(t, spec, "info")
		assert.Contains(t, spec, "paths")
		assert.Contains(t, spec, "components")

		info := spec["info"].(map[string]any)
		assert.Equal(t, "FerretDB Data API", info["title"])
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/openapi.json", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, resp.Body.Close())
		})

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

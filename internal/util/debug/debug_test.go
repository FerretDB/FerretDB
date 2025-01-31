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
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func assertProbe(t *testing.T, u string, expected int) {
	t.Helper()

	res, err := http.Get(u)
	require.NoError(t, err)
	assert.Equal(t, expected, res.StatusCode)
}

func TestDebug(t *testing.T) {
	t.Parallel()

	var livez, readyz atomic.Bool

	h := must.NotFail(Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       testutil.Logger(t),
		R:       prometheus.NewRegistry(),
		Livez:   func(context.Context) bool { return livez.Load() },
		Readyz:  func(context.Context) bool { return readyz.Load() },
	}))

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	done := make(chan struct{})

	go func() {
		h.Serve(ctx)
		close(done)
	}()

	t.Run("Probes", func(t *testing.T) {
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
	})

	t.Run("Archive", func(t *testing.T) {
		u := "http://" + h.lis.Addr().String() + "/debug/archive"

		expectedFiles := map[string]struct{}{
			"goroutine.pprof": {},
			"heap.pprof":      {},
			"profile.pprof":   {},
			"metrics.txt":     {},
			"vars.json":       {},
			"errors.txt":      {},
		}

		res, err := http.Get(u)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
		assert.Regexp(t, regexp.MustCompile(`attachment; filename=ferretdb-[\d-]+.zip`), res.Header.Get("Content-Disposition"))

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		require.NoError(t, err)

		for _, file := range zipReader.File {
			name := file.FileHeader.Name
			assert.Contains(t, expectedFiles, file.FileHeader.Name)
			delete(expectedFiles, file.FileHeader.Name)

			t.Run(name, func(t *testing.T) {
				f, e := file.Open()
				require.NoError(t, e)

				defer func() {
					assert.NoError(t, f.Close())
				}()

				b, e := io.ReadAll(f)
				require.NoError(t, e)
				assert.NotEmpty(t, b)

				if name == "errors.txt" {
					t.Logf("\n%s", b)
				}
			})
		}

		assert.Empty(t, expectedFiles)
	})

	cancel()
	<-done // prevent panic on logging after test ends
}

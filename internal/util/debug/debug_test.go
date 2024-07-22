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

	"github.com/FerretDB/FerretDB/internal/util/testutil"
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

	h, err := Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       testutil.SLogger(t),
		R:       prometheus.NewRegistry(),
		Livez:   func(context.Context) bool { return livez.Load() },
		Readyz:  func(context.Context) bool { return readyz.Load() },
	})
	require.NoError(t, err)

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

		fileList := []string{
			"metrics", "heap",
		}

		var res *http.Response

		res, err = http.Get(u)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		require.Equal(t, "application/zip", res.Header.Get("Content-Type"))

		contentDispositionRegexp := regexp.MustCompile(`attachment; filename=FerretDB-debug-\d+.zip`)
		require.Regexp(t, contentDispositionRegexp, res.Header.Get("Content-Disposition"))

		var body []byte
		body, err = io.ReadAll(res.Body)

		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		var zipReader *zip.Reader

		zipReader, err = zip.NewReader(bytes.NewReader(body), int64(len(body)))
		require.NoError(t, err)
		require.Equal(t, len(fileList), len(zipReader.File))

		for _, file := range zipReader.File {
			require.Contains(t, fileList, file.FileHeader.Name)

			var f io.ReadCloser

			f, err = file.Open()
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, f.Close())
			})

			content := make([]byte, 1)

			var n int
			n, err = f.Read(content)
			require.NoError(t, err)

			assert.Equal(t, 1, n, "file should contain any data, but was empty")
		}
	})

	cancel()
	<-done // prevent panic on logging after test ends
}

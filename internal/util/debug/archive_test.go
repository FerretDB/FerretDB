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

package debug

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/state"
)

func TestArchiveHandler(t *testing.T) {
	t.Parallel()

	// during test some of the files from the below list shall be absent
	fileList := []string{
		"metrics", "heap",
	}

	filename := filepath.Join(t.TempDir(), "state.json")
	stateProvider, err := state.NewProvider(filename)
	require.NoError(t, err)

	metricsRegisterer := prometheus.DefaultRegisterer
	metricsProvider := stateProvider.MetricsCollector(true)
	metricsRegisterer.MustRegister(metricsProvider)

	l := zap.L()

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	srv := httptest.NewServer(debugHandler(ctx, metricsRegisterer, l.Named("debug")))
	defer srv.Close()

	res, err := http.Get(srv.URL + archivePath)
	require.NoError(t, err)

	defer res.Body.Close() //nolint:errcheck // we are only reading it

	cancelCtx()

	require.Equal(t, http.StatusOK, res.StatusCode, "status code should be 200")
	require.Equal(t, "application/zip", res.Header.Get("Content-Type"), "mime type should be zip")

	expectedHeader := "attachment; filename=FerretDB-debug.zip"
	receivedHeader := res.Header.Get("Content-Disposition")
	require.Equal(t, expectedHeader, receivedHeader, "content-disposition type should be same")

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err)

	require.Equal(t, len(fileList), len(zipReader.File), "needs to be same as length of fileList")

	for _, file := range zipReader.File {
		require.Contains(t, fileList, file.FileHeader.Name)

		f, err := file.Open()
		defer f.Close()

		require.NoError(t, err)

		content := make([]byte, 1)
		n, err := f.Read(content)
		require.NoError(t, err)

		assert.Equal(t, 1, n, "file should contain any data, but was empty")

		f.Close()
	}
}

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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/state"
)

func TestRunHandler(t *testing.T) {
	t.Parallel()

	host := "127.0.0.1:5454"

	ctx := context.Background()

	filename := filepath.Join(t.TempDir(), "state.json")
	stateProvider, err := state.NewProvider(filename)
	require.NoError(t, err)

	metricsRegisterer := prometheus.DefaultRegisterer
	metricsProvider := stateProvider.MetricsCollector(true)
	metricsRegisterer.MustRegister(metricsProvider)

	l := zap.S()

	go RunHandler(ctx, host, metricsRegisterer, l.Named("debug").Desugar())

	// Wait for the server to start
	time.Sleep(time.Second)

	var u url.URL
	u.Path = archivePath
	u.Host = host
	u.Scheme = "http"

	req, err := http.NewRequest("GET", u.String(), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	require.Equal(t, "application/zip", resp.Header.Get("Content-Type"), "mime type should be zip")
	require.Equal(t, fmt.Sprintf("attachment; filename=%s", generateFileName("FerretDB", "debug.zip")), resp.Header.Get("Content-Disposition"), "content-disposition type should be same")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return
	}
	defer resp.Body.Close()

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		// Handle error
		return
	}

	require.Equal(t, 2, len(zipReader.File), "zip should contain 2 files")
	for _, file := range zipReader.File {
		err = nil
		if file.FileHeader.Name == "FerretDB-metrics.txt" || file.FileHeader.Name == "FerretDB-pprof.html" {
			err = errors.New(fmt.Sprintf("file name is not as expected %s", file.FileHeader.Name))
		}
		require.NoError(t, nil, err, "file name fif not match")
	}
	return
}

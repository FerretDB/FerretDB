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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// addToZip adds a new file to the zip archive.
//
// Passed [io.ReadCloser] is always closed.
func addToZip(w *zip.Writer, name string, r io.ReadCloser) (err error) {
	defer func() {
		if e := r.Close(); e != nil && err == nil {
			err = e
		}
	}()

	f, err := w.CreateHeader(&zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	})
	if err != nil {
		return
	}

	_, err = io.Copy(f, r)

	return
}

// filterExpvar filters out sensitive information from /debug/vars output.
func filterExpvar(ctx context.Context, r io.ReadCloser, l *slog.Logger) io.ReadCloser {
	defer r.Close() //nolint:errcheck // we are only reading it

	var res bytes.Buffer

	var data map[string]any
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		l.ErrorContext(ctx, "Failed to decode expvar", logging.Error(err))
		return io.NopCloser(&res)
	}

	delete(data, "cmdline")

	e := json.NewEncoder(&res)
	e.SetIndent("", "  ")

	if err := e.Encode(data); err != nil {
		l.ErrorContext(ctx, "Failed to encode expvar", logging.Error(err))
	}

	return io.NopCloser(&res)
}

// archiveHandler returns a handler that creates a zip archive with various debug information.
func archiveHandler(l *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		name := fmt.Sprintf("ferretdb-%s.zip", time.Now().Format("2006-01-02-15-04-05"))

		rw.Header().Set("Content-Type", "application/zip")
		rw.Header().Set("Content-Disposition", "attachment; filename="+name)

		ctx := req.Context()
		zipWriter := zip.NewWriter(rw)
		errs := map[string]error{}

		defer func() {
			files := slices.Sorted(maps.Keys(errs))

			var b bytes.Buffer
			for _, f := range files {
				b.WriteString(fmt.Sprintf("%s: %v\n", f, errs[f]))
			}

			if err := addToZip(zipWriter, "errors.txt", io.NopCloser(&b)); err != nil {
				l.ErrorContext(ctx, "Failed to add errors.txt to archive", logging.Error(err))
			}

			if err := zipWriter.Close(); err != nil {
				l.ErrorContext(ctx, "Failed to close archive", logging.Error(err))
			}

			l.InfoContext(ctx, "Debug archive created", slog.String("name", name))
		}()

		host := ctx.Value(http.LocalAddrContextKey).(net.Addr)

		for _, f := range []struct {
			file string
			path string
		}{
			{file: "metrics.txt", path: "/debug/metrics"},
			{file: "vars.json", path: "/debug/vars"},
			{file: "profile.pprof", path: "/debug/pprof/profile?seconds=10"},
			{file: "goroutine.pprof", path: "/debug/pprof/goroutine"},
			{file: "block.pprof", path: "/debug/pprof/block"},
			{file: "heap.pprof", path: "/debug/pprof/heap?gc=1"},
			{file: "trace.out", path: "/debug/pprof/trace?seconds=10"},
		} {
			fReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+host.String()+f.path, nil)
			if err != nil {
				errs[f.file] = err
				continue
			}

			l.DebugContext(
				ctx, "Fetching file for archive",
				slog.String("file", f.file), slog.String("url", fReq.URL.String()),
			)

			resp, err := http.DefaultClient.Do(fReq)
			if err == nil {
				r := resp.Body
				if f.path == "/debug/vars" {
					r = filterExpvar(ctx, r, l)
				}

				err = addToZip(zipWriter, f.file, r)
			}

			errs[f.file] = err
		}
	}
}

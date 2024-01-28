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
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// archiveHandler is a handler for creating the archive containing metrics and heap info.
func archiveHandler(rw http.ResponseWriter, req *http.Request) {
	zipWriter := zip.NewWriter(rw)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		}
	}()

	for _, path := range []string{
		metricsPath,
		pprofPath + "/heap",
	} {
		values := url.Values{}
		values.Add("debug", "1")

		u := url.URL{
			Scheme:   "http",
			Host:     req.Host,
			Path:     path,
			RawQuery: values.Encode(),
		}

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		fileName := filepath.Base(path)

		err = addFileToArchive(fileName, resp.Body, zipWriter)
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}
	}

	rw.Header().Set("Content-Type", "application/zip")
	rw.Header().Set("Content-Disposition", "attachment; filename=FerretDB-debug.zip")
}

// addFileToArchive function responsible for adding content to zip file.
func addFileToArchive(fileName string, fileReader io.ReadCloser, zipWriter *zip.Writer) error {
	defer fileReader.Close() //nolint:errcheck // we are only reading it

	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = io.Copy(fileWriter, fileReader)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

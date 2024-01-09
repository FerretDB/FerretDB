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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
)

type archiveUrlList []url.URL

// archiveHandler - is a handler for creating the archive containing all the debug and metrics info.
func archiveHandler(rw http.ResponseWriter, req *http.Request) {
	var (
		scheme  string
		urlList archiveUrlList
	)

	urlList = make(archiveUrlList, 0)

	u := new(url.URL)
	u.Path = metricsPath
	u.Host = req.Host

	if req.URL.Scheme == "" {
		scheme = "http"
	}

	u.Scheme = scheme
	urlList = append(urlList, *u)

	u = new(url.URL)
	u.Path = heapPath
	u.Host = req.Host

	if req.URL.Scheme == "" {
		scheme = "http"
	}

	u.Scheme = scheme
	urlList = append(urlList, *u)

	rw.Header().Set("Content-Type", "application/zip")

	var (
		debugFilePrefix = "FerretDB"
		debugFileName   = "debug.zip"
	)

	zipName := fmt.Sprintf("%s-%s", debugFilePrefix, debugFileName)
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipName))

	zipWriter := zip.NewWriter(rw)
	defer zipWriter.Close() //nolint:errcheck // we are only reading it

	for _, fileUrl := range urlList {
		fileName := filepath.Base(fileUrl.Path)

		resp, err := performRequest(&fileUrl)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = addFileToArchive(fileName, resp, zipWriter)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func addFileToArchive(fileName string, resp *http.Response, zipWriter *zip.Writer) error {
	defer resp.Body.Close() //nolint:errcheck // we are only reading it

	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		err = fmt.Errorf("fail creating file %s (error: %s)", fileName, err.Error())
		return err
	}

	_, err = io.Copy(fileWriter, resp.Body)
	if err != nil {
		err = fmt.Errorf("failed - adding %s to zip (error: %s)", fileName, err.Error())
		return err
	}

	return nil
}

// performRequest - performs the requests and return response.
func performRequest(u *url.URL) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		err = fmt.Errorf("request creation failed for URL %s (error: %s)", u.String(), err.Error())
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request fetch failed for URL %s (error: %s)", u.String(), err.Error())
		return nil, err
	}

	return resp, nil
}

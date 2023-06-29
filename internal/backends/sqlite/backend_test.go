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

package sqlite

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBackend(t *testing.T) {
	t.Parallel()

	err := os.MkdirAll("tmp/dir", os.ModePerm)
	require.NoError(t, err)

	_, err = os.Create("tmp/file")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.Remove("tmp/file")
		require.NoError(t, err)
		err = os.Remove("tmp/dir")
		require.NoError(t, err)
		err = os.Remove("tmp")
		require.NoError(t, err)
	})

	testCases := map[string]struct {
		params *NewBackendParams

		uri      *url.URL
		errRegex string
	}{
		"LocalDirectory": {
			params: &NewBackendParams{
				URI: "file:./",
				L:   zap.NewNop(),
			},
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./",
			},
		},
		"LocalSubDirectory": {
			params: &NewBackendParams{
				URI: "file:./tmp/",
				L:   zap.NewNop(),
			},
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./tmp/",
			},
		},
		"LocalSubSubDirectory": {
			params: &NewBackendParams{
				URI: "file:./tmp/dir/",
				L:   zap.NewNop(),
			},
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./tmp/dir/",
			},
		},
		"LocalDirectoryWithParameters": {
			params: &NewBackendParams{
				URI: "file:./tmp/?mode=ro",
				L:   zap.NewNop(),
			},
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./tmp/",
				RawQuery: "mode=ro",
			},
		},
		"HostIsNotEmpty": {
			params: &NewBackendParams{
				URI: "file://localhost/./tmp/?mode=ro",
				L:   zap.NewNop(),
			},
			errRegex: `.*backend URI should not contain host: "file://localhost/./tmp/\?mode=ro"`,
		},
		"UserIsNotEmpty": {
			params: &NewBackendParams{
				URI: "file://user:pass@./tmp/?mode=ro",
				L:   zap.NewNop(),
			},
			errRegex: `.*backend URI should not contain user: "file://user:pass@./tmp/\?mode=ro"`,
		},
		"NoDirectory": {
			params: &NewBackendParams{
				URI: "file:./nodir/",
				L:   zap.NewNop(),
			},
			errRegex: `.*"file:./nodir/" should be an existing directory: stat ./nodir/: no such file or directory`,
		},
		"FileInsteadOfDirectory": {
			params: &NewBackendParams{
				URI: "file:./tmp/file",
				L:   zap.NewNop(),
			},
			errRegex: `.*"file:./tmp/file" should be an existing directory`,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uri, err := validateParams(tc.params)
			if err != nil {
				t.Log(err)

				require.Regexp(t, tc.errRegex, err.Error())
				return
			}
			require.NoError(t, err)

			require.Equal(t, uri, tc.uri)
		})
	}
}

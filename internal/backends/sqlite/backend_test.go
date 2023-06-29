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
		value string

		uri      *url.URL
		errRegex string
	}{
		"LocalDirectory": {
			value: "file:./",
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./",
				Path:   "./",
			},
		},
		"LocalSubDirectory": {
			value: "file:./tmp/",
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./tmp/",
				Path:   "./tmp/",
			},
		},
		"LocalSubSubDirectory": {
			value: "file:./tmp/dir/",
			uri: &url.URL{
				Scheme: "file",
				Opaque: "./tmp/dir/",
				Path:   "./tmp/dir/",
			},
		},
		"LocalDirectoryWithParameters": {
			value: "file:./tmp/?mode=ro",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./tmp/",
				Path:     "./tmp/",
				RawQuery: "mode=ro",
			},
		},
		"AbsoluteDirectory": {
			value: "file:/tmp/",
			uri: &url.URL{
				Scheme:   "file",
				Path:     "/tmp/",
				OmitHost: true,
			},
		},
		"WithEmptyAuthority": {
			value: "file:///tmp/",
			uri: &url.URL{
				Scheme: "file",
				Path:   "/tmp/",
			},
		},
		"WithEmptyAuthorityAndQuery": {
			value: "file:///tmp/?mode=ro",
			uri: &url.URL{
				Scheme:   "file",
				Path:     "/tmp/",
				RawQuery: "mode=ro",
			},
		},
		"HostIsNotEmpty": {
			value:    "file://localhost/./tmp/?mode=ro",
			errRegex: `.*backend URI should not contain host: "file://localhost/./tmp/\?mode=ro"`,
		},
		"UserIsNotEmpty": {
			value:    "file://user:pass@./tmp/?mode=ro",
			errRegex: `.*backend URI should not contain user: "file://user:pass@./tmp/\?mode=ro"`,
		},
		"NoDirectory": {
			value:    "file:./nodir/",
			errRegex: `.*"file:./nodir/" should be an existing directory: stat ./nodir/: no such file or directory`,
		},
		"PathIsNotEndsWithSlash": {
			value:    "file:./tmp/file",
			errRegex: `.*backend URI should be a directory: "file:./tmp/file"`,
		},
		"FileInsteadOfDirectory": {
			value:    "file:./tmp/file/",
			errRegex: `.*file:./tmp/file/" should be an existing directory: stat ./tmp/file/: not a directory`,
		},
		"MalformedURI": {
			value:    ":./tmp/",
			errRegex: `.*failed to parse backend URI: parse ":./tmp/": missing protocol scheme`,
		},
		"NoScheme": {
			value:    "./tmp/",
			errRegex: `.*backend URI should have file scheme: "./tmp/"`,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uri, err := validateURI(tc.value)
			if tc.errRegex != "" {
				t.Log(err)

				require.Error(t, err)

				require.Regexp(t, tc.errRegex, err.Error())
				return
			}
			require.NoError(t, err)

			require.Equal(t, uri, tc.uri)
		})
	}
}

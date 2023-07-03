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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	t.Parallel()

	err := os.MkdirAll("tmp/dir", 0o777)
	require.NoError(t, err)

	_, err = os.Create("tmp/file")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll("tmp")
		require.NoError(t, err)
	})

	testCases := map[string]struct {
		value string
		uri   *url.URL
		err   string
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
				Opaque:   "/tmp/",
				Path:     "/tmp/",
				OmitHost: true,
			},
		},
		"WithEmptyAuthority": {
			value: "file:///tmp/",
			uri: &url.URL{
				Scheme: "file",
				Opaque: "/tmp/",
				Path:   "/tmp/",
			},
		},
		"WithEmptyAuthorityAndQuery": {
			value: "file:///tmp/?mode=ro",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "/tmp/",
				Path:     "/tmp/",
				RawQuery: "mode=ro",
			},
		},
		"HostIsNotEmpty": {
			value: "file://localhost/./tmp/?mode=ro",
			err:   `expected empty host, got "localhost"`,
		},
		"UserIsNotEmpty": {
			value: "file://user:pass@./tmp/?mode=ro",
			err:   `expected empty user info, got "user:pass"`,
		},
		"NoDirectory": {
			value: "file:./nodir/",
			err:   `"./nodir/" should be an existing directory, got stat ./nodir/: no such file or directory`,
		},
		"PathIsNotEndsWithSlash": {
			value: "file:./tmp/file",
			err:   `expected path ending with "/", got ""`,
		},
		"FileInsteadOfDirectory": {
			value: "file:./tmp/file/",
			err:   `"./tmp/file/" should be an existing directory, got stat ./tmp/file/: not a directory`,
		},
		"MalformedURI": {
			value: ":./tmp/",
			err:   `parse ":./tmp/": missing protocol scheme`,
		},
		"NoScheme": {
			value: "./tmp/",
			err:   `expected "file:" schema, got ""`,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u, err := validateURI(tc.value)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, u, tc.uri)
			// assert.Equal(t, tc.value, u.String())
		})
	}
}

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

package pool

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	t.Parallel()

	// tests rely on the fact that both ./tmp and /tmp exist

	err := os.MkdirAll("tmp/dir", 0o777)
	require.NoError(t, err)

	_, err = os.Create("tmp/file")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := os.RemoveAll("tmp")
		require.NoError(t, err)
	})

	testCases := map[string]struct {
		in  string
		uri *url.URL
		out string
		err string
	}{
		"LocalDirectory": {
			in: "file:./",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./",
				Path:     "./",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
			},
			out: "file:./?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
		},
		"LocalSubDirectory": {
			in: "file:./tmp/",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./tmp/",
				Path:     "./tmp/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
			},
			out: "file:./tmp/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
		},
		"LocalSubSubDirectory": {
			in: "file:./tmp/dir/",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./tmp/dir/",
				Path:     "./tmp/dir/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
			},
			out: "file:./tmp/dir/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
		},
		"LocalDirectoryWithParameters": {
			in: "file:./tmp/?mode=memory",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "./tmp/",
				Path:     "./tmp/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29&mode=memory",
			},
			out: "file:./tmp/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29&mode=memory",
		},
		"AbsoluteDirectory": {
			in: "file:/tmp/",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "/tmp/",
				Path:     "/tmp/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
			},
			out: "file:/tmp/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
		},
		"WithEmptyAuthority": {
			in: "file:///tmp/",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "/tmp/",
				Path:     "/tmp/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
			},
			out: "file:/tmp/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29",
		},
		"WithEmptyAuthorityAndQuery": {
			in: "file:///tmp/?mode=memory",
			uri: &url.URL{
				Scheme:   "file",
				Opaque:   "/tmp/",
				Path:     "/tmp/",
				OmitHost: true,
				RawQuery: "_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29&mode=memory",
			},
			out: "file:/tmp/?_pragma=busy_timeout%2810000%29&_pragma=journal_mode%28wal%29&mode=memory",
		},
		"HostIsNotEmpty": {
			in:  "file://localhost/./tmp/?mode=memory",
			err: `expected empty host, got "localhost"`,
		},
		"UserIsNotEmpty": {
			in:  "file://user:pass@./tmp/?mode=memory",
			err: `expected empty user info, got "user:pass"`,
		},
		"NoDirectory": {
			in:  "file:./nodir/",
			err: `"./nodir/" should be an existing directory, got stat ./nodir/: no such file or directory`,
		},
		"PathIsNotEndsWithSlash": {
			in:  "file:./tmp/file",
			err: `expected path ending with "/", got "./tmp/file"`,
		},
		"FileInsteadOfDirectory": {
			in:  "file:./tmp/file/",
			err: `"./tmp/file/" should be an existing directory, got stat ./tmp/file/: not a directory`,
		},
		"MalformedURI": {
			in:  ":./tmp/",
			err: `parse ":./tmp/": missing protocol scheme`,
		},
		"NoScheme": {
			in:  "./tmp/",
			err: `expected "file:" schema, got ""`,
		},
		"Shared": {
			in:  "file:./?cache=shared",
			err: `shared cache is not supported`,
		},
		"SharedMemory": {
			in:  "file:./?mode=memory&cache=shared",
			err: `shared cache is not supported`,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u, err := parseURI(tc.in)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.uri, u)
			assert.Equal(t, tc.out, u.String())
		})
	}
}

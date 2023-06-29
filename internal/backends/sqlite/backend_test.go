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

	t.Cleanup(func() {
		err := os.Remove("tmp/dir")
		require.NoError(t, err)
		err = os.Remove("tmp")
		require.NoError(t, err)
	})

	testCases := map[string]struct {
		params *NewBackendParams

		uri *url.URL
		err error
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
		"LocalSubDirectoryWithParameters": {
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
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uri, err := validateParams(tc.params)
			if err != nil {
				t.Log(err)

				require.Equal(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, uri, tc.uri)
		})
	}
}

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

package testutil

import (
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestSQLiteURI(tb testtb.TB, baseURI string) string {
	tb.Helper()

	require.NotEmpty(tb, baseURI)

	u, err := url.Parse(baseURI)
	require.NoError(tb, err)

	require.True(tb, u.Path == "")
	require.True(tb, u.Opaque != "")

	u.Opaque = path.Join(u.Opaque, DirectoryName(tb)) + "/"
	res := u.String()

	dir, err := filepath.Abs(u.Opaque)
	require.NoError(tb, err)
	require.NoError(tb, os.RemoveAll(dir))
	require.NoError(tb, os.MkdirAll(dir, 0o777))

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s (%s) for debugging.", dir, res)
			return
		}

		require.NoError(tb, os.RemoveAll(dir))
	})

	return res
}

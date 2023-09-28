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

package backends_test // to avoid import cycle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func getBackends(t *testing.T) []backends.Backend {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	l := testutil.Logger(t)

	var res []backends.Backend

	{
		p, err := state.NewProvider("")
		require.NoError(t, err)

		b, err := sqlite.NewBackend(&sqlite.NewBackendParams{
			URI: "file:./?mode=memory",
			L:   l.Named("sqlite"),
			P:   p,
		})
		require.NoError(t, err)
		t.Cleanup(b.Close)

		res = append(res, b)
	}

	return res
}

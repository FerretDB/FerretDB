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

package sql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	rows, err := pool.Query(ctx, "SELECT * FROM pagila.actor ORDER BY actor_id")
	require.NoError(t, err)
	defer rows.Close()

	ri := extractRowInfo(rows)
	assert.Equal(t, []string{"actor_id", "first_name", "last_name", "last_update"}, ri.names)

	doc, err := nextRow(rows, ri)
	require.NoError(t, err)
	require.NotNil(t, doc)

	expected := types.MustMakeDocument(
		"actor_id", int32(1),
		"first_name", "PENELOPE",
		"last_name", "GUINESS",
		"last_update", time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local(),
	)
	assert.Equal(t, &expected, doc)
}

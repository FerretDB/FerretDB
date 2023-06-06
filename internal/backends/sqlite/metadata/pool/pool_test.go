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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateDrop(t *testing.T) {
	ctx := testutil.Ctx(t)

	p, err := New(t.TempDir(), testutil.Logger(t))
	require.NoError(t, err)

	defer p.Close()

	db := p.GetExisting(ctx, t.Name())
	require.Nil(t, db)

	db, created, err := p.GetOrCreate(ctx, t.Name())
	require.NoError(t, err)
	require.NotNil(t, db)
	require.True(t, created)

	db2, created, err := p.GetOrCreate(ctx, t.Name())
	require.NoError(t, err)
	require.Same(t, db, db2)
	require.False(t, created)

	db2 = p.GetExisting(ctx, t.Name())
	require.Same(t, db, db2)

	dropped := p.Drop(ctx, t.Name())
	require.True(t, dropped)

	dropped = p.Drop(ctx, t.Name())
	require.False(t, dropped)

	db = p.GetExisting(ctx, t.Name())
	require.Nil(t, db)
}

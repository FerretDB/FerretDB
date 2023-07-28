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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(r.Close)

	_, err = r.DatabaseGetOrCreate(ctx, t.Name())
	require.NoError(t, err)

	created, err := r.CollectionCreate(ctx, t.Name(), t.Name())
	require.NoError(t, err)
	require.True(t, created)

	dropped := r.DatabaseDrop(ctx, t.Name())
	require.True(t, dropped)
}

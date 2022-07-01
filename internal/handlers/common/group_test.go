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

package common

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupContext(t *testing.T) {
	t.Parallel()

	ctx := NewGroupContext()
	require.NotNil(t, ctx)

	ctx.AddField("_id", "1")
	ctx.AddField("count", "COUNT(*)")

	assert.Equal(t, ctx.FieldAsString(), "json_build_object('$k', jsonb_build_array('_id', 'count'), 'count', COUNT(*), '_id', 1) AS _jsonb")
}

func TestUnique(t *testing.T) {
	t.Parallel()

	ctx := NewGroupContext()
	require.NotNil(t, ctx)

	group := must.NotFail(types.NewDocument("_id", "$item"))

	err := ParseGroup(&ctx, "", group)
	require.NoError(t, err)

	assert.Equal(t, "DISTINCT ON (_jsonb->'item') json_build_object('$k', jsonb_build_array('_id'), '_id', _jsonb->'item') AS _jsonb", ctx.FieldAsString())
}

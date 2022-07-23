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

package aggregate

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(1)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.NoError(t, err)

	stage := stages[0]
	assert.Equal(t, "field", stage.sortFields[0].name)
	assert.Equal(t, SortAsc, stage.sortFields[0].dir)
	assert.Equal(t, `"field"`, stage.SortToSql(false))
	assert.Equal(t, "_jsonb->'field'", stage.SortToSql(true))
}

func TestDescendingSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(-1)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.NoError(t, err)

	stage := stages[0]
	assert.Equal(t, "field", stage.sortFields[0].name)
	assert.Equal(t, SortDesc, stage.sortFields[0].dir)
	assert.Equal(t, `"field" DESC`, stage.SortToSql(false))
	assert.Equal(t, "_jsonb->'field' DESC", stage.SortToSql(true))
}

func TestFieldWithSpacesSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("The Name", int32(1)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.NoError(t, err)

	stage := stages[0]
	assert.Equal(t, "The Name", stage.sortFields[0].name)
	assert.Equal(t, SortAsc, stage.sortFields[0].dir)
	assert.Equal(t, `"The Name"`, stage.SortToSql(false))
	assert.Equal(t, "_jsonb->'The Name'", stage.SortToSql(true))
}

func TestFieldWithSpacesDescSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("The Name", int32(-1)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.NoError(t, err)

	stage := stages[0]
	assert.Equal(t, "The Name", stage.sortFields[0].name)
	assert.Equal(t, SortDesc, stage.sortFields[0].dir)
	assert.Equal(t, `"The Name" DESC`, stage.SortToSql(false))
	assert.Equal(t, "_jsonb->'The Name' DESC", stage.SortToSql(true))
}

func TestCompositeSortOrder(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field1", int32(-1), "field2", int32(1), "field3", int32(1)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.NoError(t, err)

	stage := stages[0]
	assert.Equal(t, "field1", stage.sortFields[0].name)
	assert.Equal(t, SortDesc, stage.sortFields[0].dir)
	assert.Equal(t, "field2", stage.sortFields[1].name)
	assert.Equal(t, SortAsc, stage.sortFields[1].dir)
	assert.Equal(t, "field3", stage.sortFields[2].name)
	assert.Equal(t, SortAsc, stage.sortFields[2].dir)
	assert.Equal(t, `"field1" DESC, "field2", "field3"`, stage.SortToSql(false))
	assert.Equal(t, "_jsonb->'field1' DESC, _jsonb->'field2', _jsonb->'field3'", stage.SortToSql(true))
}

func TestInvalidSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(0)))
	stages := []*Stage{}
	err := AddSortStage(&stages, doc)
	require.Error(t, err)

	assert.Equal(t, "invalid sort order: 0", err.Error())
}

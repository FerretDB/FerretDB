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

func TestInclusionProjectionType(t *testing.T) {
	t.Parallel()

	project := must.NotFail(types.NewDocument("field1", int32(1), "field2", int32(2), "field3", int32(3)))

	type_, err := CheckProjectionType(project)
	assert.NoError(t, err)

	assert.Equal(t, "inclusion", type_)
}

func TestExclusionProjectionType(t *testing.T) {
	t.Parallel()

	project := must.NotFail(types.NewDocument("field1", int32(0), "field2", int32(0)))

	type_, err := CheckProjectionType(project)
	assert.NoError(t, err)

	assert.Equal(t, "exclusion", type_)
}

func TestInvalidProjectionType(t *testing.T) {
	t.Parallel()

	project := must.NotFail(types.NewDocument("field1", int32(0), "field2", int32(1)))

	_, err := CheckProjectionType(project)
	require.Error(t, err)

	assert.Equal(t, "Invalid $project :: caused by :: Cannot do exclusion on field field2 in exclusion projection", err.Error())
}

func TestInclusionProjection(t *testing.T) {
	t.Parallel()

	stage := NewEmptyStage("match")
	stage.AddField("field1", "int", "1")
	stages := []*Stage{&stage}
	project := must.NotFail(types.NewDocument("field1", int32(1), "field2", int32(2), "field3", int32(3)))

	fieldsRef, err := ParseProjectStage(&stages, project)
	assert.NoError(t, err)

	fields := *fieldsRef
	assert.Equal(t, len(fields), 1)
	assert.Equal(t, "field1", fields[0].name)
	assert.Equal(t, "field1", fields[0].alias)
}

func TestExclusionProjection(t *testing.T) {
	t.Parallel()

	stage := NewEmptyStage("match")
	stage.AddField("field1", "int", "1")
	stage.AddField("field2", "int", "1")
	stage.AddField("field3", "int", "1")
	stages := []*Stage{&stage}
	project := must.NotFail(types.NewDocument("field2", int32(0)))

	fieldsRef, err := ParseProjectStage(&stages, project)
	assert.NoError(t, err)

	fields := *fieldsRef
	assert.Equal(t, len(fields), 2)
	assert.Equal(t, "field1", fields[0].name)
	assert.Equal(t, "field1", fields[0].alias)
	assert.Equal(t, "field3", fields[1].name)
	assert.Equal(t, "field3", fields[1].alias)
}

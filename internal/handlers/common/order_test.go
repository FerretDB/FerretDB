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

func TestSimpleSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(1)))
	sort, err := AggregateSort(doc)
	require.NoError(t, err)

	assert.Equal(t, "field", *sort)
}

func TestDescendingSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(-1)))
	sort, err := AggregateSort(doc)
	require.NoError(t, err)

	assert.Equal(t, "field DESC", *sort)
}

func TestCompositeSortOrder(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field1", int32(-1), "field2", int32(1), "field3", int32(1)))
	sort, err := AggregateSort(doc)
	require.NoError(t, err)

	assert.Equal(t, "field1 DESC, field2, field3", *sort)
}

func TestInvalidSort(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field", int32(0)))
	_, err := AggregateSort(doc)
	require.Error(t, err)

	assert.Equal(t, "invalid sort order: 0", err.Error())
}

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

func TestFieldToSql(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "_jsonb", FieldToSql(""))
	assert.Equal(t, "_jsonb->'quantity'", FieldToSql("quantity"))
	assert.Equal(t, "_jsonb->>'item'->'quantity'", FieldToSql("item.quantity"))
	assert.Equal(t, "_jsonb->>'item'->>'quantity'->'today'", FieldToSql("item.quantity.today"))
}

func TestParseField(t *testing.T) {
	t.Parallel()

	field, parents := ParseField("item.quantity")
	assert.Equal(t, "quantity", field)
	assert.Equal(t, "item", parents)

	field, parents = ParseField("order.item.quantity")
	assert.Equal(t, "quantity", field)
	assert.Equal(t, "order.item", parents)

	field, parents = ParseField("quantity")
	assert.Equal(t, "quantity", field)
	assert.Equal(t, "", parents)
}

func TestSimpleMatchStage(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("quantity", int32(1)))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "quantity", filter.field)
	assert.Equal(t, "=", filter.op)
	assert.Equal(t, int32(1), filter.value)
}

func TestComplexMatchStage(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("quantity",
		must.NotFail(types.NewDocument("$gt", int32(1))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "quantity", filter.field)
	assert.Equal(t, ">", filter.op)
	assert.Equal(t, int32(1), filter.value)
}

func TestNestedMatchStage(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("item.quantity",
		must.NotFail(types.NewDocument("$gt", int32(1))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "item.quantity", filter.field)
	assert.Equal(t, ">", filter.op)
	assert.Equal(t, int32(1), filter.value)
}

func TestToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("quantity",
		must.NotFail(types.NewDocument("$gt", int32(1))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb->'quantity' > $1", filter.ToSql())
}

func TestNestedToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("item.quantity",
		must.NotFail(types.NewDocument("$gt", int32(1))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb->>'item'->'quantity' > $1", filter.ToSql())
}

func TestAndOrToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("$or",
		must.NotFail(types.NewArray(
			must.NotFail(types.NewDocument("$and",
				must.NotFail(types.NewArray(
					must.NotFail(types.NewDocument("item.quantity",
						must.NotFail(types.NewDocument("$gt", int32(1))),
					)),
					must.NotFail(types.NewDocument("daysToExp",
						must.NotFail(types.NewDocument("$lte", int32(10))),
					)),
				)),
			)),
			must.NotFail(types.NewDocument("valid", true)),
		)),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "((_jsonb->>'item'->'quantity' > $1 AND _jsonb->'daysToExp' <= $2) OR _jsonb->'valid' = $3)", filter.ToSql())
	assert.Equal(t, []interface{}{int32(1), int32(10), true}, stage.GetValues())
}

func TestExistsToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field",
		must.NotFail(types.NewDocument("$exists", true)),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb ? $1", filter.ToSql())
	assert.Equal(t, []interface{}{"field"}, stage.GetValues())
}

func TestNotExistsToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field",
		must.NotFail(types.NewDocument("$exists", false)),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "NOT (_jsonb ? $1)", filter.ToSql())
	assert.Equal(t, []interface{}{"field"}, stage.GetValues())
}

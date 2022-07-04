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

	assert.Equal(t, "_jsonb", FieldToSql("", false))
	assert.Equal(t, "_jsonb->'quantity'", FieldToSql("quantity", false))
	assert.Equal(t, "_jsonb->>'quantity'", FieldToSql("quantity", true))
	assert.Equal(t, "_jsonb->'item'->>'quantity'", FieldToSql("item.quantity", false))
	assert.Equal(t, "_jsonb->'item'->'quantity'->>'today'", FieldToSql("item.quantity.today", false))
	assert.Equal(t, "_jsonb->'item'->'quantity'->>'today'", FieldToSql("item.quantity.today", true))
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
	assert.Equal(t, "quantity > $1", filter.ToSql(false))
	assert.Equal(t, "_jsonb->'quantity' > $1", filter.ToSql(true))
}

func TestNestedToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("item.quantity",
		must.NotFail(types.NewDocument("$gt", int32(1))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb->>'item'->'quantity' > $1", filter.ToSql(true))
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
						must.NotFail(types.NewDocument("$not",
							must.NotFail(types.NewDocument("$lte", int32(10))),
						)),
					)),
				)),
			)),
			must.NotFail(types.NewDocument("valid", true)),
		)),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "((_jsonb->>'item'->'quantity' > $1 AND NOT (_jsonb->'daysToExp' <= $2)) OR _jsonb->'valid' = $3)", filter.ToSql(true))
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
	assert.Equal(t, "_jsonb ? $1", filter.ToSql(true))
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
	assert.Equal(t, "NOT (_jsonb ? $1)", filter.ToSql(true))
	assert.Equal(t, []interface{}{"field"}, stage.GetValues())
}

func TestInToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field",
		must.NotFail(types.NewDocument("$in", must.NotFail(types.NewArray(int32(1), int32(2), int32(3))))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb->'field' = ANY($1)", filter.ToSql(true))
	assert.Equal(t, []interface{}{[]string{"1", "2", "3"}}, stage.GetValues())
}

func TestNotInToSql(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(types.NewDocument("field",
		must.NotFail(types.NewDocument("$nin", must.NotFail(types.NewArray(int32(1), int32(2), int32(3))))),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "NOT (_jsonb->'field' = ANY($1))", filter.ToSql(true))
	assert.Equal(t, []interface{}{[]string{"1", "2", "3"}}, stage.GetValues())
}

func TestRegexSql(t *testing.T) {
	doc := must.NotFail(types.NewDocument("color",
		must.NotFail(types.NewDocument("$regex", "^e.*")),
	))
	stage, err := ParseMatchStage(doc)
	require.NoError(t, err)

	filter := stage.root.children[0]
	assert.Equal(t, "_jsonb->>'color' ~ $1", filter.ToSql(true))
	assert.Equal(t, []interface{}{"^e.*"}, stage.GetValues())
}

func TestGetValuesInt32(t *testing.T) {
	t.Parallel()

	node := NewFieldFilterNode(0, "field", ">", int32(1), nil, false)
	assert.Equal(t, []interface{}{int32(1)}, node.GetValues())
}

func TestGetValuesFloat64(t *testing.T) {
	t.Parallel()

	node := NewFieldFilterNode(0, "field", ">", float64(1.5), nil, false)
	assert.Equal(t, []interface{}{"(CASE WHEN (1.5 ? '$f') THEN (1.5->>'$f')::numeric ELSE (1.5)::numeric END)"}, node.GetValues())
}

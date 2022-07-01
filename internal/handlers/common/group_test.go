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
	"fmt"
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

	assert.Equal(t, ctx.FieldAsString(), "json_build_object('$k', jsonb_build_array('_id', 'count'), '_id', 1, 'count', COUNT(*)) AS _jsonb")
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

func TestCountSumAndAverage(t *testing.T) {
	t.Parallel()

	ctx := NewGroupContext()
	require.NotNil(t, ctx)

	dateToString := must.NotFail(types.NewDocument("date", "$date", "format", "%Y-%m-%d"))
	id := must.NotFail(types.NewDocument("$dateToString", dateToString))

	// FIXME improve sale amount here
	// multiply := must.NotFail(types.NewArray("$item", "$price"))
	// sum := must.NotFail(types.NewDocument("$multiply", multiply))
	// totalSaleAmount := must.NotFail(types.NewDocument("$sum", sum))

	totalSaleAmount := must.NotFail(types.NewDocument("$sum", "$quantity"))
	averageQuantity := must.NotFail(types.NewDocument("$avg", "$quantity"))
	count := must.NotFail(types.NewDocument("$sum", int32(1)))

	group := must.NotFail(types.NewDocument(
		"_id", id,
		"totalSaleAmount", totalSaleAmount,
		"averageQuantity", averageQuantity,
		"count", count,
	))

	err := ParseGroup(&ctx, "", group)
	require.NoError(t, err)

	assert.Equal(t, "json_build_object('$k', jsonb_build_array('_id', 'totalSaleAmount', 'averageQuantity', 'count'), '_id', _id, 'totalSaleAmount', json_build_object('$f', totalSaleAmount), 'averageQuantity', json_build_object('$f', averageQuantity), 'count', json_build_object('$f', count)) AS _jsonb", ctx.FieldAsString())
	assert.Equal(t, "SELECT TO_CHAR(TO_TIMESTAMP((_jsonb->'date'->>'$d')::numeric / 1000), 'YYYY-MM-DD') AS _id, SUM((CASE WHEN (_jsonb->'quantity' ? '$f') THEN (_jsonb->'quantity'->>'$f')::numeric ELSE (_jsonb->'quantity')::numeric END)) AS totalSaleAmount, AVG((CASE WHEN (_jsonb->'quantity' ? '$f') THEN (_jsonb->'quantity'->>'$f')::numeric ELSE (_jsonb->'quantity')::numeric END)) AS averageQuantity, SUM(1) AS count FROM %s GROUP BY _id", ctx.GetSubQuery())
}

func TestSumWithNumber(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		sumValue interface{}
		expected string
	}{
		"Int32": {
			sumValue: int32(1),
			expected: "1",
		},
		"Int64": {
			sumValue: int64(2),
			expected: "2",
		},
		"Float64": {
			sumValue: float64(0.5),
			expected: "0.5",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := NewGroupContext()
			require.NotNil(t, ctx)

			group := must.NotFail(types.NewDocument(
				"_id", "$item",
				"totalSaleAmount", must.NotFail(types.NewDocument("$sum", tc.sumValue)),
			))

			err := ParseGroup(&ctx, "", group)
			require.NoError(t, err)

			assert.Equal(t, "DISTINCT ON (_jsonb->'item') json_build_object('$k', jsonb_build_array('_id', 'totalSaleAmount'), '_id', _jsonb->'item', 'totalSaleAmount', json_build_object('$f', totalSaleAmount)) AS _jsonb", ctx.FieldAsString())
			assert.Equal(t, fmt.Sprintf("SELECT SUM(%s) AS totalSaleAmount FROM %%s", tc.expected), ctx.GetSubQuery())
		})
	}
}

func TestSumWithOperators(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		sumDoc   *types.Document
		expected string
	}{
		"Int32": {
			sumDoc: must.NotFail(types.NewDocument("$sum",
				must.NotFail(types.NewDocument("$multiply",
					must.NotFail(types.NewArray("$quantity", "$price")),
				)),
			)),
			expected: "SELECT SUM(((CASE WHEN (_jsonb->'quantity' ? '$f') THEN (_jsonb->'quantity'->>'$f')::numeric ELSE (_jsonb->'quantity')::numeric END) * (CASE WHEN (_jsonb->'price' ? '$f') THEN (_jsonb->'price'->>'$f')::numeric ELSE (_jsonb->'price')::numeric END))) AS totalSaleAmount FROM %s",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := NewGroupContext()
			require.NotNil(t, ctx)

			group := must.NotFail(types.NewDocument(
				"_id", "$item",
				"totalSaleAmount", tc.sumDoc,
			))

			err := ParseGroup(&ctx, "", group)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, ctx.GetSubQuery())
			assert.Equal(t, "DISTINCT ON (_jsonb->'item') json_build_object('$k', jsonb_build_array('_id', 'totalSaleAmount'), '_id', _jsonb->'item', 'totalSaleAmount', json_build_object('$f', totalSaleAmount)) AS _jsonb", ctx.FieldAsString())
		})
	}
}

func TestCountSumAsArrayError(t *testing.T) {
	t.Parallel()

	ctx := NewGroupContext()
	require.NotNil(t, ctx)

	dateToString := must.NotFail(types.NewDocument("date", "$date", "format", "%Y-%m-%d"))
	id := must.NotFail(types.NewDocument("$dateToString", dateToString))
	sumParts := must.NotFail(types.NewArray("$item", "$price"))
	totalSaleAmount := must.NotFail(types.NewDocument("$sum", sumParts))
	group := must.NotFail(types.NewDocument(
		"_id", id,
		"totalSaleAmount", totalSaleAmount,
	))

	err := ParseGroup(&ctx, "", group)
	assert.Equal(t, "The $sum accumulator is a unary operator,", err.Error())
}

func TestMultiplyOperator(t *testing.T) {
	t.Parallel()

	ctx := NewGroupContext()
	require.NotNil(t, ctx)

	multiply := must.NotFail(types.NewArray("$quantity", "$price"))
	multiplyDoc := must.NotFail(types.NewDocument("$multiply", multiply))

	groupCtx := NewGroupContext()
	parsed, err := ParseOperators(&groupCtx, multiplyDoc)
	require.NoError(t, err)

	qty := "(CASE WHEN (_jsonb->'quantity' ? '$f') THEN (_jsonb->'quantity'->>'$f')::numeric ELSE (_jsonb->'quantity')::numeric END)"
	price := "(CASE WHEN (_jsonb->'price' ? '$f') THEN (_jsonb->'price'->>'$f')::numeric ELSE (_jsonb->'price')::numeric END)"
	assert.Equal(t, fmt.Sprintf("(%s * %s)", qty, price), *parsed)
}

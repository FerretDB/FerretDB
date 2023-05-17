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

// Package operators provides aggregation operators.
package operators

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type typeOp struct {
	expression types.Expression
}

func newType(accumulation *types.Document) (Accumulator, error) {
	typeParam, err := common.GetRequiredParam[string](accumulation, "$type")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"TODO",
			"$type (accumulator)",
		)
	}

	expression, err := types.NewExpression(typeParam)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"TODO",
			"$type (accumulator)",
		)
	}

	return &typeOp{
		expression: expression,
	}, nil
}

func (t *typeOp) Accumulate(ctx context.Context, groupID any, grouped []*types.Document) (any, error) {
	if t.expression != nil {
	}

	var resTypes []any

	for _, doc := range grouped {
		v := t.expression.Evaluate(doc)
		var resDoc *types.Document

		if v == nil {
			resDoc = must.NotFail(types.NewDocument(t.expression.GetExpressionSuffix(), "missing"))
		} else {
			resDoc = must.NotFail(types.NewDocument(t.expression.GetExpressionSuffix(), sjson.GetTypeOfValue(v)))
		}

		resTypes = append(resTypes, resDoc)
	}

	return resTypes, nil
}

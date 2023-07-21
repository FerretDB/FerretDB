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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// typeOp represents `$type` operator.
type typeOp struct {
	param any
}

// newType returns `$type` operator.
func newType(operation *types.Document) (Operator, error) {
	param := must.NotFail(operation.Get("$type"))

	return &typeOp{
		param: param,
	}, nil
}

// Process implements Operator interface.
func (t *typeOp) Process(doc *types.Document) (any, error) {
	typeParam := t.param

	var paramEvaluated bool

	var res any

	for !paramEvaluated {
		paramEvaluated = true

		switch param := typeParam.(type) {
		case *types.Document:
			if !IsOperator(param) {
				res = param
				break
			}

			operator, err := NewOperator(param)
			if err != nil {
				var opErr OperatorError
				if !errors.As(err, &opErr) {
					return nil, lazyerrors.Error(err)
				}

				if opErr.Code() == ErrInvalidExpression {
					opErr.code = ErrInvalidNestedExpression
				}

				return nil, opErr
			}

			if typeParam, err = operator.Process(doc); err != nil {
				var opErr OperatorError
				if !errors.As(err, &opErr) {
					return nil, lazyerrors.Error(err)
				}

				return nil, err
			}

			// the result of nested operator needs to be evaluated
			paramEvaluated = false

		case *types.Array:
			if param.Len() != 1 {
				return nil, newOperatorError(
					ErrArgsInvalidLen,
					fmt.Sprintf("Expression $type takes exactly 1 arguments. %d were passed in.", param.Len()),
				)
			}

			value, err := param.Get(0)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			res = value

		case float64, types.Binary, types.ObjectID, bool, time.Time,
			types.NullType, types.Regex, int32, types.Timestamp, int64:
			res = param

		case string:
			if strings.HasPrefix(param, "$") {
				expression, err := aggregations.NewExpression(param, nil)
				if err != nil {
					return nil, err
				}

				value, err := expression.Evaluate(doc)
				if err != nil {
					return "missing", nil
				}

				res = value

				continue
			}

			res = param

		default:
			panic(fmt.Sprint("wrong type of value: ", typeParam))
		}
	}

	return commonparams.AliasFromType(res), nil
}

// check interfaces
var (
	_ Operator = (*typeOp)(nil)
)

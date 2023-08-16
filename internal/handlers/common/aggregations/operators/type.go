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
)

// typeOp represents `$type` operator.
type typeOp struct {
	param any

	operator Operator
}

// newType returns `$type` operator.
func newType(args ...any) (Operator, error) {
	if len(args) != 1 {
		return nil, newOperatorError(
			ErrArgsInvalidLen,
			"$type",
			fmt.Sprintf("Expression $type takes exactly 1 arguments. %d were passed in.", len(args)),
		)
	}

	operator := new(typeOp)

	switch param := args[0].(type) {
	case *types.Document:
		if !IsOperator(param) {
			operator.param = param
			break
		}

		op, err := NewOperator(param)
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

		operator.operator = op

	case *types.Array, string, float64, types.Binary, types.ObjectID, bool, time.Time,
		types.NullType, types.Regex, int32, types.Timestamp, int64:
		operator.param = param

	default:
		panic(fmt.Sprint("wrong type of value: ", param))
	}

	return operator, nil
}

// Process implements Operator interface.
func (t *typeOp) Process(doc *types.Document) (any, error) {
	typeParam := t.param

	if t.operator != nil {
		var err error
		typeParam, err = t.operator.Process(doc)

		if err != nil {
			var opErr OperatorError
			if !errors.As(err, &opErr) {
				return nil, lazyerrors.Error(err)
			}

			return nil, err
		}
	}

	var res any

	switch param := typeParam.(type) {
	case *types.Document, *types.Array, float64, types.Binary, types.ObjectID, bool, time.Time,
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
			break
		}

		res = param

	default:
		panic(fmt.Sprint("wrong type of value: ", typeParam))
	}

	return commonparams.AliasFromType(res), nil
}

// check interfaces
var (
	_ Operator = (*typeOp)(nil)
)

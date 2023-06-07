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

type typeOp struct {
	param any
}

func newType(operation *types.Document) (Operator, error) {
	param := must.NotFail(operation.Get("$type"))

	return &typeOp{
		param: param,
	}, nil
}

func (t *typeOp) Process(doc *types.Document) (any, error) {
	var value any

	typeParam := t.param

	var paramEvaluated bool

	for !paramEvaluated {
		paramEvaluated = true
		switch param := typeParam.(type) {
		case *types.Document:
			operator, err := NewOperator(param)
			if errors.Is(err, ErrNoOperator) {
				value = param
				continue
			}

			if err != nil {
				return nil, err
			}

			if typeParam, err = operator.Process(doc); err != nil {
				return nil, lazyerrors.Error(err)
			}

			// the result of nested operator needs to be evaluated
			paramEvaluated = false

		case string:
			if strings.HasPrefix(param, "$") {
				expression, err := aggregations.NewExpression(param)
				if err != nil {
					return nil, err
				}

				value = expression.Evaluate(doc)
				continue
			}

			value = param

		case *types.Array, float64, types.Binary, types.ObjectID, bool, time.Time, types.NullType, types.Regex, int32, types.Timestamp, int64:
			value = param
		default:
			panic(fmt.Sprint("wrong type of value: ", typeParam))
		}
	}

	var res string
	if value == nil {
		res = "missing"
	} else {
		res = commonparams.AliasFromType(value)
	}

	return res, nil
}

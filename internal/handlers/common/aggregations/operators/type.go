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
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
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

func (t *typeOp) Process(ctx context.Context, doc *types.Document) (any, error) {
	var value any

	typeParam := t.param
	for i := 0; i <= 1; i++ {

		switch param := typeParam.(type) {
		case *types.Document:
			operator, err := Get(param)
			if err != nil {
				panic("TODO")
			}
			i--

			if typeParam, err = operator.Process(context.TODO(), doc); err != nil {
				panic("TODO")
			}

		case *types.Array:
			panic("TODO")
		case string:
			if strings.HasPrefix("$", param) {
				expression, err := aggregations.NewExpression(param)
				if err != nil {
					return nil, err
				}

				value = expression.Evaluate(doc)
				continue
			}
			value = param

		case float64, types.Binary, types.ObjectID, bool, time.Time, types.NullType, types.Regex, int32, types.Timestamp, int64:
			value = param
		default:
			panic(fmt.Sprint("wrong type of value: ", typeParam))
		}

	}

	var res string
	if value == nil {
		res = "missing"
	} else {
		res = sjson.GetTypeOfValue(value)
	}

	return res, nil
}

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

package accumulators

import (
	"errors"
	"reflect"
	"sort"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// first represents $first aggregation operator.
type first struct {
	expression  *aggregations.Expression
	resultFound bool
	result      any
}

// newFirst creates a new $first aggregation operator.
func newFirst(args ...any) (Accumulator, error) {
	accumulator := new(first)

	if len(args) != 1 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupUnaryOperator,
			"The $first accumulator is a unary operator",
			"$first (accumulator)",
		)
	}

	arg := args[0]
	switch arg := arg.(type) {
	case string:
		var err error
		if accumulator.expression, err = aggregations.NewExpression(arg, nil); err != nil {
			return nil, err
		}
	case *types.Array:
		// No need to iterate over all the documents
		accumulator.resultFound = true
		if arg.Len() == 0 {
			break
		}
		var err error
		accumulator.result, err = arg.Get(0)
		if err != nil {
			return nil, err
		}
	default:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidArg,
			"Argument should be a string",
			"$first (accumulator)")
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (f *first) Accumulate(iter types.DocumentsIterator) (any, error) {
	if f.resultFound {
		return f.result, nil
	}

	var values []any

	for {
		_, doc, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		value, err := f.expression.Evaluate(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		values = append(values, value)
	}

	sort.SliceStable(values, func(i, j int) bool {
		v1 := values[i]
		v2 := values[i]

		if reflect.TypeOf(v1) == reflect.TypeOf(v2) {
			// Types can't be compared
			return false
		}

		switch v1 := v1.(type) {
		case int:
			return v1 < v2.(int)
		case float64:
			return v1 < v2.(float64)
		case string:
			return v1 < v2.(string)
		default:
			return false
		}
	})

	return values[0], nil
}

// check interfaces
var (
	_ Accumulator = (*first)(nil)
)

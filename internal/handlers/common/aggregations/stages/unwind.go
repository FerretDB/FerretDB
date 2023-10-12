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

package stages

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonpath"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// unwind represents $unwind stage.
type unwind struct {
	field *aggregations.Expression
}

// newUnwind creates a new $unwind stage.
func newUnwind(stage *types.Document) (aggregations.Stage, error) {
	field, err := stage.Get("$unwind")
	if err != nil {
		return nil, err
	}

	var expr *aggregations.Expression

	switch field := field.(type) {
	case *types.Document:
		return nil, common.Unimplemented(stage, "$unwind")
	case string:
		if field == "" {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageUnwindNoPath,
				"no path specified to $unwind stage",
				"$unwind (stage)",
			)
		}

		// For $unwind to deconstruct an array from dot notation, array must be at the suffix.
		// It returns empty result if array is found at other parts of dot notation,
		// so it does not return value by index of array nor values for given key in array's document.
		expr, err = aggregations.NewExpression(field, &commonpath.FindValuesOpts{
			FindArrayIndex:     false,
			FindArrayDocuments: false,
		})

		if err != nil {
			var exprErr *aggregations.ExpressionError
			if !errors.As(err, &exprErr) {
				return nil, lazyerrors.Error(err)
			}

			switch exprErr.Code() {
			case aggregations.ErrNotExpression:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrStageUnwindNoPrefix,
					fmt.Sprintf("path option to $unwind stage should be prefixed with a '$': %v", types.FormatAnyValue(field)),
					"$unwind (stage)",
				)
			case aggregations.ErrEmptyFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrEmptyFieldPath,
					"Expression cannot be constructed with empty string",
					"$unwind (stage)",
				)
			case aggregations.ErrEmptyVariable, aggregations.ErrInvalidExpression, aggregations.ErrUndefinedVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFieldPathInvalidName,
					"Expression field names may not start with '$'. Consider using $getField or $setField",
					"$unwind (stage)",
				)
			default:
				return nil, lazyerrors.Error(err)
			}
		}
	default:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageUnwindWrongType,
			fmt.Sprintf(
				"expected either a string or an object as specification for $unwind stage, got %s",
				types.FormatAnyValue(field),
			),
			"$unwind (Stage)",
		)
	}

	return &unwind{
		field: expr,
	}, nil
}

// Process implements Stage interface.
func (u *unwind) Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	// TODO https://github.com/FerretDB/FerretDB/issues/2490
	docs, err := iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var out []*types.Document

	if u.field == nil {
		return nil, nil
	}

	key := u.field.GetExpressionSuffix()

	for _, doc := range docs {
		d, err := u.field.Evaluate(doc)
		if err != nil {
			// Ignore non-existent values
			continue
		}

		switch d := d.(type) {
		case *types.Array:
			iter := d.Iterator()
			defer iter.Close()

			id := must.NotFail(doc.Get("_id"))

			for {
				_, v, err := iter.Next()
				if err != nil {
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}

					return nil, err
				}

				newDoc := must.NotFail(types.NewDocument("_id", id, key, v))
				out = append(out, newDoc)
			}
		case types.NullType:
			// Ignore Nulls
		default:
			out = append(out, doc)
		}
	}

	iter = iterator.Values(iterator.ForSlice(out))
	closer.Add(iter)

	return iter, nil
}

// check interfaces
var (
	_ aggregations.Stage = (*unwind)(nil)
)

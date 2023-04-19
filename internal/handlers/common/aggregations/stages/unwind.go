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
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// unwind represents $unwind stage.
type unwind struct {
	field types.Expression
}

// newUnwind creates a new $unwind stage.
func newUnwind(stage *types.Document) (Stage, error) {
	field, err := stage.Get("$unwind")
	if err != nil {
		return nil, err
	}

	var expr types.Expression

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

		opts := types.ExpressionOpts{
			IgnoreArrays: true,
		}
		expr, err = types.NewExpressionWithOpts(field, &opts)

		if err != nil {
			var fieldPathErr *types.FieldPathError
			if !errors.As(err, &fieldPathErr) {
				return nil, lazyerrors.Error(err)
			}

			switch fieldPathErr.Code() {
			case types.ErrNotFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrStageUnwindNoPrefix,
					fmt.Sprintf("path option to $unwind stage should be prefixed with a '$': %v", types.FormatAnyValue(field)),
					"$unwind (stage)",
				)
			case types.ErrEmptyFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrEmptyFieldPath,
					"FieldPath cannot be constructed with empty string",
					"$unwind (stage)",
				)
			case types.ErrEmptyVariable, types.ErrInvalidFieldPath, types.ErrUndefinedVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFieldPathInvalidName,
					"FieldPath field names may not start with '$'. Consider using $getField or $setField",
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
func (u *unwind) Process(ctx context.Context, iter types.DocumentsIterator) (types.DocumentsIterator, error) {
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
		d := u.field.Evaluate(doc)
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

	return iterator.Values(iterator.ForSlice(out)), nil
}

// Type implements Stage interface.
func (u *unwind) Type() StageType {
	return StageTypeDocuments
}

// check interfaces
var (
	_ Stage = (*unwind)(nil)
)

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

package aggregations

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
		expr, err = types.NewExpression(field)
		if err != nil {
			var fieldPathErr *types.FieldPathError
			if !errors.As(err, &fieldPathErr) {
				return nil, lazyerrors.Error(err)
			}

			switch fieldPathErr.Code() {
			case types.ErrNotFieldPath:
			case types.ErrEmptyFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrGroupInvalidFieldPath,
					"'$' by itself is not a valid FieldPath",
					"$group (stage)",
				)
			case types.ErrInvalidFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					fmt.Sprintf("'%s' starts with an invalid character for a user variable name", types.FormatAnyValue(field)),
					"$group (stage)",
				)
			case types.ErrEmptyVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					"empty variable names are not allowed",
					"$group (stage)",
				)
			case types.ErrUndefinedVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrGroupUndefinedVariable,
					fmt.Sprintf("Use of undefined variable: %s", types.FormatAnyValue(field)),
					"$group (stage)",
				)
			default:
				panic(fmt.Sprintf("unhandled field path error %s", fieldPathErr.Error()))
			}

		}
	}

	return &unwind{
		field: expr,
	}, nil
}

func (m *unwind) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	var out []*types.Document
	key := m.field.GetPath().Suffix()

	for _, doc := range in {
		d := m.field.Evaluate(doc)
		switch d := d.(type) {
		//case *types.Document:
		case nil:
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

		default:
			continue
		}
	}

	return out, nil
}

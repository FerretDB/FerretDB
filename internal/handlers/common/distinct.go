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

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// DistinctParams contains `distinct` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type DistinctParams struct {
	DB         string
	Collection string
	Key        string
	Filter     *types.Document
	Comment    string
}

// GetDistinctParams returns `distinct` command parameters.
func GetDistinctParams(document *types.Document, l *zap.Logger) (*DistinctParams, error) {
	var err error

	unimplementedFields := []string{
		"collation",
	}
	if err = Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"readConcern",
	}
	Ignored(document, l, ignoredFields...)

	var dp DistinctParams

	if dp.DB, err = GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if dp.Collection, ok = collectionParam.(string); !ok {
		return nil, NewCommandErrorMsgWithArgument(
			ErrInvalidNamespace,
			fmt.Sprintf("collection name has invalid type %s", AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	if dp.Key, err = GetRequiredParam[string](document, "key"); err != nil {
		return nil, err
	}

	if dp.Key == "" {
		return nil, NewCommandErrorMsg(ErrEmptyFieldPath, "FieldPath cannot be constructed with empty string")
	}

	if dp.Filter, err = GetOptionalParam(document, "query", dp.Filter); err != nil {
		return nil, err
	}

	if dp.Comment, err = GetOptionalParam(document, "comment", dp.Comment); err != nil {
		return nil, err
	}

	return &dp, nil
}

// FilterDistinctValues returns distinct values from the given slice of documents with the given key.
//
// If the key is not found in the document, the document is ignored.
//
// If the key is found in the document, and the value is an array, each element of the array is added to the result.
// Otherwise, the value itself is added to the result.
func FilterDistinctValues(docs []*types.Document, key string) (*types.Array, error) {
	distinct := types.MakeArray(len(docs))

	for _, doc := range docs {
		var val any

		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		val, err = doc.GetByPath(path)
		if err != nil {
			continue
		}

		switch v := val.(type) {
		case *types.Array:
			for i := 0; i < v.Len(); i++ {
				el, err := v.Get(i)
				if err != nil {
					return nil, lazyerrors.Error(err)
				}

				if !distinct.Contains(el) {
					distinct.Append(el)
				}
			}

		default:
			if !distinct.Contains(v) {
				distinct.Append(v)
			}
		}
	}

	SortArray(distinct, types.Ascending)

	return distinct, nil
}

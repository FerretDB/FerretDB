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
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/commonpath"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// DistinctParams contains `distinct` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type DistinctParams struct {
	DB         string          `ferretdb:"$db"`
	Collection string          `ferretdb:"collection"`
	Key        string          `ferretdb:"key"`
	Filter     *types.Document `ferretdb:"-"`
	Comment    string          `ferretdb:"comment,opt"`

	Query any `ferretdb:"query,opt"`

	Collation *types.Document `ferretdb:"collation,unimplemented"`

	ReadConcern *types.Document `ferretdb:"readConcern,ignored"`
	LSID        any             `ferretdb:"lsid,ignored"`
}

// GetDistinctParams returns `distinct` command parameters.
func GetDistinctParams(document *types.Document, l *zap.Logger) (*DistinctParams, error) {
	var dp DistinctParams

	err := commonparams.ExtractParams(document, "distinct", &dp, l)
	if err != nil {
		return nil, err
	}

	switch filter := dp.Query.(type) {
	case *types.Document:
		dp.Filter = filter
	case types.NullType, nil:
		dp.Filter = types.MakeDocument(0)
	default:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				"BSON field 'distinct.query' is the wrong type '%s', expected type 'object'",
				commonparams.AliasFromType(dp.Query),
			),
			"distinct",
		)
	}

	if dp.Key == "" {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrEmptyFieldPath,
			"FieldPath cannot be constructed with empty string",
		)
	}

	return &dp, nil
}

// FilterDistinctValues returns distinct values from the given slice of documents with the given key.
//
// If the key is not found in the document, the document is ignored.
//
// If the key is found in the document, and the value is an array, each element of the array is added to the result.
// Otherwise, the value itself is added to the result.
func FilterDistinctValues(iter types.DocumentsIterator, key string) (*types.Array, error) {
	distinct := types.MakeArray(0)

	defer iter.Close()

	for {
		_, doc, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		// vals contains all values exist at the suffix of the path
		vals, err := commonpath.FindValues(doc, path, &commonpath.FindValuesOpts{
			FindArrayIndex:    true,
			FindArrayElements: true,
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		for _, val := range vals {
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
	}

	SortArray(distinct, types.Ascending)

	return distinct, nil
}

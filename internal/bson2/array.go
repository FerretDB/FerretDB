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

package bson2

import (
	"errors"
	"log/slog"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements []any
}

// ConvertArray converts [*types.Array] to Array.
func ConvertArray(arr *types.Array) (*Array, error) {
	iter := arr.Iterator()
	defer iter.Close()

	elements := make([]any, arr.Len())

	for {
		i, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return &Array{
					elements: elements,
				}, nil
			}

			return nil, lazyerrors.Error(err)
		}

		v, err = convertFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		elements[i] = v
	}
}

// Convert converts Array to [*types.Array], decoding raw documents and arrays on the fly.
func (arr *Array) Convert() (*types.Array, error) {
	values := make([]any, len(arr.elements))

	for i, f := range arr.elements {
		v, err := convertToTypes(f)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		values[i] = v
	}

	res, err := types.NewArray(values...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// LogValue implements slog.LogValuer interface.
func (arr *Array) LogValue() slog.Value {
	return slogValue(arr)
}

// check interfaces
var (
	_ slog.LogValuer = (*Array)(nil)
)

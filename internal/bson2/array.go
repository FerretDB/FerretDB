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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// RawArray represents a BSON array in the binary encoded form.
//
// It generally references a part of a larger slice, not a copy.
type RawArray []byte

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements []any
}

// Convert converts the Array to the *types.Array, decoding raw documents and arrays on the fly.
func (arr *Array) Convert() (*types.Array, error) {
	values := make([]any, len(arr.elements))

	for i, f := range arr.elements {
		switch v := f.(type) {
		case *Document:
			d, err := v.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			values[i] = d

		case RawDocument:
			doc, err := DecodeDocument(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			d, err := doc.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			values[i] = d

		case *Array:
			a, err := v.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			values[i] = a

		case RawArray:
			arr, err := DecodeArray(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			a, err := arr.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			values[i] = a

		default:
			values[i] = convertScalar(v)
		}
	}

	res, err := types.NewArray(values...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

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
	"bytes"
	"encoding/binary"
	"errors"
	"log/slog"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements []any
}

// newArray creates a new Array from the given values.
func newArray(values ...any) (*Array, error) {
	res := &Array{
		elements: make([]any, 0, len(values)),
	}

	for i, v := range values {
		if err := res.add(v); err != nil {
			return nil, lazyerrors.Errorf("%d: %w", i, err)
		}
	}

	return res, nil
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

// add adds a new element to the Array.
func (arr *Array) add(value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.elements = append(arr.elements, value)

	return nil
}

// Encode encodes BSON array.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
// This method should accept a slice of bytes, not return it.
// That would allow to avoid unnecessary allocations.
func (arr *Array) Encode() (RawArray, error) {
	size := sizeAny(arr)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for i, v := range arr.elements {
		if err := encodeField(buf, strconv.Itoa(i), v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

// LogValue implements slog.LogValuer interface.
func (arr *Array) LogValue() slog.Value {
	return slogValue(arr)
}

func (arr *Array) LogMessage() string {
	return slogMessage(arr)
}

// check interfaces
var (
	_ slog.LogValuer = (*Array)(nil)
)

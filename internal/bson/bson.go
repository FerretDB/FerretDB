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

// Package bson provides convertors between wirebson and types packages.
package bson

import (
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// convertFromTypes converts types package value to wirebson package value.
//
// Invalid types cause panics.
func convertFromTypes(v any) (any, error) {
	switch v := v.(type) {
	case *types.Document:
		doc, err := FromDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *types.Array:
		arr, err := FromArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case types.Binary:
		return wirebson.Binary{
			B:       v.B,
			Subtype: wirebson.BinarySubtype(v.Subtype),
		}, nil
	case types.ObjectID:
		return wirebson.ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case types.NullType:
		return wirebson.Null, nil
	case types.Regex:
		return wirebson.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case types.Timestamp:
		return wirebson.Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid type %T", v))
	}
}

// From converts types package value to wirebson package value.
func From[T types.Type](v T) (any, error) {
	return convertFromTypes(v)
}

// FromArray converts [*types.Array] to [*wirebson.Array].
func FromArray(arr *types.Array) (*wirebson.Array, error) {
	iter := arr.Iterator()
	defer iter.Close()

	elements := wirebson.MakeArray(arr.Len())

	for {
		_, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return elements, nil
			}

			return nil, lazyerrors.Error(err)
		}

		v, err = convertFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = elements.Add(v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}
}

// FromDocument converts [*types.Document] to [*wirebson.Document].
func FromDocument(doc *types.Document) (*wirebson.Document, error) {
	iter := doc.Iterator()
	defer iter.Close()

	res := wirebson.MakeDocument(doc.Len())

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return res, nil
			}

			return nil, lazyerrors.Error(err)
		}

		v, err = convertFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = res.Add(k, v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}
}

// convertToTypes converts wirebson package value to types package value.
//
// Invalid types cause panics.
func convertToTypes(v any) (any, error) {
	switch v := v.(type) {
	case *wirebson.Document:
		doc, err := ToDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case wirebson.RawDocument:
		doc, err := ToDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *wirebson.Array:
		arr, err := ToArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case wirebson.RawArray:
		arr, err := ToArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case wirebson.Binary:
		// Special case to prevent it from being stored as null in sjson.
		// TODO https://github.com/FerretDB/FerretDB/issues/260
		if v.B == nil {
			v.B = []byte{}
		}

		return types.Binary{
			B:       v.B,
			Subtype: types.BinarySubtype(v.Subtype),
		}, nil
	case wirebson.ObjectID:
		return types.ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case wirebson.NullType:
		return types.Null, nil
	case wirebson.Regex:
		return types.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case wirebson.Timestamp:
		return types.Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

// ToArray converts wirebson array to [*types.Array].
func ToArray(a wirebson.AnyArray) (*types.Array, error) {
	arr, err := a.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	values := make([]any, arr.Len())

	for i := range arr.Len() {
		var v any

		if v, err = convertToTypes(arr.Get(i)); err != nil {
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

// ToDocument converts wirebson document to [*types.Document].
func ToDocument(d wirebson.AnyDocument) (*types.Document, error) {
	doc, err := d.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	fields := doc.FieldNames()
	pairs := make([]any, 0, len(fields)*2)

	for i := range fields {
		f, v := doc.GetByIndex(i)

		if v, err = convertToTypes(v); err != nil {
			return nil, lazyerrors.Error(err)
		}

		pairs = append(pairs, f, v)
	}

	res, err := types.NewDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

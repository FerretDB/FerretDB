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

// Convert converts valid types package values to BSON values of that package.
//
// Conversions of composite types may cause errors.
func Convert[T types.Type](v T) (any, error) {
	return convertFromTypes(v)
}

// convertFromTypes is a variant of [Convert] without type parameters (generics).
//
// Invalid types cause panics.
func convertFromTypes(v any) (any, error) {
	switch v := v.(type) {
	case *types.Document:
		doc, err := ConvertDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *types.Array:
		arr, err := ConvertArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case types.Binary:
		return Binary{
			B:       v.B,
			Subtype: BinarySubtype(v.Subtype),
		}, nil
	case types.ObjectID:
		return ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case types.NullType:
		return Null, nil
	case types.Regex:
		return Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case types.Timestamp:
		return Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid type %T", v))
	}
}

// convertToTypes converts valid BSON value of that package to types package type.
//
// Conversions of composite types (including raw types) may cause errors.
// Invalid types cause panics.
func convertToTypes(v any) (any, error) {
	switch v := v.(type) {
	case *Document:
		doc, err := v.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case RawDocument:
		d, err := v.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		doc, err := d.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *Array:
		arr, err := v.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case RawArray:
		a, err := v.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		arr, err := a.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case Binary:
		// Special case to prevent it from being stored as null in sjson.
		// TODO https://github.com/FerretDB/FerretDB/issues/260
		if v.B == nil {
			v.B = []byte{}
		}

		return types.Binary{
			B:       v.B,
			Subtype: types.BinarySubtype(v.Subtype),
		}, nil
	case ObjectID:
		return types.ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case NullType:
		return types.Null, nil
	case Regex:
		return types.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case Timestamp:
		return types.Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

// convertBSONFromTypes converts types package value to BSON value of wirebson package.
//
// Invalid types cause panics.
func convertBSONFromTypes(v any) (any, error) {
	switch v := v.(type) {
	case *types.Document:
		doc, err := ConvertBSONDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *types.Array:
		arr, err := ConvertBSONArray(v)
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

// convertBSONToTypes converts valid BSON value of wirebson package to types package type.
//
// Invalid types cause panics.
func convertBSONToTypes(v any) (any, error) {
	switch v := v.(type) {
	case *wirebson.Document:
		doc, err := TypesDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case wirebson.RawDocument:
		doc, err := TypesDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *wirebson.Array:
		arr, err := TypesArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case wirebson.RawArray:
		arr, err := TypesArray(v)
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

// ConvertDocument converts [*types.Document] to Document.
func ConvertDocument(doc *types.Document) (*Document, error) {
	iter := doc.Iterator()
	defer iter.Close()

	res := MakeDocument(doc.Len())

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

// Convert converts Document to [*types.Document], decoding raw documents and arrays on the fly.
func (doc *Document) Convert() (*types.Document, error) {
	pairs := make([]any, 0, len(doc.fields)*2)

	for _, f := range doc.fields {
		v, err := convertToTypes(f.value)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		pairs = append(pairs, f.name, v)
	}

	res, err := types.NewDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// Convert converts a single valid BSON array that takes the whole byte slice into [*types.Array].
func (raw RawArray) Convert() (*types.Array, error) {
	arr, err := raw.decode(decodeShallow)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := arr.Convert()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// Convert converts a single valid BSON document that takes the whole byte slice into [*types.Document].
func (raw RawDocument) Convert() (*types.Document, error) {
	doc, err := raw.decode(decodeShallow)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := doc.Convert()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// ConvertBSONArray converts [*types.Array] to [*wirebson.Array].
func ConvertBSONArray(arr *types.Array) (*wirebson.Array, error) {
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

		v, err = convertBSONFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = elements.Add(v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}
}

// TypesArray decodes an array and converts to [*types.Array].
func TypesArray(a wirebson.AnyArray) (*types.Array, error) {
	arr, err := a.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	values := make([]any, arr.Len())

	for i := range arr.Len() {
		var v any

		if v, err = convertBSONToTypes(arr.Get(i)); err != nil {
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

// ConvertBSONDocument converts [*types.Document] to [*wirebson.Document].
func ConvertBSONDocument(doc *types.Document) (*wirebson.Document, error) {
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

		v, err = convertBSONFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = res.Add(k, v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}
}

// TypesDocument decodes a document and converts to [*types.Document].
func TypesDocument(d wirebson.AnyDocument) (*types.Document, error) {
	doc, err := d.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	fields := doc.FieldNames()
	pairs := make([]any, 0, len(fields)*2)

	for i := range fields {
		f, v := doc.GetByIndex(i)

		if v, err = convertBSONToTypes(v); err != nil {
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

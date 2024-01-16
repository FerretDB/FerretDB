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
	"fmt"
	"log/slog"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// RawDocument represents a BSON document a.k.a object in the binary encoded form.
//
// It generally references a part of a larger slice, not a copy.
type RawDocument []byte

// LogValue implements slog.LogValuer interface.
func (doc *RawDocument) LogValue() slog.Value {
	return slogValue(doc)
}

// field represents a single Document field in the (partially) decoded form.
type field struct {
	value any
	name  string
}

// Document represents a BSON document a.k.a object in the (partially) decoded form.
//
// It may contain duplicate field names.
type Document struct {
	fields []field
}

// NewDocument creates a new Document from the given pairs of field names and values.
func NewDocument(pairs ...any) (*Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return nil, fmt.Errorf("invalid number of arguments: %d", l)
	}

	res := &Document{
		fields: make([]field, l/2),
	}

	for i := 0; i < l; i += 2 {
		name, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("invalid field name type: %T", pairs[i])
		}

		value := pairs[i+1]
		if !validBSONType(value) {
			return nil, fmt.Errorf("invalid field value type: %T", value)
		}

		res.fields[i/2] = field{
			name:  name,
			value: value,
		}
	}

	return res, nil
}

// MakeDocument creates a new empty Document with the given capacity.
func MakeDocument(cap int) *Document {
	return &Document{
		fields: make([]field, 0, cap),
	}
}

// ConvertDocument converts [*types.Document] to Document.
func ConvertDocument(doc *types.Document) (*Document, error) {
	iter := doc.Iterator()
	defer iter.Close()

	pairs := make([]any, 0, doc.Len()*2)

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				var res *Document

				if res, err = NewDocument(pairs...); err != nil {
					return nil, lazyerrors.Error(err)
				}

				return res, nil
			}

			return nil, lazyerrors.Error(err)
		}

		v, err = convertFromTypes(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		pairs = append(pairs, k, v)
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

// add adds a new field to the Document.
func (doc *Document) add(name string, value any) error {
	if !validBSONType(value) {
		return fmt.Errorf("invalid field value type: %T", value)
	}

	doc.fields = append(doc.fields, field{
		name:  name,
		value: value,
	})

	return nil
}

// LogValue implements slog.LogValuer interface.
func (doc *Document) LogValue() slog.Value {
	return slogValue(doc)
}

// check interfaces
var (
	_ slog.LogValuer = (*Document)(nil)
)

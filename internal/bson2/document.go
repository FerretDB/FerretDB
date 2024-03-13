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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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
		return nil, lazyerrors.Errorf("invalid number of arguments: %d", l)
	}

	res := MakeDocument(l / 2)

	for i := 0; i < l; i += 2 {
		name, ok := pairs[i].(string)
		if !ok {
			return nil, lazyerrors.Errorf("invalid field name type: %T", pairs[i])
		}

		value := pairs[i+1]

		if err := res.Add(name, value); err != nil {
			return nil, lazyerrors.Error(err)
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

// Get returns a value of the field with the given name.
//
// It returns nil if the field is not found.
// If document contains duplicate field names, it returns the first one.
func (doc *Document) Get(name string) any {
	for _, f := range doc.fields {
		if f.name == name {
			return f.value
		}
	}

	return nil
}

// Add adds a new field to the Document.
func (doc *Document) Add(name string, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Errorf("%q: %w", name, err)
	}

	doc.fields = append(doc.fields, field{
		name:  name,
		value: value,
	})

	return nil
}

// Encode encodes BSON document.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
// This method should accept a slice of bytes, not return it.
// That would allow to avoid unnecessary allocations.
func (doc *Document) Encode() (RawDocument, error) {
	size := sizeAny(doc)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, f := range doc.fields {
		if err := encodeField(buf, f.name, f.value); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

// LogValue implements slog.LogValuer interface.
func (doc *Document) LogValue() slog.Value {
	return slogValue(doc, 1)
}

// LogMessage returns an indented representation as a string,
// somewhat similar (but not identical) to JSON or Go syntax.
// It may change over time.
func (doc *Document) LogMessage() string {
	return logMessage(doc, logMaxFlowLength, "", 1)
}

// LogMessageBlock is a variant of [Document.LogMessage] that never uses a flow style.
func (doc *Document) LogMessageBlock() string {
	return logMessage(doc, 0, "", 1)
}

// check interfaces
var (
	_ slog.LogValuer = (*Document)(nil)
)

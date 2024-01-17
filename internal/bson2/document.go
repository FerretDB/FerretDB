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
	"fmt"
	"log/slog"
	"time"

	"github.com/cristalhq/bson/bsonproto"

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

		if err := res.add(name, value); err != nil {
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

		if err = res.add(k, v); err != nil {
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

// add adds a new field to the Document.
func (doc *Document) add(name string, value any) error {
	if !validBSONType(value) {
		return lazyerrors.Errorf("invalid field value type: %T", value)
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

// encodeDocument encodes BSON document.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
func encodeDocument(doc *Document) ([]byte, error) {
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

// encodeField encodes document field.
//
// It panics if v is not a valid type.
func encodeField(buf *bytes.Buffer, name string, v any) error {
	switch v := v.(type) {
	case *Document:
		if err := buf.WriteByte(byte(tagDocument)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, bsonproto.SizeCString(name))
		bsonproto.EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		b, err := encodeDocument(v)
		if err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

	case RawDocument:
		if err := buf.WriteByte(byte(tagDocument)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, bsonproto.SizeCString(name))
		bsonproto.EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(v); err != nil {
			return lazyerrors.Error(err)
		}

	case *Array:
		if err := buf.WriteByte(byte(tagArray)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, bsonproto.SizeCString(name))
		bsonproto.EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		b, err := encodeArray(v)
		if err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

	case RawArray:
		if err := buf.WriteByte(byte(tagArray)); err != nil {
			return lazyerrors.Error(err)
		}

		b := make([]byte, bsonproto.SizeCString(name))
		bsonproto.EncodeCString(b, name)

		if _, err := buf.Write(b); err != nil {
			return lazyerrors.Error(err)
		}

		if _, err := buf.Write(v); err != nil {
			return lazyerrors.Error(err)
		}

	default:
		return encodeScalarField(buf, name, v)
	}

	return nil
}

// encodeScalarField encodes scalar document field.
//
// It panics if v is not a scalar value.
func encodeScalarField(buf *bytes.Buffer, name string, v any) error {
	switch v.(type) {
	case float64:
		buf.WriteByte(byte(tagFloat64))
	case string:
		buf.WriteByte(byte(tagString))
	case Binary:
		buf.WriteByte(byte(tagBinary))
	case ObjectID:
		buf.WriteByte(byte(tagObjectID))
	case bool:
		buf.WriteByte(byte(tagBool))
	case time.Time:
		buf.WriteByte(byte(tagTime))
	case NullType:
		buf.WriteByte(byte(tagNull))
	case Regex:
		buf.WriteByte(byte(tagRegex))
	case int32:
		buf.WriteByte(byte(tagInt32))
	case Timestamp:
		buf.WriteByte(byte(tagTimestamp))
	case int64:
		buf.WriteByte(byte(tagInt64))
	default:
		panic(fmt.Sprintf("invalid type %T", v))
	}

	b := make([]byte, bsonproto.SizeCString(name))
	bsonproto.EncodeCString(b, name)

	if _, err := buf.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

	b = make([]byte, bsonproto.SizeAny(v))
	bsonproto.EncodeAny(b, v)

	if _, err := buf.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// check interfaces
var (
	_ slog.LogValuer = (*Document)(nil)
)

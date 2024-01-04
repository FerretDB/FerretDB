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
	"fmt"
	"strconv"
	"time"

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// encodeDocument encodes BSON document.
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

// encodeArray encodes BSON array.
func encodeArray(arr *Array) ([]byte, error) {
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

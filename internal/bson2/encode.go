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
	"strconv"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/cristalhq/bson/bsonproto"
)

func encodeDocument(doc *Document) ([]byte, error) {
	size := sizeAny(doc)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, f := range doc.fields {
		switch v := f.value.(type) {
		case *Document:
			if err := buf.WriteByte(byte(tagDocument)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(f.name))
			bsonproto.EncodeCString(b, f.name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b, err := encodeDocument(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case RawDocument:
			if err := buf.WriteByte(byte(tagDocument)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(f.name))
			bsonproto.EncodeCString(b, f.name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if _, err := buf.Write(v); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case *Array:
			if err := buf.WriteByte(byte(tagArray)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(f.name))
			bsonproto.EncodeCString(b, f.name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b, err := encodeArray(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case RawArray:
			if err := buf.WriteByte(byte(tagArray)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(f.name))
			bsonproto.EncodeCString(b, f.name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if _, err := buf.Write(v); err != nil {
				return nil, lazyerrors.Error(err)
			}

		default:
			encodeScalarField(buf, v)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

func encodeArray(arr *Array) ([]byte, error) {
	size := sizeAny(arr)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for i, v := range arr.elements {
		name := strconv.Itoa(i)

		switch v := v.(type) {
		case *Document:
			if err := buf.WriteByte(byte(tagDocument)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(name))
			bsonproto.EncodeCString(b, name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b, err := encodeDocument(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case RawDocument:
			if err := buf.WriteByte(byte(tagDocument)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(name))
			bsonproto.EncodeCString(b, name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if _, err := buf.Write(v); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case *Array:
			if err := buf.WriteByte(byte(tagArray)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(name))
			bsonproto.EncodeCString(b, name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b, err := encodeArray(v)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case RawArray:
			if err := buf.WriteByte(byte(tagArray)); err != nil {
				return nil, lazyerrors.Error(err)
			}

			b := make([]byte, bsonproto.SizeCString(name))
			bsonproto.EncodeCString(b, name)
			if _, err := buf.Write(b); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if _, err := buf.Write(v); err != nil {
				return nil, lazyerrors.Error(err)
			}

		default:
			encodeScalarField(buf, v)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

func encodeScalarField(buf *bytes.Buffer, v any) {
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
		panic("TODO")
	}

	b := make([]byte, bsonproto.SizeAny(v))
	bsonproto.EncodeAny(b, v)
	must.NotFail(buf.Write(b))
}

func sizeAny(v any) int {
	switch v := v.(type) {
	case *Document:
		return sizeDocument(v)
	case RawDocument:
		return len(v)
	case *Array:
		return sizeArray(v)
	case RawArray:
		return len(v)
	default:
		return bsonproto.SizeAny(v)
	}
}

func sizeDocument(doc *Document) int {
	size := 5

	for _, f := range doc.fields {
		size += 1 + len(f.name) + 1 + sizeAny(f.value)
	}

	return size
}

func sizeArray(arr *Array) int {
	size := 5

	for i, v := range arr.elements {
		size += 1 + len(strconv.Itoa(i)) + 1 + sizeAny(v)
	}

	return size
}

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
	"encoding/binary"
	"log/slog"
	"math"

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// RawDocument represents a singe BSON document a.k.a object in the binary encoded form.
//
// It generally references a part of a larger slice, not a copy.
type RawDocument []byte

// FindRawDocument returns the first BSON document in the byte slice.
// It should start at the first byte.
//
// Returned document might not be valid. It is the caller's responsibility to check it.
//
// Use RawDocument(b) conversion instead of b contains exactly one document and no extra bytes.
func FindRawDocument(b []byte) RawDocument {
	bl := len(b)
	if bl < 5 {
		return nil
	}

	dl := int(binary.LittleEndian.Uint32(b))
	if bl < dl {
		return nil
	}

	if b[dl-1] != 0 {
		return nil
	}

	return b[:dl]
}

// LogValue implements slog.LogValuer interface.
func (doc RawDocument) LogValue() slog.Value {
	return slogValue(doc)
}

// Decode decodes a single BSON document that takes the whole byte slice.
//
// Only top-level fields are decoded;
// nested documents and arrays are converted to RawDocument and RawArray respectively,
// using raw's subslices without copying.
func (raw RawDocument) Decode() (*Document, error) {
	res, err := raw.decode(decodeShallow)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// DecodeDeep decodes a single valid BSON document that takes the whole byte slice.
//
// All nested documents and arrays are decoded recursively.
func (raw RawDocument) DecodeDeep() (*Document, error) {
	res, err := raw.decode(decodeDeep)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// Check recursively checks that the whole byte slice contains a single valid BSON document.
func (raw RawDocument) Check() error {
	_, err := raw.decode(decodeCheckOnly)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
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

// decode decodes a single BSON document that takes the whole byte slice.
func (raw RawDocument) decode(mode decodeMode) (*Document, error) {
	bl := len(raw)
	if bl < 5 {
		return nil, lazyerrors.Errorf("len(b) = %d: %w", bl, ErrDecodeShortInput)
	}

	if dl := int(binary.LittleEndian.Uint32(raw)); bl != dl {
		return nil, lazyerrors.Errorf("len(b) = %d, document length = %d: %w", bl, dl, ErrDecodeInvalidInput)
	}

	if last := raw[bl-1]; last != 0 {
		return nil, lazyerrors.Errorf("last = %d: %w", last, ErrDecodeInvalidInput)
	}

	var res *Document
	if mode != decodeCheckOnly {
		res = MakeDocument(1)
	}

	offset := 4
	for offset != len(raw)-1 {
		if err := decodeCheckOffset(raw, offset, 1); err != nil {
			return nil, lazyerrors.Error(err)
		}

		t := tag(raw[offset])
		offset++

		if err := decodeCheckOffset(raw, offset, 1); err != nil {
			return nil, lazyerrors.Error(err)
		}

		name, err := bsonproto.DecodeCString(raw[offset:])
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		offset += len(name) + 1

		var v any

		switch t {
		case tagFloat64:
			var f float64
			f, err = bsonproto.DecodeFloat64(raw[offset:])
			offset += bsonproto.SizeFloat64
			v = f

			if noNaN && math.IsNaN(f) {
				return nil, lazyerrors.Errorf("got NaN value: %w", ErrDecodeInvalidInput)
			}

		case tagString:
			var s string
			s, err = bsonproto.DecodeString(raw[offset:])
			offset += bsonproto.SizeString(s)
			v = s

		case tagDocument:
			if err = decodeCheckOffset(raw, offset, 4); err != nil {
				return nil, lazyerrors.Error(err)
			}

			l := int(binary.LittleEndian.Uint32(raw[offset:]))

			if err = decodeCheckOffset(raw, offset, l); err != nil {
				return nil, lazyerrors.Error(err)
			}

			doc := RawDocument(raw[offset : offset+l])
			offset += l

			switch mode {
			case decodeShallow:
				v = doc
			case decodeDeep:
				v, err = doc.decode(decodeDeep)
			case decodeCheckOnly:
				_, err = doc.decode(decodeCheckOnly)
			}

		case tagArray:
			if err = decodeCheckOffset(raw, offset, 4); err != nil {
				return nil, lazyerrors.Error(err)
			}

			l := int(binary.LittleEndian.Uint32(raw[offset:]))

			if err = decodeCheckOffset(raw, offset, l); err != nil {
				return nil, lazyerrors.Error(err)
			}

			raw := RawArray(raw[offset : offset+l])
			offset += l

			switch mode {
			case decodeShallow:
				v = raw
			case decodeDeep:
				v, err = raw.decode(decodeDeep)
			case decodeCheckOnly:
				_, err = raw.decode(decodeCheckOnly)
			}

		case tagBinary:
			var s Binary
			s, err = bsonproto.DecodeBinary(raw[offset:])
			offset += bsonproto.SizeBinary(s)
			v = s

		case tagObjectID:
			v, err = bsonproto.DecodeObjectID(raw[offset:])
			offset += bsonproto.SizeObjectID

		case tagBool:
			v, err = bsonproto.DecodeBool(raw[offset:])
			offset += bsonproto.SizeBool

		case tagTime:
			v, err = bsonproto.DecodeTime(raw[offset:])
			offset += bsonproto.SizeTime

		case tagNull:
			v = Null

		case tagRegex:
			var s Regex
			s, err = bsonproto.DecodeRegex(raw[offset:])
			offset += bsonproto.SizeRegex(s)
			v = s

		case tagInt32:
			v, err = bsonproto.DecodeInt32(raw[offset:])
			offset += bsonproto.SizeInt32

		case tagTimestamp:
			v, err = bsonproto.DecodeTimestamp(raw[offset:])
			offset += bsonproto.SizeTimestamp

		case tagInt64:
			v, err = bsonproto.DecodeInt64(raw[offset:])
			offset += bsonproto.SizeInt64

		case tagUndefined, tagDBPointer, tagJavaScript, tagSymbol, tagJavaScriptScope, tagDecimal, tagMinKey, tagMaxKey:
			return nil, lazyerrors.Errorf("unsupported tag: %s", t)

		default:
			return nil, lazyerrors.Errorf("unexpected tag: %s", t)
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if mode != decodeCheckOnly {
			must.NoError(res.add(name, v))
		}
	}

	return res, nil
}

// decodeCheckOffset checks that b has enough bytes to decode size bytes starting from offset.
func decodeCheckOffset(b []byte, offset, size int) error {
	if len(b[offset:]) < size+1 {
		return lazyerrors.Errorf("offset = %d, size = %d: %w", offset, size, ErrDecodeShortInput)
	}

	return nil
}

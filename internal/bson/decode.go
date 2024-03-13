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
	"encoding/binary"

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FindRaw finds the first raw BSON document or array in b and returns its length l.
// It should start from the first byte of b.
// RawDocument(b[:l] / RawArray(b[:l]) might not be valid. It is the caller's responsibility to check it.
//
// Use RawDocument(b) / RawArray(b) conversion instead if b contains exactly one document/array and no extra bytes.
func FindRaw(b []byte) (int, error) {
	bl := len(b)
	if bl < 5 {
		return 0, lazyerrors.Errorf("len(b) = %d: %w", bl, ErrDecodeShortInput)
	}

	dl := int(binary.LittleEndian.Uint32(b))
	if dl < 5 {
		return 0, lazyerrors.Errorf("dl = %d: %w", dl, ErrDecodeInvalidInput)
	}

	if bl < dl {
		return 0, lazyerrors.Errorf("len(b) = %d, dl = %d: %w", bl, dl, ErrDecodeShortInput)
	}

	if b[dl-1] != 0 {
		return 0, lazyerrors.Errorf("invalid last byte: %w", ErrDecodeInvalidInput)
	}

	return dl, nil
}

// decodeCheckOffset checks that b has enough bytes to decode size bytes starting from offset.
func decodeCheckOffset(b []byte, offset, size int) error {
	if l := len(b); l < offset+size {
		return lazyerrors.Errorf("len(b) = %d, offset = %d, size = %d: %w", l, offset, size, ErrDecodeShortInput)
	}

	return nil
}

func decodeScalarField(b []byte, t tag) (v any, size int, err error) {
	switch t {
	case tagFloat64:
		var f float64
		f, err = bsonproto.DecodeFloat64(b)
		v = f
		size = bsonproto.SizeFloat64

	case tagString:
		var s string
		s, err = bsonproto.DecodeString(b)
		v = s
		size = bsonproto.SizeString(s)

	case tagBinary:
		var bin Binary
		bin, err = bsonproto.DecodeBinary(b)
		v = bin
		size = bsonproto.SizeBinary(bin)

	case tagObjectID:
		v, err = bsonproto.DecodeObjectID(b)
		size = bsonproto.SizeObjectID

	case tagBool:
		v, err = bsonproto.DecodeBool(b)
		size = bsonproto.SizeBool

	case tagTime:
		v, err = bsonproto.DecodeTime(b)
		size = bsonproto.SizeTime

	case tagNull:
		v = Null

	case tagRegex:
		var re Regex
		re, err = bsonproto.DecodeRegex(b)
		v = re
		size = bsonproto.SizeRegex(re)

	case tagInt32:
		v, err = bsonproto.DecodeInt32(b)
		size = bsonproto.SizeInt32

	case tagTimestamp:
		v, err = bsonproto.DecodeTimestamp(b)
		size = bsonproto.SizeTimestamp

	case tagInt64:
		v, err = bsonproto.DecodeInt64(b)
		size = bsonproto.SizeInt64

	case tagUndefined, tagDBPointer, tagJavaScript, tagSymbol, tagJavaScriptScope, tagDecimal, tagMinKey, tagMaxKey:
		err = lazyerrors.Errorf("unsupported tag %s: %w", t, ErrDecodeInvalidInput)

	case tagDocument, tagArray:
		err = lazyerrors.Errorf("non-scalar tag: %s", t)

	default:
		err = lazyerrors.Errorf("unexpected tag %s: %w", t, ErrDecodeInvalidInput)
	}

	return
}

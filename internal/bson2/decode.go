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

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func decodeCString(b []byte) (string, error) {
	var i int
	var v byte
	for i, v = range b {
		if v == 0 {
			break
		}
	}

	if v != 0 {
		return "", lazyerrors.Error(ErrDecodeInvalidInput)
	}

	return string(b[:i]), nil
}

func DecodeDocument(b RawDocument) (*Document, error) {
	l := binary.LittleEndian.Uint32(b)
	if len(b) != int(l) {
		return nil, lazyerrors.Error(ErrDecodeInvalidInput)
	}
	if b[len(b)-1] != 0 {
		return nil, lazyerrors.Error(ErrDecodeInvalidInput)
	}

	res := MakeDocument(1)

	offset := 4
	for offset != len(b)-1 {
		t := tag(b[offset])
		offset++

		name, err := decodeCString(b[offset:])
		offset += len(name) + 1
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var v any
		switch t {
		case tagFloat64:
			v, err = bsonproto.DecodeFloat64(b[offset:])
			offset += bsonproto.SizeFloat64

		case tagString:
			var s string
			s, err = bsonproto.DecodeString(b[offset:])
			offset += bsonproto.SizeString(s)
			v = s

		case tagDocument:
			panic("TagDocument")

		case tagArray:
			panic("TagArray")

		case tagBinary:
			var s bsonproto.Binary
			s, err = bsonproto.DecodeBinary(b[offset:])
			offset += bsonproto.SizeBinary(s)
			v = s

		case tagObjectID:
			v, err = bsonproto.DecodeObjectID(b[offset:])
			offset += bsonproto.SizeObjectID

		case tagBool:
			v, err = bsonproto.DecodeBool(b[offset:])
			offset += bsonproto.SizeBool

		case tagTime:
			v, err = bsonproto.DecodeTime(b[offset:])
			offset += bsonproto.SizeTime

		case tagNull:
			v = bsonproto.Null

		case tagRegex:
			var s bsonproto.Regex
			s, err = bsonproto.DecodeRegex(b[offset:])
			offset += bsonproto.SizeRegex(s)
			v = s

		case tagInt32:
			v, err = bsonproto.DecodeInt32(b[offset:])
			offset += bsonproto.SizeInt32

		case tagTimestamp:
			v, err = bsonproto.DecodeTimestamp(b[offset:])
			offset += bsonproto.SizeTimestamp

		case tagInt64:
			v, err = bsonproto.DecodeInt64(b[offset:])
			offset += bsonproto.SizeInt64

		default:
			return nil, lazyerrors.Errorf("unsupported tag: %d", t)
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		must.NoError(res.add(name, v))
	}

	return res, nil
}

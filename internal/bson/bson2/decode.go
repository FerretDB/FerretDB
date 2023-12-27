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

	i := 4
	for i != len(b)-1 {
		tag := b[i]
		i++

		name, err := decodeCString(b[i:])
		i += len(name) + 1
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var v any
		switch tag {
		case TagFloat64:
			v, err = bsonproto.DecodeFloat64(b[i:])
			i += bsonproto.SizeFloat64

		default:
			return nil, lazyerrors.Errorf("unsupported tag: %d", tag)
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		must.NoError(res.Add(name, v))
	}

	return res, nil
}

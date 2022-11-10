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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	minDocumentLen = 5
)

// Document represents BSON Document type.
type Document types.Document

func (doc *Document) bsontype() {}

// ReadFrom implements bsontype interface.
func (doc *Document) ReadFrom(r *bufio.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return lazyerrors.Errorf("bson.Document.ReadFrom (binary.Read): %w", err)
	}
	if l < minDocumentLen || l > types.MaxDocumentLen {
		return lazyerrors.Errorf("bson.Document.ReadFrom: invalid length %d", l)
	}

	// make buffer
	b := make([]byte, l)

	binary.LittleEndian.PutUint32(b, uint32(l))

	// read e_list and terminating zero
	n, err := io.ReadFull(r, b[4:])
	if err != nil {
		return lazyerrors.Errorf("bson.Document.ReadFrom (io.ReadFull, expected %d, read %d): %w", len(b), n, err)
	}

	bufr := bufio.NewReader(bytes.NewReader(b[4:]))

	td := must.NotFail(types.NewDocument())
	for {
		t, err := bufr.ReadByte()
		if err != nil {
			return lazyerrors.Errorf("bson.Document.ReadFrom (ReadByte): %w", err)
		}

		if t == 0 {
			// documented ended
			if _, err := bufr.Peek(1); err != io.EOF {
				return lazyerrors.Errorf("unexpected end of the document: %w", err)
			}
			break
		}

		var ename CString
		if err := ename.ReadFrom(bufr); err != nil {
			return lazyerrors.Errorf("bson.Document.ReadFrom (ename.ReadFrom): %w", err)
		}

		key := string(ename)

		switch tag(t) {
		case tagDocument:
			// TODO check maximum nesting

			var v Document
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (embedded document): %w", err)
			}
			d := types.Document(v)
			td.Add(key, &d)

		case tagArray:
			// TODO check maximum nesting

			var v arrayType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Array): %w", err)
			}
			a := types.Array(v)
			td.Add(key, &a)

		case tagDouble:
			var v doubleType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Double): %w", err)
			}

			td.Add(key, float64(v))

		case tagString:
			var v stringType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (String): %w", err)
			}

			td.Add(key, string(v))

		case tagBinary:
			var v binaryType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Binary): %w", err)
			}

			td.Add(key, types.Binary(v))

		case tagUndefined:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type `Undefined (value) — Deprecated`")

		case tagObjectID:
			var v objectIDType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (ObjectID): %w", err)
			}

			td.Add(key, types.ObjectID(v))

		case tagBool:
			var v boolType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Bool): %w", err)
			}

			td.Add(key, bool(v))

		case tagDateTime:
			var v dateTimeType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (DateTime): %w", err)
			}

			td.Add(key, time.Time(v))

		case tagNull:
			// skip calling ReadFrom that does nothing
			td.Add(key, types.Null)

		case tagRegex:
			var v regexType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Regex): %w", err)
			}

			td.Add(key, types.Regex(v))

		case tagInt32:
			var v int32Type
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Int32): %w", err)
			}

			td.Add(key, int32(v))

		case tagTimestamp:
			var v timestampType
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Timestamp): %w", err)
			}

			td.Add(key, types.Timestamp(v))

		case tagInt64:
			var v int64Type
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Int64): %w", err)
			}

			td.Add(key, int64(v))

		case tagDBPointer, tagDecimal, tagJavaScript, tagJavaScriptScope, tagMaxKey, tagMinKey, tagSymbol:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type %#02x (%s)", t, tag(t))
		default:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type %#02x (%s)", t, tag(t))
		}
	}

	*doc = Document(*td)
	return nil
}

// WriteTo implements bsontype interface.
func (doc Document) WriteTo(w *bufio.Writer) error {
	v, err := doc.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Document.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Document.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (doc Document) MarshalBinary() ([]byte, error) {
	td := types.Document(doc)

	var elist bytes.Buffer
	bufw := bufio.NewWriter(&elist)

	for _, elK := range td.Keys() {
		ename := CString(elK)
		elV := must.NotFail(td.Get(elK))

		switch elV := elV.(type) {
		case *types.Document:
			bufw.WriteByte(byte(tagDocument))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

			doc := Document(*elV)
			if err := doc.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case *types.Array:
			bufw.WriteByte(byte(tagArray))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := arrayType(*elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case float64:
			bufw.WriteByte(byte(tagDouble))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := doubleType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case string:
			bufw.WriteByte(byte(tagString))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := stringType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Binary:
			bufw.WriteByte(byte(tagBinary))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := binaryType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.ObjectID:
			bufw.WriteByte(byte(tagObjectID))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := objectIDType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case bool:
			bufw.WriteByte(byte(tagBool))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := boolType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case time.Time:
			bufw.WriteByte(byte(tagDateTime))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := dateTimeType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.NullType:
			bufw.WriteByte(byte(tagNull))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			// skip calling WriteTo that does nothing

		case types.Regex:
			bufw.WriteByte(byte(tagRegex))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := regexType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case int32:
			bufw.WriteByte(byte(tagInt32))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := int32Type(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Timestamp:
			bufw.WriteByte(byte(tagTimestamp))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := timestampType(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case int64:
			bufw.WriteByte(byte(tagInt64))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := int64Type(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		default:
			return nil, lazyerrors.Errorf("bson.Document.MarshalBinary: unhandled element type %T", elV)
		}
	}

	if err := bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res bytes.Buffer
	l := int32(elist.Len() + 5)
	binary.Write(&res, binary.LittleEndian, l)
	if _, err := elist.WriteTo(&res); err != nil {
		panic(err)
	}
	res.WriteByte(0)
	if int32(res.Len()) != l {
		panic(fmt.Sprintf("got %d, expected %d", res.Len(), l))
	}
	return res.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*Document)(nil)
)

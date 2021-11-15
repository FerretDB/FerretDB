// Copyright 2021 Baltoro OÜ.
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
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

const (
	MaxDocumentLen = 16777216

	minDocumentLen = 5
)

// Common interface with types.Document.
type document interface {
	Map() map[string]interface{}
	Keys() []string
}

type Document struct {
	m    map[string]interface{}
	keys []string
}

// ConvertDocument converts types.Document to bson.Document and validates it.
// It references the same data without copying it.
func ConvertDocument(d document) (*Document, error) {
	doc := &Document{
		m:    d.Map(),
		keys: d.Keys(),
	}

	if doc.m == nil {
		doc.m = map[string]interface{}{}
	}
	if doc.keys == nil {
		doc.keys = []string{}
	}

	// for validation
	if _, err := types.ConvertDocument(doc); err != nil {
		return nil, fmt.Errorf("bson.ConvertDocument: %w", err)
	}

	return doc, nil
}

func MustConvertDocument(d document) *Document {
	doc, err := ConvertDocument(d)
	if err != nil {
		panic(err)
	}
	return doc
}

func (doc *Document) bsontype() {}

func (doc *Document) Map() map[string]interface{} {
	return doc.m
}

func (doc *Document) Keys() []string {
	return doc.keys
}

func (doc *Document) ReadFrom(r *bufio.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return lazyerrors.Errorf("bson.Document.ReadFrom (binary.Read): %w", err)
	}
	if l < minDocumentLen || l > MaxDocumentLen {
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
	doc.m = make(map[string]interface{})
	doc.keys = make([]string, 0, 2)

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

		doc.keys = append(doc.keys, string(ename))

		switch tag(t) {
		case tagDouble:
			var v Double
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Double): %w", err)
			}
			doc.m[string(ename)] = float64(v)

		case tagString:
			var v String
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (String): %w", err)
			}
			doc.m[string(ename)] = string(v)

		case tagDocument:
			// TODO check maximum nesting

			var v Document
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (embedded document): %w", err)
			}
			doc.m[string(ename)], err = types.ConvertDocument(&v)
			if err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (embedded document): %w", err)
			}

		case tagArray:
			// TODO check maximum nesting

			var v Array
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Array): %w", err)
			}
			doc.m[string(ename)] = types.Array(v)

		case tagBinary:
			var v Binary
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Binary): %w", err)
			}
			doc.m[string(ename)] = types.Binary(v)

		case tagUndefined:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type `Undefined (value) — Deprecated`")

		case tagObjectID:
			var v ObjectID
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (ObjectID): %w", err)
			}
			doc.m[string(ename)] = types.ObjectID(v)

		case tagBool:
			var v Bool
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Bool): %w", err)
			}
			doc.m[string(ename)] = bool(v)

		case tagDateTime:
			var v DateTime
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (DateTime): %w", err)
			}
			doc.m[string(ename)] = time.Time(v)

		case tagNull:
			doc.m[string(ename)] = nil

		case tagRegex:
			var v Regex
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Regex): %w", err)
			}
			doc.m[string(ename)] = types.Regex(v)

		case tagInt32:
			var v Int32
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Int32): %w", err)
			}
			doc.m[string(ename)] = int32(v)

		case tagTimestamp:
			var v Timestamp
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Timestamp): %w", err)
			}
			doc.m[string(ename)] = types.Timestamp(v)

		case tagInt64:
			var v Int64
			if err := v.ReadFrom(bufr); err != nil {
				return lazyerrors.Errorf("bson.Document.ReadFrom (Int64): %w", err)
			}
			doc.m[string(ename)] = int64(v)

		case tagDBPointer, tagDecimal, tagJavaScript, tagJavaScriptScope, tagMaxKey, tagMinKey, tagSymbol:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type %#02x (%s)", t, tag(t))
		default:
			return lazyerrors.Errorf("bson.Document.ReadFrom: unhandled element type %#02x (%s)", t, tag(t))
		}
	}

	if _, err := types.ConvertDocument(doc); err != nil {
		return lazyerrors.Errorf("bson.Document.ReadFrom: %w", err)
	}

	return nil
}

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

func (doc Document) MarshalBinary() ([]byte, error) {
	var elist bytes.Buffer
	bufw := bufio.NewWriter(&elist)

	for _, elK := range doc.keys {
		ename := CString(elK)
		elV, ok := doc.m[elK]
		if !ok {
			panic(fmt.Sprintf("%q not found in map", elK))
		}

		switch elV := elV.(type) {
		case float64:
			bufw.WriteByte(byte(tagDouble))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Double(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case string:
			bufw.WriteByte(byte(tagString))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := String(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Document:
			bufw.WriteByte(byte(tagDocument))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			doc, err := ConvertDocument(elV)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := doc.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Array:
			bufw.WriteByte(byte(tagArray))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Array(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Binary:
			bufw.WriteByte(byte(tagBinary))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Binary(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.ObjectID:
			bufw.WriteByte(byte(tagObjectID))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := ObjectID(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case bool:
			bufw.WriteByte(byte(tagBool))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Bool(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case time.Time:
			bufw.WriteByte(byte(tagDateTime))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := DateTime(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case nil:
			bufw.WriteByte(byte(tagNull))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Regex:
			bufw.WriteByte(byte(tagRegex))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Regex(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case int32:
			bufw.WriteByte(byte(tagInt32))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Int32(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case types.Timestamp:
			bufw.WriteByte(byte(tagTimestamp))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Timestamp(elV).WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case int64:
			bufw.WriteByte(byte(tagInt64))
			if err := ename.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err := Int64(elV).WriteTo(bufw); err != nil {
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

func unmarshalJSONValue(data []byte) (interface{}, error) {
	var v interface{}
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	if err := checkConsumed(dec, r); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res interface{}
	switch v := v.(type) {
	case map[string]interface{}:
		switch {
		case v["$f"] != nil:
			var o Double
			err = o.UnmarshalJSON(data)
			res = float64(o)
		case v["$k"] != nil:
			var o Document
			err = o.UnmarshalJSON(data)
			if err == nil {
				res, err = types.ConvertDocument(&o)
			}
		case v["$b"] != nil:
			var o Binary
			err = o.UnmarshalJSON(data)
			res = types.Binary(o)
		case v["$o"] != nil:
			var o ObjectID
			err = o.UnmarshalJSON(data)
			res = types.ObjectID(o)
		case v["$d"] != nil:
			var o DateTime
			err = o.UnmarshalJSON(data)
			res = time.Time(o)
		case v["$r"] != nil:
			var o Regex
			err = o.UnmarshalJSON(data)
			res = types.Regex(o)
		case v["$t"] != nil:
			var o Timestamp
			err = o.UnmarshalJSON(data)
			res = types.Timestamp(o)
		case v["$l"] != nil:
			var o Int64
			err = o.UnmarshalJSON(data)
			res = int64(o)
		default:
			err = lazyerrors.Errorf("unmarshalJSONValue: unhandled map %v", v)
		}
	case string:
		res = v
	case []interface{}:
		var o Array
		err = o.UnmarshalJSON(data)
		res = types.Array(o)
	case bool:
		res = v
	case nil:
		res = v
	case float64:
		res = int32(v)
	default:
		err = lazyerrors.Errorf("unmarshalJSONValue: unhandled element %[1]T (%[1]v)", v)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (doc *Document) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var rawMessages map[string]json.RawMessage
	if err := dec.Decode(&rawMessages); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	b, ok := rawMessages["$k"]
	if !ok {
		return lazyerrors.Errorf("bson.Document.UnmarshalJSON: missing $k")
	}

	var keys []string
	if err := json.Unmarshal(b, &keys); err != nil {
		return lazyerrors.Error(err)
	}
	if len(keys)+1 != len(rawMessages) {
		return lazyerrors.Errorf("bson.Document.UnmarshalJSON: %d elements in $k, %d in total", len(keys), len(rawMessages))
	}

	doc.keys = keys
	doc.m = make(map[string]interface{}, len(keys))

	for _, key := range keys {
		b, ok = rawMessages[key]
		if !ok {
			return lazyerrors.Errorf("bson.Document.UnmarshalJSON: missing key %q", key)
		}
		v, err := unmarshalJSONValue(b)
		if err != nil {
			return lazyerrors.Error(err)
		}
		doc.m[key] = v
	}

	if _, err := types.ConvertDocument(doc); err != nil {
		return lazyerrors.Errorf("bson.Document.UnmarshalJSON: %w", err)
	}

	return nil
}

func marshalJSONValue(v interface{}) ([]byte, error) {
	var o json.Marshaler
	var err error
	switch v := v.(type) {
	case float64:
		o = Double(v)
	case string:
		o = String(v)
	case types.Document:
		o, err = ConvertDocument(v)
	case types.Array:
		o = Array(v)
	case types.Binary:
		o = Binary(v)
	case types.ObjectID:
		o = ObjectID(v)
	case bool:
		o = Bool(v)
	case time.Time:
		o = DateTime(v)
	case nil:
		return []byte("null"), nil
	case types.Regex:
		o = Regex(v)
	case int32:
		o = Int32(v)
	case types.Timestamp:
		o = Timestamp(v)
	case int64:
		o = Int64(v)
	default:
		return nil, lazyerrors.Errorf("marshalJSONValue: unhandled type %T", v)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	b, err := o.MarshalJSON()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

func (doc Document) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`{"$k":`)
	b, err := json.Marshal(doc.keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	buf.Write(b)

	for _, key := range doc.keys {
		buf.WriteByte(',')

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}
		buf.Write(b)
		buf.WriteByte(':')

		value := doc.m[key]
		b, err := marshalJSONValue(value)
		if err != nil {
			return nil, lazyerrors.Errorf("bson.Document.MarshalJSON: %w", err)
		}

		buf.Write(b)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*Document)(nil)
	_ document = (*Document)(nil)
)

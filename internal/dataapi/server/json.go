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

package server

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// marshalJSON encodes wirebson.RawDocument into extended JSON.
func marshalJSON(raw wirebson.RawDocument, jsonDst io.Writer) error {
	vw, err := bsonrw.NewExtJSONValueWriter(jsonDst, false, false)
	if err != nil {
		return lazyerrors.Error(err)
	}

	encoder, err := bson.NewEncoder(vw)
	if err != nil {
		return lazyerrors.Error(err)
	}

	encoder.IntMinSize()

	err = encoder.Encode(bson.Raw(raw))
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// unmarshalSingleJSON takes extended JSON object and unmarshals it into the wirebson composite.
// If provided json is nil it also returns nil.
func unmarshalSingleJSON(json *json.RawMessage) (out any, err error) {
	if json == nil {
		return nil, nil
	}

	if len(*json) == 0 {
		return nil, fmt.Errorf("Invalid object: %v", *json)
	}

	var raw any

	err = bson.UnmarshalExtJSON(*json, false, &raw)
	if err != nil {
		return nil, err
	}

	t, b, err := bson.MarshalValue(raw)
	if err != nil {
		return nil, err
	}

	switch t {
	case bson.TypeEmbeddedDocument:
		out = wirebson.RawDocument(b)
	case bson.TypeArray:
		out = wirebson.RawArray(b)
	case bson.TypeDouble,
		bson.TypeString,
		bson.TypeBinary,
		bson.TypeUndefined,
		bson.TypeObjectID,
		bson.TypeBoolean,
		bson.TypeDateTime,
		bson.TypeNull,
		bson.TypeRegex,
		bson.TypeDBPointer,
		bson.TypeJavaScript,
		bson.TypeSymbol,
		bson.TypeCodeWithScope,
		bson.TypeInt32,
		bson.TypeTimestamp,
		bson.TypeInt64,
		bson.TypeDecimal128,
		bson.TypeMinKey,
		bson.TypeMaxKey:
		fallthrough
	default:
		panic(fmt.Errorf("Unrecognized type: %T", t))
	}

	return
}

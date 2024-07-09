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

// Package testdata provides vet tool test data.
package testdata

import (
	"time"

	"./types"
)

type tag byte

const (
	tagFloat64         = tag(0x01) // Float64
	tagString          = tag(0x02) // String
	tagDocument        = tag(0x03) // Document
	tagArray           = tag(0x04) // Array
	tagBinary          = tag(0x05) // Binary
	tagUndefined       = tag(0x06) // Undefined
	tagObjectID        = tag(0x07) // ObjectID
	tagBool            = tag(0x08) // Bool
	tagTime            = tag(0x09) // Time
	tagNull            = tag(0x0a) // Null
	tagRegex           = tag(0x0b) // Regex
	tagDBPointer       = tag(0x0c) // DBPointer
	tagJavaScript      = tag(0x0d) // JavaScript
	tagSymbol          = tag(0x0e) // Symbol
	tagJavaScriptScope = tag(0x0f) // JavaScriptScope
	tagInt32           = tag(0x10) // Int32
	tagTimestamp       = tag(0x11) // Timestamp
	tagInt64           = tag(0x12) // Int64
	tagDecimal128      = tag(0x13) // Decimal128
	tagMinKey          = tag(0xff) // MinKey
	tagMaxKey          = tag(0x7f) // MaxKey
)

func testCorrect(v any) {
	switch v.(type) {
	case *types.Document:
	case *types.Array:
	case float64, int32, int64: // multiple types
	case int8: // unexpected type
	case string:
	case types.Binary:
	case types.ObjectID:
	case bool:
	case time.Time:
	case types.NullType:
	case types.Regex:
	case types.Timestamp:
	}
}

func testIncorrectSimple(v any) {
	switch v.(type) { // want "Document should go before Array in the switch"
	case *types.Array:
	case *types.Document:
	}
}

func testIncorrectMixed(v any) {
	switch v.(type) { // want "Document should go before Time in the switch"
	case time.Time:
	case *types.Document:
	}
}

func testIncorrectMultiple(v any) {
	switch v.(type) { // want "int32 should go before int64 in the switch"
	case float64, int64, int32:
	}
}

func testCorrectTag(v tag) {
	switch v {
	case tagFloat64:
	case tagString:
	case tagDocument:
	case tagArray:
	case tagBinary:
	case tagUndefined:
	case tagObjectID:
	case tagBool:
	case tagTime:
	case tagNull:
	case tagRegex:
	case tagDBPointer:
	case tagJavaScript:
	case tagSymbol:
	case tagJavaScriptScope:
	case tagInt32:
	case tagTimestamp:
	case tagInt64:
	case tagDecimal128:
	case tagMinKey:
	case tagMaxKey:
	}
}

func testIncorrectSimpleTag(v tag) {
	switch v { // want "tagFloat64 should go before tagString in the switch"
	case tagString:
	case tagFloat64:
	}
}

func testIncorrectMixedTag(v tag) {
	switch v { // want "tagDocument should go before tagTime in the switch"
	case tagTime:
	case tagDocument:
	}
}

func testCorrectMultipleTag(v tag) {
	switch v {
	case tagArray, tagBinary, tagUndefined:
	}
}

func testIncorrectMultipleTag(v tag) {
	switch v { // want "tagBinary should go before tagUndefined in the switch"
	case tagArray, tagUndefined, tagBinary:
	}
}

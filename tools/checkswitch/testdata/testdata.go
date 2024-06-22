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

	"./tags"
	"./types"
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

func testCorrectTag(v tags.Tag) {
	switch v {
	case tags.TagFloat64:
	case tags.TagString:
	case tags.TagDocument:
	case tags.TagArray:
	case tags.TagBinary:
	case tags.TagUndefined:
	case tags.TagObjectID:
	case tags.TagBool:
	case tags.TagTime:
	case tags.TagNull:
	case tags.TagRegex:
	case tags.TagDBPointer:
	case tags.TagJavaScript:
	case tags.TagSymbol:
	case tags.TagJavaScriptScope:
	case tags.TagInt32:
	case tags.TagTimestamp:
	case tags.TagInt64:
	case tags.TagDecimal128:
	case tags.TagMinKey:
	case tags.TagMaxKey:
	}
}

func testIncorrectSimpleTag(v tags.Tag) {
	switch v { // want "tagfloat64 should go before tagstring in the switch"
	case tags.TagString:
	case tags.TagFloat64:
	}
}

func testIncorrectMixedTag(v tags.Tag) {
	switch v { // want "tagdocument should go before tagtime in the switch"
	case tags.TagTime:
	case tags.TagDocument:
	}
}

func testCorrectMultipleTag(v tags.Tag) {
	switch v {
	case tags.TagArray, tags.TagBinary, tags.TagUndefined:
	}
}

func testIncorrectMultipleTag(v tags.Tag) {
	switch v { // want "tagbinary should go before tagundefined in the switch"
	case tags.TagArray, tags.TagUndefined, tags.TagBinary:
	}
}

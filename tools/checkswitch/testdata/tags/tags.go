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

// Package tags provides stubs for testing.
package tags

type Tag byte

const (
	TagFloat64         = Tag(0x01) // Float64
	TagString          = Tag(0x02) // String
	TagDocument        = Tag(0x03) // Document
	TagArray           = Tag(0x04) // Array
	TagBinary          = Tag(0x05) // Binary
	TagUndefined       = Tag(0x06) // Undefined
	TagObjectID        = Tag(0x07) // ObjectID
	TagBool            = Tag(0x08) // Bool
	TagTime            = Tag(0x09) // Time
	TagNull            = Tag(0x0a) // Null
	TagRegex           = Tag(0x0b) // Regex
	TagDBPointer       = Tag(0x0c) // DBPointer
	TagJavaScript      = Tag(0x0d) // JavaScript
	TagSymbol          = Tag(0x0e) // Symbol
	TagJavaScriptScope = Tag(0x0f) // JavaScriptScope
	TagInt32           = Tag(0x10) // Int32
	TagTimestamp       = Tag(0x11) // Timestamp
	TagInt64           = Tag(0x12) // Int64
	TagDecimal128      = Tag(0x13) // Decimal128
	TagMinKey          = Tag(0xff) // MinKey
	TagMaxKey          = Tag(0x7f) // MaxKey
)

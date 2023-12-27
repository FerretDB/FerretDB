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

//go:generate ../../bin/stringer -linecomment -type tag

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
	tagDecimal         = tag(0x13) // Decimal
	tagMinKey          = tag(0xff) // MinKey
	tagMaxKey          = tag(0x7f) // MaxKey
)

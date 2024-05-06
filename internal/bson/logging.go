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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

// logMaxDepth is the maximum depth of a recursive representation of a BSON value.
const logMaxDepth = 20

// logMaxFlowLength is the maximum length of a flow/inline/compact representation of a BSON value.
// It may be set to 0 to always disable flow representation.
const logMaxFlowLength = 80

// nanBits is the most common pattern of a NaN float64 value, the same as math.Float64bits(math.NaN()).
const nanBits = 0b111111111111000000000000000000000000000000000000000000000000001

// slogValue returns a compact representation of any BSON value as [slog.Value].
// It may change over time.
//
// The result is optimized for small values such as function parameters.
// Some information is lost;
// for example, both int32 and int64 values are returned with [slog.KindInt64],
// arrays are treated as documents, and empty documents are omitted.
// More information is subsequently lost in handlers output;
// for example, float64(42), int32(42), and int64(42) values would all look the same
// (`f64=42 i32=42 i64=42` or `{"f64":42,"i32":42,"i64":42}`).
func slogValue(v any, depth int) slog.Value {
	switch v := v.(type) {
	case *Document:
		if v == nil {
			return slog.StringValue("Document<nil>")
		}

		if depth > logMaxDepth {
			return slog.StringValue("Document<...>")
		}

		var attrs []slog.Attr

		for _, f := range v.fields {
			attrs = append(attrs, slog.Attr{Key: f.name, Value: slogValue(f.value, depth+1)})
		}

		return slog.GroupValue(attrs...)

	case RawDocument:
		if v == nil {
			return slog.StringValue("RawDocument<nil>")
		}

		return slog.StringValue("RawDocument<" + strconv.Itoa(len(v)) + ">")

	case *Array:
		if v == nil {
			return slog.StringValue("Array<nil>")
		}

		if depth > logMaxDepth {
			return slog.StringValue("Array<...>")
		}

		var attrs []slog.Attr

		for i, v := range v.elements {
			attrs = append(attrs, slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v, depth+1)})
		}

		return slog.GroupValue(attrs...)

	case RawArray:
		if v == nil {
			return slog.StringValue("RawArray<nil>")
		}

		return slog.StringValue("RawArray<" + strconv.Itoa(len(v)) + ">")

	case float64:
		// for JSON handler to work
		switch {
		case math.IsNaN(v):
			return slog.StringValue("NaN")
		case math.IsInf(v, 1):
			return slog.StringValue("+Inf")
		case math.IsInf(v, -1):
			return slog.StringValue("-Inf")
		}

		return slog.Float64Value(v)

	case string:
		return slog.StringValue(v)

	case Binary:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case ObjectID:
		return slog.StringValue("ObjectID(" + hex.EncodeToString(v[:]) + ")")

	case bool:
		return slog.BoolValue(v)

	case time.Time:
		return slog.TimeValue(v.Truncate(time.Millisecond).UTC())

	case NullType:
		return slog.Value{}

	case Regex:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case int32:
		return slog.Int64Value(int64(v))

	case Timestamp:
		return slog.StringValue(fmt.Sprintf("%#v", v))

	case int64:
		return slog.Int64Value(v)

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

// LogMessage returns a representation as a string.
// It may change over time.
func LogMessage(v any) string {
	return logMessage(v, logMaxFlowLength, "", 1)
}

// LogMessageBlock is a variant of [RawArray.LogMessage] that never uses a flow style.
func LogMessageBlock(v any) string {
	return logMessage(v, 0, "", 1)
}

// LogMessageFlow is a variant of [RawArray.LogMessage] that always uses a flow style.
func LogMessageFlow(v any) string {
	return logMessage(v, math.MaxInt, "", 1)
}

// logMessage returns an indented representation of any BSON value as a string,
// somewhat similar (but not identical) to JSON or Go syntax.
// It may change over time.
//
// The result is optimized for large values such as full request documents.
// All information is preserved.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
// That function should be benchmarked and optimized.
func logMessage(v any, maxFlowLength int, indent string, depth int) string {
	switch v := v.(type) {
	case *Document:
		if v == nil {
			return "{<nil>}"
		}

		l := len(v.fields)
		if l == 0 {
			return "{}"
		}

		if depth > logMaxDepth {
			return "{...}"
		}

		if maxFlowLength > 0 {
			res := "{"

			for i, f := range v.fields {
				res += strconv.Quote(f.name) + `: `
				res += logMessage(f.value, maxFlowLength, "", depth+1)

				if i != l-1 {
					res += ", "
				}

				if len(res) >= maxFlowLength {
					break
				}
			}

			res += `}`

			if len(res) < maxFlowLength {
				return res
			}
		}

		res := "{\n"

		for _, f := range v.fields {
			res += indent + "  "
			res += strconv.Quote(f.name) + `: `
			res += logMessage(f.value, maxFlowLength, indent+"  ", depth+1) + ",\n"
		}

		res += indent + `}`

		return res

	case RawDocument:
		return "RawDocument<" + strconv.FormatInt(int64(len(v)), 10) + ">"

	case *Array:
		if v == nil {
			return "[<nil>]"
		}

		l := len(v.elements)
		if l == 0 {
			return "[]"
		}

		if depth > logMaxDepth {
			return "[...]"
		}

		if maxFlowLength > 0 {
			res := "["

			for i, e := range v.elements {
				res += logMessage(e, maxFlowLength, "", depth+1)

				if i != l-1 {
					res += ", "
				}

				if len(res) >= maxFlowLength {
					break
				}
			}

			res += `]`

			if len(res) < maxFlowLength {
				return res
			}
		}

		res := "[\n"

		for _, e := range v.elements {
			res += indent + "  "
			res += logMessage(e, maxFlowLength, indent+"  ", depth+1) + ",\n"
		}

		res += indent + `]`

		return res

	case RawArray:
		return "RawArray<" + strconv.FormatInt(int64(len(v)), 10) + ">"

	case float64:
		switch {
		case math.IsNaN(v):
			if bits := math.Float64bits(v); bits != nanBits {
				return fmt.Sprintf("NaN(%b)", bits)
			}

			return "NaN"

		case math.IsInf(v, 1):
			return "+Inf"
		case math.IsInf(v, -1):
			return "-Inf"
		default:
			res := strconv.FormatFloat(v, 'f', -1, 64)
			if !strings.Contains(res, ".") {
				res += ".0"
			}

			return res
		}

	case string:
		return strconv.Quote(v)

	case Binary:
		return "Binary(" + v.Subtype.String() + ":" + base64.StdEncoding.EncodeToString(v.B) + ")"

	case ObjectID:
		return "ObjectID(" + hex.EncodeToString(v[:]) + ")"

	case bool:
		return strconv.FormatBool(v)

	case time.Time:
		return v.Truncate(time.Millisecond).UTC().Format(time.RFC3339Nano)

	case NullType:
		return "null"

	case Regex:
		return "/" + v.Pattern + "/" + v.Options

	case int32:
		return strconv.FormatInt(int64(v), 10)

	case Timestamp:
		return "Timestamp(" + strconv.FormatUint(uint64(v), 10) + ")"

	case int64:
		return "int64(" + strconv.FormatInt(int64(v), 10) + ")"

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

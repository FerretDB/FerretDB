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
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

var nanBits = math.Float64bits(math.NaN())

// slogValue returns a compact representation of any BSON value as [slog.Value].
//
// The result is optimized for small values such as function parameters.
// Some information is lost;
// for example, both int32 and int64 values are returned with [slog.KindInt64],
// arrays are treated as documents, and empty documents are omitted.
// More information is subsequently lost in handlers output;
// for example, float64(42), int32(42), and int64(42) values would all look the same
// (`f64=42 i32=42 i64=42` or `{"f64":42,"i32":42,"i64":42}`).
func slogValue(v any) slog.Value {
	switch v := v.(type) {
	case *Document:
		var attrs []slog.Attr

		for _, f := range v.fields {
			attrs = append(attrs, slog.Attr{Key: f.name, Value: slogValue(f.value)})
		}

		return slog.GroupValue(attrs...)

	case RawDocument:
		if v == nil {
			return slog.StringValue("RawDocument<nil>")
		}

		return slog.StringValue("RawDocument<" + strconv.Itoa(len(v)) + ">")

	case *Array:
		var attrs []slog.Attr

		for i, v := range v.elements {
			attrs = append(attrs, slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v)})
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

func slogMessage(v any) string {
	return slogMessageIndent(v, "")
}

func slogMessageIndent(v any, indent string) string {
	switch v := v.(type) {
	case *Document:
		if len(v.fields) == 0 {
			return "{}"
		}

		res := "{\n"

		for _, f := range v.fields {
			res += indent + "\t"
			res += strconv.Quote(f.name) + `: `
			res += slogMessageIndent(f.value, indent+"\t") + ",\n"
		}

		res += indent + `}`

		return res

	case RawDocument:
		return "RawDocument<" + strconv.FormatInt(int64(len(v)), 10) + ">"

	case *Array:
		if len(v.elements) == 0 {
			return "[]"
		}

		res := "[\n"

		for _, e := range v.elements {
			res += indent + "\t"
			res += slogMessageIndent(e, indent+"\t") + ",\n"
		}

		res += indent + `]`

		return res

	case RawArray:
		panic("RawArray is not supported")

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
			if !strings.HasSuffix(res, ".0") {
				res += ".0"
			}

			return res
		}

	case string:
		return strconv.Quote(v)

	case Binary:
		panic("Binary is not supported")

	case ObjectID:
		return "ObjectID(" + hex.EncodeToString(v[:]) + ")"

	case bool:
		return strconv.FormatBool(v)

	case time.Time:
		return v.Truncate(time.Millisecond).UTC().Format(time.RFC3339Nano)

	case NullType:
		return "null"

	case Regex:
		panic("Regex is not supported")

	case int32:
		return "int32(" + strconv.FormatInt(int64(v), 10) + ")"

	case Timestamp:
		panic("Timestamp is not supported")

	case int64:
		return "int64(" + strconv.FormatInt(int64(v), 10) + ")"

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

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
	"time"
)

// slogValue returns a compact representation of any BSON value as [slog.Value].
//
// The result is optimized for small values such as function parameters.
// Some type information is lost;
// for example, both int32 and int64 values are returned with [slog.KindInt64].
// More type information is subsequently lost in handlers output;
// for example, float64(42), int32(42), and int64(42) would all look the same
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
			return slog.StringValue("RawDocument(nil)")
		}

		return slog.StringValue("RawDocument(" + strconv.Itoa(len(v)) + " bytes)")

	case *Array:
		var attrs []slog.Attr

		for i, v := range v.elements {
			attrs = append(attrs, slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v)})
		}

		return slog.GroupValue(attrs...)

	case RawArray:
		if v == nil {
			return slog.StringValue("RawArray(nil)")
		}

		return slog.StringValue("RawArray(" + strconv.Itoa(len(v)) + " bytes)")

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

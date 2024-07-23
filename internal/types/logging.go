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

package types

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"
)

// logMaxDepth is the maximum depth of a recursive representation of a BSON value.
const logMaxDepth = 20

// slogValue is a copy of [bson.slogValue] implementation modified to fit `types` value types.
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
			attrs = append(attrs, slog.Attr{Key: f.key, Value: slogValue(f.value, depth+1)})
		}

		return slog.GroupValue(attrs...)

	case *Array:
		if v == nil {
			return slog.StringValue("Array<nil>")
		}

		if depth > logMaxDepth {
			return slog.StringValue("Array<...>")
		}

		var attrs []slog.Attr

		for i, v := range v.s {
			attrs = append(attrs, slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v, depth+1)})
		}

		return slog.GroupValue(attrs...)

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

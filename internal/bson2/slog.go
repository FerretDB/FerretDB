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
	"strconv"
	"time"
)

// slogValue converts any BSON value to [slog.Value].
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
// It is not clear if slog.Value represents is a good one and even if it is handler-independent.
func slogValue(v any) slog.Value {
	switch v := v.(type) {
	case *Document:
		var attrs []slog.Attr

		for _, f := range v.fields {
			attrs = append(attrs, slog.Attr{Key: f.name, Value: slogValue(f.value)})
		}

		return slog.GroupValue(attrs...)

	case RawDocument:
		return slog.StringValue("RawDocument(" + strconv.Itoa(len(v)) + " bytes)")

	case *Array:
		var attrs []slog.Attr

		for i, v := range v.elements {
			attrs = append(attrs, slog.Attr{Key: strconv.Itoa(i), Value: slogValue(v)})
		}

		return slog.GroupValue(attrs...)

	case RawArray:
		return slog.StringValue("RawArray(" + strconv.Itoa(len(v)) + " bytes)")

	default:
		return slogScalarValue(v)
	}
}

// slogScalarValue converts any scalar BSON value to [slog.Value].
func slogScalarValue(v any) slog.Value {
	switch v := v.(type) {
	case float64:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	case string:
		return slog.StringValue(v)
	case Binary:
		return slog.AnyValue(v)
	case ObjectID:
		return slog.StringValue("ObjectID(" + hex.EncodeToString(v[:]) + ")")
	case bool:
		return slog.BoolValue(v)
	case time.Time:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	case NullType:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	case Regex:
		return slog.AnyValue(v)
	case int32:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	case Timestamp:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	case int64:
		return slog.StringValue(fmt.Sprintf("%[1]T(%[1]v)", v))
	default:
		panic(fmt.Sprintf("invalid type %T", v))
	}
}

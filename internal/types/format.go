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
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatAnyValue formats value for error message output.
func FormatAnyValue(v any) string {
	switch v := v.(type) {
	case *Document:
		return formatDocument(v)
	case *Array:
		return formatArray(v)
	case float64:
		switch {
		case math.IsNaN(v):
			return "nan.0"

		case math.IsInf(v, -1):
			return "-inf.0"
		case math.IsInf(v, +1):
			return "inf.0"
		case v == 0 && math.Signbit(v):
			return "-0.0"
		case v == 0.0:
			return "0.0"
		case v > 1000 || v < -1000 || v == math.SmallestNonzeroFloat64:
			return fmt.Sprintf("%.15e", v)
		case math.Trunc(v) == v:
			return fmt.Sprintf("%d.0", int64(v))
		default:
			res := fmt.Sprintf("%.2f", v)

			return strings.TrimSuffix(res, "0")
		}

	case string:
		return fmt.Sprintf(`"%v"`, v)
	case Binary:
		return fmt.Sprintf("BinData(%d, %X)", v.Subtype, v.B)
	case ObjectID:
		return fmt.Sprintf("ObjectId('%x')", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case time.Time:
		return fmt.Sprintf("new Date(%d)", v.UnixMilli())
	case NullType:
		return "null"
	case Regex:
		return fmt.Sprintf("/%s/%s", v.Pattern, v.Options)
	case int32:
		return fmt.Sprintf("%d", v)
	case Timestamp:
		return fmt.Sprintf("Timestamp(%v, %v)", int64(v)>>32, int32(v))
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		panic(fmt.Sprintf("unknown type %T", v))
	}
}

// formatDocument formats Document for error output.
func formatDocument(doc *Document) string {
	result := "{ "

	for i, f := range doc.fields {
		if i > 0 {
			result += ", "
		}

		result += fmt.Sprintf("%s: %s", f.key, FormatAnyValue(f.value))
	}

	return result + " }"
}

// formatArray formats Array for error output.
func formatArray(array *Array) string {
	if len(array.s) == 0 {
		return "[]"
	}

	result := "[ "

	for _, elem := range array.s {
		result += fmt.Sprintf("%s, ", FormatAnyValue(elem))
	}

	result = strings.TrimSuffix(result, ", ")

	return result + " ]"
}

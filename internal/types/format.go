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

// FormatAnyValue formats BSON value for error message output.
func FormatAnyValue(value any) string {
	assertType(value)

	switch value := value.(type) {
	case *Document:
		return formatDocument(value)
	case *Array:
		return formatArray(value)

	case float64:
		switch {
		case math.IsNaN(value):
			return "nan.0"

		case math.IsInf(value, -1):
			return "-inf.0"
		case math.IsInf(value, +1):
			return "inf.0"
		case value == 0 && math.Signbit(value):
			return "-0.0"
		case value == 0.0:
			return "0.0"
		case value > 1000 || value < -1000 || value == math.SmallestNonzeroFloat64:
			return fmt.Sprintf("%.15e", value)
		case math.Trunc(value) == value:
			return fmt.Sprintf("%d.0", int64(value))
		default:
			res := fmt.Sprintf("%.2f", value)

			return strings.TrimSuffix(res, "0")
		}

	case string:
		return fmt.Sprintf(`"%v"`, value)
	case Binary:
		return fmt.Sprintf("BinData(%d, %X)", value.Subtype, value.B)
	case ObjectID:
		return fmt.Sprintf("ObjectId('%x')", value)
	case bool:
		return fmt.Sprintf("%v", value)
	case time.Time:
		return fmt.Sprintf("new Date(%d)", value.UnixMilli())
	case NullType:
		return "null"
	case Regex:
		return fmt.Sprintf("/%s/%s", value.Pattern, value.Options)
	case int32:
		return fmt.Sprintf("%d", value)
	case Timestamp:
		return fmt.Sprintf("Timestamp(%v, %v)", int64(value)>>32, int32(value))
	case int64:
		return fmt.Sprintf("%d", value)
	default:
		panic(fmt.Sprintf("unknown type %T", value))
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

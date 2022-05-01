package types

import "fmt"

// String is needed to conveniently inform about type or value
// mismatch when returning an error. For example,
// when informing that `int32` is an invalid type, in the
// error message we would like to get `int` and not `int32`.
func String(v any) string {
	switch v.(type) {
	case NullType:
		return "null"
	case int32:
		return "int"
	case int64, float64:
		return "double"
	default:
		return fmt.Sprintf("%s", v)
	}
}

package testdata

import (
	"fmt"
	"time"

	"./types"
)

func switchOK(v interface{}) {
	switch v := v.(type) {
	case types.Document:
		fmt.Println(v)
	case *types.Array:
		fmt.Println(v)
	case float64:
		fmt.Println(v)
	case string:
		fmt.Println(v)
	case types.Binary:
		fmt.Println(v)
	case types.ObjectID:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case time.Time:
		fmt.Println(v)
	case types.NullType:
		fmt.Println(v)
	case types.Regex:
		fmt.Println(v)
	case int32:
		fmt.Println(v)
	case types.Timestamp:
		fmt.Println(v)
	case int64:
		fmt.Println(v)
	default:
		fmt.Println(v)
	}
}

func caseOK(v interface{}) {
	switch v := v.(type) {
	case *types.Document:
		fmt.Println(v)
	case *types.Array:
		fmt.Println(v)
	case float64, int32, int64:
		fmt.Println(v)
	case string:
		fmt.Println(v)
	case types.Binary:
		fmt.Println(v)
	case types.ObjectID:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case time.Time:
		fmt.Println(v)
	case types.NullType:
		fmt.Println(v)
	case types.Regex:
		fmt.Println(v)
	case types.Timestamp:
		fmt.Println(v)
	default:
		fmt.Println(v)
	}
}

func unknownTypeOK(v interface{}) {
	switch v := v.(type) {
	case types.Document:
		fmt.Println(v)
	case *types.Array:
		fmt.Println(v)
	case float64:
		fmt.Println(v)
	case int8:
		fmt.Println(v)
	case string:
		fmt.Println(v)
	case types.Binary:
		fmt.Println(v)
	case types.ObjectID:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case time.Time:
		fmt.Println(v)
	case types.NullType:
		fmt.Println(v)
	case types.Regex:
		fmt.Println(v)
	case int32:
		fmt.Println(v)
	case types.Timestamp:
		fmt.Println(v)
	case int64:
		fmt.Println(v)
	default:
		fmt.Println(v)
	}
}

func switchWrong(v interface{}) {
	switch v := v.(type) { // want "non-observance of the preferred order of types"
	case *types.Array:
		fmt.Println(v)
	case *types.Document:
		fmt.Println(v)
	case float64:
		fmt.Println(v)
	case string:
		fmt.Println(v)
	case types.Binary:
		fmt.Println(v)
	case types.ObjectID:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case time.Time:
		fmt.Println(v)
	case types.NullType:
		fmt.Println(v)
	case types.Regex:
		fmt.Println(v)
	case int32:
		fmt.Println(v)
	case types.Timestamp:
		fmt.Println(v)
	case int64:
		fmt.Println(v)
	default:
		fmt.Println(v)
	}
}

func caseWrong(v interface{}) {
	switch v := v.(type) { // want "non-observance of the preferred order of types"
	case *types.Document:
		fmt.Println(v)
	case *types.Array:
		fmt.Println(v)
	case float64, int64, int32:
		fmt.Println(v)
	case string:
		fmt.Println(v)
	case types.Binary:
		fmt.Println(v)
	case types.ObjectID:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	case time.Time:
		fmt.Println(v)
	case types.NullType:
		fmt.Println(v)
	case types.Regex:
		fmt.Println(v)
	case types.Timestamp:
		fmt.Println(v)
	default:
		fmt.Println(v)
	}
}

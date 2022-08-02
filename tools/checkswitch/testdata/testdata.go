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

package testdata

import (
	"./types"
	"fmt"
	"time"
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

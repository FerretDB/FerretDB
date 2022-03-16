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
	"bytes"
	"fmt"
	"math"
	"time"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Equal compares any BSON values.
func Equal[T Type](v1, v2 T) bool {
	return equal(v1, v2)
}

// equal compares any BSON values.
func equal(v1, v2 any) bool {
	switch v1 := v1.(type) {
	case *Document:
		d, ok := v2.(*Document)
		if !ok {
			return false
		}
		if !equalDocuments(v1, d) {
			return false
		}

	case *Array:
		a, ok := v2.(*Array)
		if !ok {
			return false
		}
		if !equalArrays(v1, a) {
			return false
		}

	default:
		if !EqualScalars(v1, v2) {
			return false
		}
	}

	return true
}

// equalDocuments compares BSON documents. Nils are not allowed.
func equalDocuments(v1, v2 *Document) bool {
	if v1 == nil {
		panic("v1 is nil")
	}
	if v2 == nil {
		panic("v2 is nil")
	}

	keys := v1.Keys()
	if !slices.Equal(keys, v2.Keys()) {
		return false
	}

	for _, k := range keys {
		f1 := must.NotFail(v1.Get(k))
		f2 := must.NotFail(v2.Get(k))
		if !equal(f1, f2) {
			return false
		}
	}

	return true
}

// equalArrays compares BSON arrays. Nils are not allowed.
func equalArrays(v1, v2 *Array) bool {
	if v1 == nil {
		panic("v1 is nil")
	}
	if v2 == nil {
		panic("v2 is nil")
	}

	l := v1.Len()
	if l != v2.Len() {
		return false
	}

	for i := 0; i < l; i++ {
		el1 := must.NotFail(v1.Get(i))
		el2 := must.NotFail(v2.Get(i))
		if !equal(el1, el2) {
			return false
		}
	}

	return true
}

// EqualScalars compares BSON scalar values in a way that is useful for tests:
//  * float64 NaNs are equal to each other;
//  * time.Time values are compared using Equal method.
func EqualScalars(v1, v2 any) bool {
	switch s1 := v1.(type) {
	case float64:
		s2, ok := v2.(float64)
		if !ok {
			s3, ok := v2.(float32)
			if !ok {
				return false
			}
			return s1 == float64(s3)
		}
		if math.IsNaN(s1) {
			return math.IsNaN(s2)
		}
		return s1 == s2
	case float32:
		s2, ok := v2.(float32)
		if !ok {
			s3, ok := v2.(float64)
			if !ok {
				return false
			}
			return float64(s1) == s3
		}
		if math.IsNaN(float64(s1)) {
			return math.IsNaN(float64(s2))
		}
		return s1 == s2

	case string:
		s2, ok := v2.(string)
		if !ok {
			return false
		}
		return s1 == s2

	case Binary:
		s2, ok := v2.(Binary)
		if !ok {
			return false
		}
		return s1.Subtype == s2.Subtype && bytes.Equal(s1.B, s2.B)

	case ObjectID:
		s2, ok := v2.(ObjectID)
		if !ok {
			return false
		}
		return s1 == s2

	case bool:
		s2, ok := v2.(bool)
		if !ok {
			return false
		}
		return s1 == s2

	case time.Time:
		s2, ok := v2.(time.Time)
		if !ok {
			return false
		}
		return s1.Equal(s2)

	case NullType:
		_, ok := v2.(NullType)
		return ok

	case Regex:
		s2, ok := v2.(Regex)
		if !ok {
			return false
		}
		return s1.Pattern == s2.Pattern && s1.Options == s2.Options

	case int32:
		s2, ok := v2.(int32)
		if !ok {
			s3, ok := v2.(int64)
			if !ok {
				return false
			}
			return int64(s1) == s3
		}
		return s1 == s2

	case Timestamp:
		s2, ok := v2.(Timestamp)
		if !ok {
			return false
		}
		return s1 == s2

	case int64:
		s2, ok := v2.(int64)
		if !ok {
			s3, ok := v2.(int32)
			if !ok {
				return false
			}
			return s1 == int64(s3)
		}
		return s1 == s2

	case CString:
		s2, ok := v2.(CString)
		if !ok {
			return false
		}
		return s1 == s2

	default:
		panic(fmt.Sprintf("unhandled types %T, %T", v1, v2))
	}
}

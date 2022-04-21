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

package shareddata

import "math"

// Scalars contain scalar values for tests.
var Scalars = &TypedValues[string]{
	data: map[string]any{
		"double":               42.13,
		"double-whole":         42.0,
		"double-zero":          0.0, // the same as math.Copysign(0, +1) in Go
		"double-negative-zero": math.Copysign(0, -1),
		"double-max":           math.MaxFloat64,
		"double-smallest":      math.SmallestNonzeroFloat64,

		"string":        "foo",
		"string-double": "42.13",
		"string-whole":  "42",
		"string-empty":  "",
	},
}

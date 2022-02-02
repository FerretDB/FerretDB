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

package common

import (
	"fmt"
	"math"
)

type Number interface {
	float64 | int32 | int64
}

func IsPositiveInteger(value any) error {
	var v float64
	switch n := value.(type) {
	case int32:
		v = float64(n)
	case float64:
		v = n
	case int64:
		v = float64(n)
	default:
		return fmt.Errorf("$size needs a number")
	}

	if math.Signbit(v) || v < 0 {
		return fmt.Errorf("$size may not be negative")
	}

	if v != math.Trunc(v) || math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("$size must be a whole number")
	}

	return nil
}

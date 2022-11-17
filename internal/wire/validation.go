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

package wire

import (
	"log"
	"math"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// validateValue checks given value and return error if not supported value was encountered.
func validateValue(v any) error {
	switch v := v.(type) {
	case *types.Document:
		for _, k := range v.Keys() {
			vv, err := v.Get(k)
			if err != nil {
				log.Fatal("can't get value from array")
			}

			if err := validateValue(vv); err != nil {
				return err
			}
		}
	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			vv, err := v.Get(i)
			if err != nil {
				log.Fatal("can't get value from array")
			}

			if err := validateValue(vv); err != nil {
				return err
			}
		}
	case float64:
		if math.IsNaN(v) {
			return lazyerrors.Errorf("NaN is not supported")
		}

		if v == 0 && math.Signbit(0) {
			return lazyerrors.Errorf("-0 is not supported")
		}
	}

	return nil
}

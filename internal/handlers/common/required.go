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

	"github.com/FerretDB/FerretDB/internal/types"
)

// GetRequiredParam returns doc's value for key or protocol error for missing or invalid parameter.
func GetRequiredParam[T types.ScalarType](doc *types.Document, key string) (T, error) {
	var zero T

	v, err := doc.Get(key)
	if err != nil {
		msg := fmt.Sprintf("required parameter %q is missing", key)
		return zero, NewErrorMsg(ErrBadValue, msg)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, NewErrorMsg(ErrBadValue, msg)
	}

	return res, nil
}

// Copyright 2022 FerretDB Inc.
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

// Sanizises scalar values
package pg

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
)

// Sanitize prepares a scalar value of any type to be passed to sql query and gives a typo to cast type in postgres query
// Sanitize(v any) (escapedString, type cast string, error)
func Sanitize(v any) (string, string, error) {
	switch value := v.(type) {
	case float64:
		val := strconv.FormatFloat(value, 'g', 6, 64)
		return pgx.Identifier{val}.Sanitize(), nil
	case string:
		return pgx.Identifier{value}.Sanitize(), nil
	case bool:
		val := fmt.Sprintf("v", value)
		return pgx.Identifier{val}.Sanitize(), nil
	case time.Time:
		return pgx.Identifier{value.String()}.Sanitize(), nil
	case int32:
		return pgx.Identifier{strconv.FormatInt(int64(value), 10)}.Sanitize(), nil
	case types.Timestamp:
		return pgx.Identifier{strconv.FormatUint(uint64(value), 10)}.Sanitize(), nil
	case int64:
		return pgx.Identifier{strconv.FormatInt(value, 10)}.Sanitize(), nil
	case types.CString:
		return pgx.Identifier{string(value)}.Sanitize(), nil
	default:
		return "", fmt.Errorf("sanitize: unsupported type: %[1]T (%[1]v)", value)
	}
}

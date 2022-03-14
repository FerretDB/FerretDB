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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Sanitize prepares a scalar value of any type to be passed to sql query
func Sanitize(v any) (val string, err error) {
	switch value := v.(type) {
	case float64:
		val = pgx.Identifier{strconv.FormatFloat(value, 'g', 6, 64)}.Sanitize()
		return
	case string:
		val = pgx.Identifier{value}.Sanitize()
		return
	case bool:
		val = pgx.Identifier{fmt.Sprintf("v", value)}.Sanitize()
		return
	case time.Time:
		val = pgx.Identifier{value.String()}.Sanitize()
		return
	case int32:
		val = pgx.Identifier{strconv.FormatInt(int64(value), 10)}.Sanitize()
		return
	case types.Timestamp:
		val = pgx.Identifier{strconv.FormatUint(uint64(value), 10)}.Sanitize()
		return
	case int64:
		val = pgx.Identifier{strconv.FormatInt(value, 10)}.Sanitize()
		return
	case types.CString:
		val = pgx.Identifier{string(value)}.Sanitize()
		return
	default:
		err = lazyerrors.Errorf("sanitize: unsupported type: %[1]T (%[1]v)", value)
		return
	}
}

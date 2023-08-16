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

package accumulators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// count represents $count operator.
type count struct{}

// newCount creates a new $count aggregation operator.
func newCount(args ...[]any) (Accumulator, error) {
	expression, err := common.GetRequiredParam[*types.Document](expr, "$count")
	if err != nil || expression.Len() != 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"$count takes no arguments, i.e. $count:{}",
			"$count (accumulator)",
		)
	}

	return new(count), nil
}

// Accumulate implements Accumulator interface.
func (c *count) Accumulate(iter types.DocumentsIterator) (any, error) {
	defer iter.Close()
	var count int32

	for {
		_, _, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, err
		}
		count++
	}

	return count, nil
}

// check interfaces
var (
	_ Accumulator = (*count)(nil)
)

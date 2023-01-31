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
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// processPopArrayUpdateExpression changes document according to $pop operator.
// If the document was changed it returns true.
func processPopArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, popValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		popValue, err := GetWholeNumberParam(popValueRaw)
		if err != nil {
			return false, NewWriteErrorMsg(ErrFailedToParse, fmt.Sprintf(`Expected a number in: %s: "%v"`, key, popValueRaw))
		}

		if popValue != 1 && popValue != -1 {
			return false, NewWriteErrorMsg(ErrFailedToParse, fmt.Sprintf("$pop expects 1 or -1, found: %d", popValue))
		}

		path := types.NewPathFromString(key)

		if !doc.HasByPath(path) {
			continue
		}

		val, err := doc.GetExactByPath(path)
		if err != nil {
			return false, err
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, NewWriteErrorMsg(
				ErrTypeMismatch,
				fmt.Sprintf("Path '%s' contains an element of non-array type '%s'", key, AliasFromType(val)),
			)
		}

		if array.Len() == 0 {
			continue
		}

		if popValue == -1 {
			array.Remove(0)
		} else {
			array.Remove(array.Len() - 1)
		}

		err = doc.SetByPath(path, array)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

// processPushArrayUpdateExpression changes document according to $push array update operator.
// If the document was changed it returns true.
func processPushArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, pushValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		path := types.NewPathFromString(key)

		// If the path does not exist, create a new array and set it.
		if !doc.HasByPath(path) {
			if err = doc.SetByPath(path, types.MakeArray(1)); err != nil {
				return false, NewWriteErrorMsg(
					ErrUnsuitableValueType,
					err.Error(),
				)
			}
		}

		val, err := doc.GetExactByPath(path)
		if err != nil {
			return false, err
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, NewWriteErrorMsg(
				ErrBadValue,
				fmt.Sprintf(
					"The field '%s' must be an array but is of type '%s' in document {_id: %s}",
					key, AliasFromType(val), must.NotFail(doc.Get("_id")),
				),
			)
		}

		array.Append(pushValueRaw)

		if err = doc.SetByPath(path, array); err != nil {
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

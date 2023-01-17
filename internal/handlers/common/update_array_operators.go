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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// processPopArrayUpdateExpression changes document according to $pop operator.
// If the document was changed it returns true.
func processPopArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	for _, key := range update.Keys() {
		popValueRaw := must.NotFail(update.Get(key))

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

		val, err := doc.GetByPath(path)
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

	for _, key := range update.Keys() {
		rawValue := must.NotFail(update.Get(key))
		print(rawValue)

		changed = true
	}

	return changed, nil
}

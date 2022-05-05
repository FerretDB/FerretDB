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

//go:generate ../../../bin/stringer -linecomment -type updateOperator

// UpdateDocument updates the given document with a series of update operators.
func UpdateDocument(doc, update *types.Document) error {
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOperators[updateOp] { //nolint:exhaustive // not implemented yet
		case updateSet:
			setDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return err
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if err = doc.Set(setKey, setValue); err != nil {
					return lazyerrors.Error(err)
				}
			}

		default:
			return NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return nil
}

// updateOperator represents update operators aliases.
type updateOperator int8

const (
	updateCurrentDate = updateOperator(1) // $currentDate
	updateInc         = updateOperator(2) // $inc
	updateMin         = updateOperator(3) // $min
	updateMax         = updateOperator(4) // $max
	updateMul         = updateOperator(5) // $mul
	updateRename      = updateOperator(6) // $rename
	updateSet         = updateOperator(7) // $set
	updateSetOnInsert = updateOperator(8) // $setOnInsert
	updateUnset       = updateOperator(9) // $unset
)

// updateOperators matches update operator string to the corresponding updateOperator value.
var updateOperators = map[string]updateOperator{}

func init() {
	for _, i := range []updateOperator{
		updateCurrentDate, updateInc, updateMin, updateMax,
		updateMul, updateRename, updateSet, updateSetOnInsert, updateUnset,
	} {
		updateOperators[i.String()] = i
	}
}

// HasUpdateOperator checks if document has updates operators.
func HasUpdateOperator(d *types.Document) bool {
	for k := range d.Map() {
		if _, ok := updateOperators[k]; ok {
			return true
		}
	}
	return false
}

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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
// Returns true if document was changed.
func UpdateDocument(doc, update *types.Document) (bool, error) {
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			msgFmt := `Modifiers operate on fields but we found type %s instead. ` +
				`For example: {$mod: {<field>: ...}} not {$set: %s}`

			if updateV == nil ||
				updateV == types.Null {
				return false, NewWriteErrorMsg(ErrFailedToParse, fmt.Sprintf(msgFmt, "null", "null"))
			}

			switch setDoc := updateV.(type) {
			case *types.Document:
				if setDoc.Len() == 0 {
					return false, nil
				}
				var err error
				for _, setKey := range setDoc.Keys() {
					setValue := must.NotFail(setDoc.Get(setKey))
					if err = doc.Set(setKey, setValue); err != nil {
						NewWriteErrorMsg(ErrFailedToParse, fmt.Sprintf(msgFmt, "array", "[]"))
						return false, err
					}
				}
				return true, nil

			case *types.Array:
				return false, NewWriteErrorMsg(ErrFailedToParse, fmt.Sprintf(msgFmt, "array", "[]"))

			case float64:
				return false, NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf(`Modifiers operate on fields but we found type double instead. `+
						`For example: {$mod: {<field>: ...}} not {$set: %.2f}`, setDoc,
					))
			default:
				return false, NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf("Modifiers operate on fields but we found type %[1]T instead. "+
						"For example: {$mod: {<field>: ...}} not {$set: \"%[1]T\"}", updateV,
					))
			}

		default:
			return false, NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return true, nil
}

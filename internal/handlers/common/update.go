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

// UpdateDocument updates the given document with a series of update operators.
func UpdateDocument(doc, update *types.Document) error {
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			switch setDoc := updateV.(type) {
			case *types.Document:
				var err error
				for _, setKey := range setDoc.Keys() {
					setValue := must.NotFail(setDoc.Get(setKey))
					if err = doc.Set(setKey, setValue); err != nil {
						return lazyerrors.Error(err)
					}
				}
				return nil

			case *types.Array:
				return NewWriteErrorMsg(
					ErrFailedToParse,
					`Modifiers operate on fields but we found type array instead. `+
						`For example: {$mod: {<field>: ...}} not {$set: []}`,
				)

			default:
				return NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf("Modifiers operate on fields but we found type %[1]T instead. "+
						"For example: {$mod: {<field>: ...}} not {$set: \"%[1]T\"}", updateV,
					))
			}

		default:
			return NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return nil
}

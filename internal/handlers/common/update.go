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
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
func UpdateDocument(doc, update *types.Document) (*types.Document, error) {
	result := doc.DeepCopy()
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			setDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return nil, err
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if err = result.Set(setKey, setValue); err != nil {
					return nil, lazyerrors.Error(err)
				}
			}

		case "$inc":
			incDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return nil, NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf(
						`Modifiers operate on fields but we found type string instead. `+
							`For example: {$mod: {<field>: ...}} not {%s: %#v}`,
						updateOp,
						updateV,
					),
				)
			}

			for _, incKey := range incDoc.Keys() {
				if strings.ContainsRune(incKey, '.') {
					return nil, NewErrorMsg(ErrNotImplemented, "dot notation not supported yet")
				}

				incValue := must.NotFail(incDoc.Get(incKey))

				if !doc.Has(incKey) {
					must.NoError(result.Set(incKey, incValue))
					return result, nil
				}

				docValue := must.NotFail(doc.Get(incKey))

				incremented, err := addNumbers(incValue, docValue)
				if err == nil {
					must.NoError(result.Set(incKey, incremented))

					return result, nil
				}

				switch err {
				case errBadLeftOperandType:
					return nil, NewWriteErrorMsg(
						ErrTypeMismatch,
						fmt.Sprintf(
							`Cannot increment with non-numeric argument: {%s: %#v}`,
							incKey,
							incValue,
						),
					)
				case errBadRightOperandType:
					return nil, NewWriteErrorMsg(
						ErrTypeMismatch,
						fmt.Sprintf(
							`Cannot apply $inc to a value of non-numeric type. `+
								`{_id: "%s"} has the field '%s' of non-numeric type %s`,
							must.NotFail(doc.Get("_id")),
							incKey,
							AliasFromType(docValue),
						),
					)
				default:
					return nil, err
				}
			}

		default:
			return nil, NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return result, nil
}

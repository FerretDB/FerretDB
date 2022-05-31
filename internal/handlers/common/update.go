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
	"sort"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
// Returns true if document was changed.
func UpdateDocument(doc, update *types.Document) (bool, error) {
	var changed bool
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			fallthrough
		case "$setOnInsert":

			switch setDoc := updateV.(type) {
			case *types.Document:
				if setDoc.Len() == 0 {
					continue
				}
				sort.Strings(setDoc.Keys())
				for _, setKey := range setDoc.Keys() {
					setValue := must.NotFail(setDoc.Get(setKey))
					if err := doc.Set(setKey, setValue); err != nil {
						return false, err
					}
				}
				changed = true
			default:
				msgFmt := fmt.Sprintf(
					`Modifiers operate on fields but we found type %[1]s instead. `+
						`For example: {$mod: {<field>: ...}} not {$set: %[1]s}`,
					AliasFromType(updateV),
				)
				return false, NewWriteErrorMsg(ErrFailedToParse, msgFmt)
			}

		case "$inc":
			incDoc, ok := updateV.(*types.Document)
			if !ok {
				return false, NewWriteErrorMsg(
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
					return false, NewErrorMsg(ErrNotImplemented, "dot notation not supported yet")
				}

				incValue := must.NotFail(incDoc.Get(incKey))

				if !doc.Has(incKey) {
					must.NoError(doc.Set(incKey, incValue))
					changed = true
					continue
				}

				docValue := must.NotFail(doc.Get(incKey))

				incremented, err := addNumbers(incValue, docValue)
				if err == nil {
					must.NoError(doc.Set(incKey, incremented))
					changed = true
					continue
				}

				switch err {
				case errUnexpectedLeftOpType:
					return false, NewWriteErrorMsg(
						ErrTypeMismatch,
						fmt.Sprintf(
							`Cannot increment with non-numeric argument: {%s: %#v}`,
							incKey,
							incValue,
						),
					)
				case errUnexpectedRightOpType:
					return false, NewWriteErrorMsg(
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
					return false, err
				}
			}

		default:
			return false, NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return changed, nil
}

// ValidateUpdateOperators validates update statement.
func ValidateUpdateOperators(update *types.Document) error {
	var err error
	set, err := getUpdateOperatorDocument("$set", update)
	if err != nil {
		return err
	}
	_, err = getUpdateOperatorDocument("$setOnInsert", update)
	if err != nil {
		return err
	}

	inc, err := getUpdateOperatorDocument("$inc", update)
	if err != nil {
		return err
	}

	if set != nil && inc != nil {
		for _, setKey := range set.Keys() {
			if inc.Has(setKey) {
				return NewWriteErrorMsg(
					ErrConflictingUpdateOperators,
					fmt.Sprintf(
						"Updating the path '%[1]s' would create a conflict at '%[1]s'", setKey,
					),
				)
			}
		}
	}

	for _, updateOp := range update.Keys() {
		switch updateOp {
		case "$inc":
			fallthrough
		case "$set":
			fallthrough
		case "$setOnInsert":
			// supported
		default:
			return NewWriteErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf(
					"Unknown modifier: %s. Expected a valid update modifier or pipeline-style "+
						"update specified as an array", updateOp,
				))
		}
	}
	return nil
}

// getUpdateOperatorDocument gets document by key `op` and returns WriteError error if it is not a document.
func getUpdateOperatorDocument(op string, update *types.Document) (*types.Document, error) {
	if !update.Has(op) {
		return nil, nil
	}
	updateExpression := must.NotFail(update.Get(op))
	switch doc := updateExpression.(type) {
	case *types.Document:
		for _, v := range doc.Keys() {
			if strings.Contains(v, ".") {
				return nil, NewError(ErrNotImplemented, fmt.Errorf("Dot notation is not implemented"))
			}
		}

		return doc, nil
	default:
		msgFmt := fmt.Sprintf(
			`Modifiers operate on fields but we found type %[1]s instead. `+
				`For example: {$mod: {<field>: ...}} not {%s: %[1]s}`,
			AliasFromType(updateExpression),
			op,
		)
		return nil, NewWriteErrorMsg(ErrFailedToParse, msgFmt)
	}
}

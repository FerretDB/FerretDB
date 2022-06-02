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
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// processCurrentdateFieldExpression changes document according to $currentDate operator.
// If the document was changed it returns true.
func processCurrentdateFieldExpression(doc *types.Document, currentDateExpression any) (bool, error) {
	var changed bool
	var err error
	switch currentDateExpression := currentDateExpression.(type) {
	case *types.Document:
		now := time.Now()
		sort.Strings(currentDateExpression.Keys())

		for _, field := range currentDateExpression.Keys() {
			setValue := must.NotFail(currentDateExpression.Get(field))

			switch setValue := setValue.(type) {
			case bool:
				if err = doc.Set(field, now); err != nil {
					return false, err
				}
				changed = true

			case *types.Document:
				for _, k := range setValue.Keys() {
					if k != "$type" {
						return false,
							NewWriteErrorMsg(
								ErrBadValue,
								fmt.Sprintf("Unrecognized $currentDate option: %s", k),
							)
					}
				}
				currentDateType, err := setValue.Get("$type")
				if err != nil { // default is date
					if err := doc.Set(field, now); err != nil {
						return false, err
					}
					changed = true
					continue
				}

				switch currentDateType := currentDateType.(type) {
				case string:
					switch currentDateType {
					case "timestamp":
						if err := doc.Set(field, now); err != nil {
							return false, err
						}
						changed = true
					case "date":
						if err := doc.Set(field, now); err != nil {
							return false, err
						}
						changed = true

					default:
						return false,
							NewWriteErrorMsg(
								ErrBadValue,
								"The '$type' string field is required to be 'date' or 'timestamp'",
							)
					}

				default:
					return false,
						NewWriteErrorMsg(
							ErrBadValue,
							"The '$type' string field is required to be 'date' or 'timestamp'",
						)
				}

			default:
				return false,
					NewWriteErrorMsg(
						ErrBadValue,
						fmt.Sprintf("%s is not valid type for $currentDate. Please use a boolean ('true') "+
							"or a $type expression ({$type: 'timestamp/date'}).", AliasFromType(setValue)),
					)
			}
		}
	default:
		return false,
			NewWriteErrorMsg(
				ErrFailedToParse,
				"Modifiers operate on fields but we found another type instead",
			)
	}
	return changed, nil
}

// UpdateDocument updates the given document with a series of update operators.
// Returns true if document was changed.
func UpdateDocument(doc, update *types.Document) (bool, error) {
	var changed bool
	var err error
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$currentDate":
			changed, err = processCurrentdateFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$set":

			switch setDoc := updateV.(type) {
			case *types.Document:
				if setDoc.Len() == 0 {
					return false, nil
				}
				sort.Strings(setDoc.Keys())
				for _, setKey := range setDoc.Keys() {
					setValue := must.NotFail(setDoc.Get(setKey))
					if err := doc.Set(setKey, setValue); err != nil {
						return false, err
					}
				}
				return true, nil
			default:
				msgFmt := fmt.Sprintf(`Modifiers operate on fields but we found type %[1]s instead. `+
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
					return true, nil
				}

				docValue := must.NotFail(doc.Get(incKey))

				incremented, err := addNumbers(incValue, docValue)
				if err == nil {
					must.NoError(doc.Set(incKey, incremented))
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

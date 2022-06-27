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

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
// Returns true if document was changed.
func UpdateDocument(doc, update *types.Document) (bool, error) {
	var changed bool
	var err error
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$currentDate":
			changed, err = processCurrentDateFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$set":
			fallthrough
		case "$setOnInsert":

			// expecting here a document since all checks were made in ValidateUpdateOperators func
			setDoc := updateV.(*types.Document)

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

		case "$unset":
			unsetDoc := updateV.(*types.Document)
			if unsetDoc.Len() == 0 {
				continue
			}
			for _, key := range unsetDoc.Keys() {
				doc.Remove(key)
			}
			changed = true

		case "$inc":
			// expecting here a document since all checks were made in ValidateUpdateOperators func
			incDoc := updateV.(*types.Document)

			for _, incKey := range incDoc.Keys() {
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
			// Treats the update as a Replacement object
			// https://www.mongodb.com/docs/manual/reference/method/db.collection.update/#std-label-update-parameter

			setDoc := update

			if setDoc.Len() == 0 {
				continue
			}
			sort.Strings(setDoc.Keys())

			for _, setKey := range doc.Keys() {
				if !setDoc.Has(setKey) && !(setKey == "_id") {
					doc.Remove(setKey)
				}
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if err := doc.Set(setKey, setValue); err != nil {
					return false, err
				}
			}

			changed = true
		}
	}

	return changed, nil
}

// processCurrentDateFieldExpression changes document according to $currentDate operator.
// If the document was changed it returns true.
func processCurrentDateFieldExpression(doc *types.Document, currentDateVal any) (bool, error) {
	var changed bool
	var err error
	currentDateExpression := currentDateVal.(*types.Document)

	now := time.Now().UTC()
	sort.Strings(currentDateExpression.Keys())

	for _, field := range currentDateExpression.Keys() {
		currentDateField := must.NotFail(currentDateExpression.Get(field))

		switch currentDateField := currentDateField.(type) {
		case bool:
			if err = doc.Set(field, now); err != nil {
				return false, err
			}
			changed = true

		case *types.Document:
			currentDateType, err := currentDateField.Get("$type")
			if err != nil { // default is date
				if err := doc.Set(field, now); err != nil {
					return false, err
				}
				changed = true
				continue
			}

			currentDateType = currentDateType.(string)
			switch currentDateType {
			case "timestamp":
				if err := doc.Set(field, types.NextTimestamp(now)); err != nil {
					return false, err
				}
				changed = true

			case "date":
				if err := doc.Set(field, now); err != nil {
					return false, err
				}
				changed = true
			}
		}
	}
	return changed, nil
}

// ValidateUpdateOperators validates update statement.
func ValidateUpdateOperators(update *types.Document) error {
	var err error
	if err = checkAllModifiersSupported(update); err != nil {
		return err
	}
	inc, err := extractValueFromUpdateOperator("$inc", update)
	if err != nil {
		return err
	}
	set, err := extractValueFromUpdateOperator("$set", update)
	if err != nil {
		return err
	}
	_, err = extractValueFromUpdateOperator("$unset", update)
	if err != nil {
		return err
	}
	_, err = extractValueFromUpdateOperator("$setOnInsert", update)
	if err != nil {
		return err
	}
	if err = checkConflictingChanges(set, inc); err != nil {
		return err
	}
	if err = validateCurrentDateExpression(update); err != nil {
		return err
	}
	return nil
}

// checkAllModifiersSupported checks that update document contains only modifiers that are supported.
func checkAllModifiersSupported(update *types.Document) error {
	for _, updateOp := range update.Keys() {
		switch updateOp {
		case "$currentDate":
			fallthrough
		case "$inc":
			fallthrough
		case "$set":
			fallthrough
		case "$setOnInsert":
			fallthrough
		case "$unset":
			// supported
		default:
			// Treats the update as a Replacement object
			// https://www.mongodb.com/docs/manual/reference/method/db.collection.update/#std-label-update-parameter
		}
	}
	return nil
}

// checkConflictingChanges checks if there are the same keys in these documents and returns an error, if any.
func checkConflictingChanges(a, b *types.Document) error {
	if a == nil {
		return nil
	}
	if b == nil {
		return nil
	}

	for _, key := range a.Keys() {
		if b.Has(key) {
			return NewWriteErrorMsg(
				ErrConflictingUpdateOperators,
				fmt.Sprintf(
					"Updating the path '%[1]s' would create a conflict at '%[1]s'", key,
				),
			)
		}
	}
	return nil
}

// extractValueFromUpdateOperator gets operator "op" value and returns WriteError error if it is not a document.
// For example, for update document
//
//  bson.D{
// 	{"$set", bson.D{{"foo", int32(12)}}},
// 	{"$inc", bson.D{{"foo", int32(1)}}},
// 	{"$setOnInsert", bson.D{{"value", math.NaN()}}},
//  }
//
// The result returned for "$setOnInsert" operator is
//  bson.D{{"value", math.NaN()}}.
func extractValueFromUpdateOperator(op string, update *types.Document) (*types.Document, error) {
	if !update.Has(op) {
		return nil, nil
	}
	updateExpression := must.NotFail(update.Get(op))
	switch doc := updateExpression.(type) {
	case *types.Document:
		for _, v := range doc.Keys() {
			if strings.Contains(v, ".") {
				return nil, NewError(ErrNotImplemented, fmt.Errorf("dot notation for operator %s is not supported yet", op))
			}
		}

		return doc, nil
	default:
		return nil, NewWriteErrorMsg(ErrFailedToParse, "Modifiers operate on fields but we found another type instead")
	}
}

// validateCurrentDateExpression validates $currentDate input on correctness.
func validateCurrentDateExpression(update *types.Document) error {
	currentDateTopField, err := update.Get("$currentDate")
	if err != nil {
		return nil // it is ok: key is absent
	}

	currentDateExpression, ok := currentDateTopField.(*types.Document)
	if !ok {
		return NewWriteErrorMsg(
			ErrFailedToParse,
			"Modifiers operate on fields but we found another type instead",
		)
	}

	for _, field := range currentDateExpression.Keys() {
		setValue := must.NotFail(currentDateExpression.Get(field))

		switch setValue := setValue.(type) {
		case bool:
			continue

		case *types.Document:
			for _, k := range setValue.Keys() {
				if k != "$type" {
					return NewWriteErrorMsg(
						ErrBadValue,
						fmt.Sprintf("Unrecognized $currentDate option: %s", k),
					)
				}
			}
			currentDateType, err := setValue.Get("$type")
			if err != nil { // ok, default is date
				continue
			}

			currentDateTypeString, ok := currentDateType.(string)
			if !ok {
				return NewWriteErrorMsg(
					ErrBadValue,
					"The '$type' string field is required to be 'date' or 'timestamp'",
				)
			}
			if !slices.Contains([]string{"date", "timestamp"}, currentDateTypeString) {
				return NewWriteErrorMsg(
					ErrBadValue,
					"The '$type' string field is required to be 'date' or 'timestamp'",
				)
			}

		default:
			return NewWriteErrorMsg(
				ErrBadValue,
				fmt.Sprintf("%s is not valid type for $currentDate. Please use a boolean ('true') "+
					"or a $type expression ({$type: 'timestamp/date'}).", AliasFromType(setValue),
				),
			)
		}
	}

	return nil
}

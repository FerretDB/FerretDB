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
	"math"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
// Returns true if document was changed.
func UpdateDocument(doc, update *types.Document) (bool, error) {
	var changed bool
	var err error

	if update.Len() == 0 {
		// replace to empty doc
		for _, key := range doc.Keys() {
			changed = true

			if key != "_id" {
				doc.Remove(key)
			}
		}

		return changed, nil
	}

	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$currentDate":
			changed, err = processCurrentDateFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$set":
			changed, err = processSetFieldExpression(doc, updateV.(*types.Document), false)
			if err != nil {
				return false, err
			}

		case "$setOnInsert":
			changed, err = processSetFieldExpression(doc, updateV.(*types.Document), true)
			if err != nil {
				return false, err
			}

		case "$unset":
			unsetDoc := updateV.(*types.Document)

			for _, key := range unsetDoc.Keys() {
				path := types.NewPathFromString(key)
				if doc.HasByPath(path) {
					doc.RemoveByPath(path)
					changed = true
				}
			}

		case "$inc":
			changed, err = processIncFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$max":
			changed, err = processMaxFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$min":
			changed, err = processMinFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$pop":
			changed, err = processPopFieldExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$rename":
			changed, err = processRenameFieldExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		default:
			if strings.HasPrefix(updateOp, "$") {
				return false, NewCommandError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
			}

			// Treats the update as a Replacement object.
			setDoc := update

			for _, setKey := range doc.Keys() {
				if !setDoc.Has(setKey) && setKey != "_id" {
					doc.Remove(setKey)
				}
			}

			setDocKeys := setDoc.Keys()
			sort.Strings(setDocKeys)

			for _, setKey := range setDocKeys {
				setValue := must.NotFail(setDoc.Get(setKey))
				doc.Set(setKey, setValue)
			}

			changed = true
		}
	}

	return changed, nil
}

// processSetFieldExpression changes document according to $set and $setOnInsert operators.
// If the document was changed it returns true.
func processSetFieldExpression(doc, setDoc *types.Document, setOnInsert bool) (bool, error) {
	var changed bool

	setDocKeys := setDoc.Keys()
	sort.Strings(setDocKeys)

	for _, setKey := range setDocKeys {
		setValue := must.NotFail(setDoc.Get(setKey))

		if setOnInsert {
			// $setOnInsert do not set null and empty array value.
			if _, ok := setValue.(types.NullType); ok {
				continue
			}

			if arr, ok := setValue.(*types.Array); ok && arr.Len() == 0 {
				continue
			}
		}

		path := types.NewPathFromString(setKey)

		if doc.HasByPath(path) {
			result := types.Compare(setValue, must.NotFail(doc.GetByPath(path)))
			if result == types.Equal {
				continue
			}
		}

		// we should insert the value if it's a regular key
		if setOnInsert && path.Len() > 1 {
			continue
		}

		err := doc.SetByPath(path, setValue)
		if err != nil {
			return false, err
		}

		changed = true
	}

	return changed, nil
}

// processPopFieldExpression changes document according to $pop operator.
// If the document was changed it returns true.
func processPopFieldExpression(doc *types.Document, update *types.Document) (bool, error) {
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
			return false, err
		}

		changed = true
	}

	return changed, nil
}

// processRenameFieldExpression changes document according to $rename operator.
// If the document was changed it returns true.
func processRenameFieldExpression(doc *types.Document, update *types.Document) (bool, error) {
	renameExpression := update.SortFieldsByKey()

	var changed bool

	// set
	// {"v":"foo"},{"}

	for _, key := range renameExpression.Keys() {
		renameRawValue, err := renameExpression.Get(key)
		if err != nil {
			// if $rename field does not exist, don't change anything
			continue
		}

		if key == "" || renameRawValue == "" {
			return changed, NewWriteErrorMsg(ErrEmptyName, "An empty update path is not valid.")
		}

		// this is covered in extractValueFromUpdateOperator
		renameValue := renameRawValue.(string)

		if key == renameRawValue {
			return changed, NewWriteErrorMsg(
				ErrBadValue,
				fmt.Sprintf("The source and target field for $rename must differ: %s: %#v", key, renameValue),
			)
		}

		sourcePath := types.NewPathFromString(key)
		targetPath := types.NewPathFromString(renameValue)

		// Get value to move
		val, err := doc.GetByPath(sourcePath)
		if err != nil {

			var dpe *types.DocumentPathError
			if !errors.As(err, &dpe) {
				panic("getByPath returned error with invalid type")
			}

			if dpe.Code() == types.ErrDocumentPathKeyNotFound {
				continue
			}

			return changed, NewWriteErrorMsg(ErrUnsuitableValueType, dpe.Error())
		}

		// Remove old document
		doc.RemoveByPath(sourcePath)

		// Set new path with old value
		err = doc.SetByPath(targetPath, val)
		if err != nil {
			return changed, err
		}

		changed = true
	}

	return changed, nil
}

// processIncFieldExpression changes document according to $inc operator.
// If the document was changed it returns true.
func processIncFieldExpression(doc *types.Document, updateV any) (bool, error) {
	// expecting here a document since all checks were made in ValidateUpdateOperators func
	incDoc := updateV.(*types.Document)

	var changed bool

	for _, incKey := range incDoc.Keys() {
		incValue := must.NotFail(incDoc.Get(incKey))

		path := types.NewPathFromString(incKey)

		if !doc.HasByPath(path) {
			// ensure incValue is a valid number type.
			switch incValue.(type) {
			case float64, int32, int64:
			default:
				return false, NewWriteErrorMsg(
					ErrTypeMismatch,
					fmt.Sprintf(`Cannot increment with non-numeric argument: {%s: %#v}`, incKey, incValue),
				)
			}

			// $inc sets the field if it does not exist.
			err := doc.SetByPath(path, incValue)
			if err != nil {
				return false, NewWriteErrorMsg(
					ErrUnsuitableValueType,
					err.Error(),
				)
			}

			changed = true

			continue
		}

		docValue, err := doc.GetByPath(types.NewPathFromString(incKey))
		if err != nil {
			return false, err
		}

		incremented, err := addNumbers(incValue, docValue)
		if err == nil {
			err := doc.SetByPath(path, incremented)
			if err != nil {
				return false, NewWriteErrorMsg(
					ErrUnsuitableValueType,
					fmt.Sprintf(`Cannot create field in element {%s: %v}`, path.Prefix(), docValue),
				)
			}

			result := types.Compare(docValue, incremented)

			docFloat, ok := docValue.(float64)
			if result == types.Equal &&
				// if the document value is NaN we should consider it as changed.
				(ok && !math.IsNaN(docFloat)) {
				continue
			}

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
			k := incKey
			if path.Len() > 1 {
				k = path.Suffix()
			}
			return false, NewWriteErrorMsg(
				ErrTypeMismatch,
				fmt.Sprintf(
					`Cannot apply $inc to a value of non-numeric type. `+
						`{_id: %s} has the field '%s' of non-numeric type %s`,
					types.FormatAnyValue(must.NotFail(doc.Get("_id"))),
					k,
					AliasFromType(docValue),
				),
			)
		case errLongExceeded:
			return false, NewWriteErrorMsg(
				ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $inc operations to current value ((NumberLong)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
			)
		case errIntExceeded:
			return false, NewWriteErrorMsg(
				ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $inc operations to current value ((NumberInt)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
			)
		default:
			return false, err
		}
	}

	return changed, nil
}

// processMaxFieldExpression changes document according to $max operator.
// If the document was changed it returns true.
func processMaxFieldExpression(doc *types.Document, updateV any) (bool, error) {
	maxExpression := updateV.(*types.Document)
	maxExpression = maxExpression.SortFieldsByKey()

	var changed bool

	for _, field := range maxExpression.Keys() {
		val, _ := doc.Get(field)

		maxVal, err := maxExpression.Get(field)
		if err != nil {
			// if max field does not exist, don't change anything
			continue
		}

		// if the document value was found, compare it with max value
		if val != nil {
			res := types.CompareOrder(val, maxVal, types.Ascending)
			switch res {
			case types.Equal:
				fallthrough
			case types.Greater:
				continue
			case types.Less:
				// if document value is less than max value, update the value
			default:
				return changed, NewCommandErrorMsgWithArgument(
					ErrNotImplemented,
					"document comparison is not implemented",
					"$max",
				)
			}
		}

		doc.Set(field, maxVal)
		changed = true
	}

	return changed, nil
}

// processMinFieldExpression changes document according to $min operator.
// If the document was changed it returns true.
func processMinFieldExpression(doc *types.Document, updateV any) (bool, error) {
	minExpression := updateV.(*types.Document)
	minExpression = minExpression.SortFieldsByKey()

	var changed bool

	for _, field := range minExpression.Keys() {
		minVal, err := minExpression.Get(field)
		if err != nil {
			// if min field does not exist, don't change anything
			continue
		}

		val, _ := doc.Get(field)

		// if the document value was found, compare it with min value
		if val != nil {
			res := types.CompareOrder(val, minVal, types.Ascending)
			switch res {
			case types.Equal:
				fallthrough
			case types.Less:
				continue
			case types.Greater:
			}
		}

		doc.Set(field, minVal)

		changed = true
	}

	return changed, nil
}

// processCurrentDateFieldExpression changes document according to $currentDate operator.
// If the document was changed it returns true.
func processCurrentDateFieldExpression(doc *types.Document, currentDateVal any) (bool, error) {
	var changed bool
	currentDateExpression := currentDateVal.(*types.Document)

	now := time.Now().UTC()
	keys := currentDateExpression.Keys()
	sort.Strings(keys)

	for _, field := range keys {
		currentDateField := must.NotFail(currentDateExpression.Get(field))

		switch currentDateField := currentDateField.(type) {
		case *types.Document:
			currentDateType, err := currentDateField.Get("$type")
			if err != nil { // default is date
				doc.Set(field, now)
				changed = true
				continue
			}

			currentDateType = currentDateType.(string)
			switch currentDateType {
			case "timestamp":
				doc.Set(field, types.NextTimestamp(now))
				changed = true

			case "date":
				doc.Set(field, now)
				changed = true
			}

		case bool:
			doc.Set(field, now)
			changed = true
		}
	}
	return changed, nil
}

// ValidateUpdateOperators validates update statement.
func ValidateUpdateOperators(update *types.Document) error {
	var err error
	if _, err = HasSupportedUpdateModifiers(update); err != nil {
		return err
	}

	_, err = extractValueFromUpdateOperator("$currentDate", update)
	if err != nil {
		return err
	}

	inc, err := extractValueFromUpdateOperator("$inc", update)
	if err != nil {
		return err
	}

	_, err = extractValueFromUpdateOperator("$max", update)
	if err != nil {
		return err
	}

	_, err = extractValueFromUpdateOperator("$min", update)
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

	_, err = extractValueFromUpdateOperator("$pop", update)
	if err != nil {
		return err
	}

	_, err = extractValueFromUpdateOperator("$rename", update)
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

// HasSupportedUpdateModifiers checks that update document contains only modifiers that are supported.
func HasSupportedUpdateModifiers(update *types.Document) (bool, error) {
	updateModifier := false
	for _, updateOp := range update.Keys() {
		switch updateOp {
		case "$currentDate":
			fallthrough
		case "$inc":
			fallthrough
		case "$max":
			fallthrough
		case "$min":
			fallthrough
		case "$set":
			fallthrough
		case "$setOnInsert":
			fallthrough
		case "$unset":
			fallthrough
		case "$pop":
			fallthrough
		case "$rename":
			updateModifier = true
		case "$mul":
			return false, NewCommandErrorMsgWithArgument(
				ErrNotImplemented,
				fmt.Sprintf("update operator %s is not implemented", updateOp),
				updateOp,
			)
		default:
			if strings.HasPrefix(updateOp, "$") {
				return false, NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf(
						"Unknown modifier: %s. Expected a valid update modifier or pipeline-style "+
							"update specified as an array", updateOp,
					),
				)
			}

			// In case the operator doesn't start with $, treats the update as a Replacement object
		}
	}

	return updateModifier, nil
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
//	 bson.D{
//		{"$set", bson.D{{"foo", int32(12)}}},
//		{"$inc", bson.D{{"foo", int32(1)}}},
//		{"$setOnInsert", bson.D{{"v", nil}}},
//	 }
//
// The result returned for "$setOnInsert" operator is
//
//	bson.D{{"v", nil}}.
func extractValueFromUpdateOperator(op string, update *types.Document) (*types.Document, error) {
	if !update.Has(op) {
		return nil, nil
	}
	updateExpression := must.NotFail(update.Get(op))

	doc, ok := updateExpression.(*types.Document)
	if !ok {
		return nil, NewWriteErrorMsg(ErrFailedToParse, "Modifiers operate on fields but we found another type instead")
	}

	if op == "$rename" {
		iter := doc.Iterator()
		keys := map[string]struct{}{}

		for {
			k, v, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}
				return nil, err
			}

			vStr, ok := v.(string)
			if !ok {
				return nil, NewWriteErrorMsg(
					ErrBadValue,
					fmt.Sprintf("The 'to' field for $rename must be a string: %s: %v", k, vStr),
				)
			}

			if k == vStr {
				return nil, NewWriteErrorMsg(ErrBadValue, fmt.Sprintf("The source and target field for $rename must differ: %s: %v", k, vStr))
			}

			if _, ok := keys[k]; ok {
				return nil, NewWriteErrorMsg(ErrConflictingUpdateOperators, fmt.Sprintf("Updating the '%s' would create a conflict at '%s'", k, k))
			}

			keys[k] = struct{}{}

			if _, ok := keys[vStr]; ok {
				return nil, NewWriteErrorMsg(ErrConflictingUpdateOperators, fmt.Sprintf("Updating the '%s' would create a conflict at '%s'", vStr, vStr))
			}

			keys[vStr] = struct{}{}
		}
	}

	duplicate, ok := doc.FindDuplicateKey()
	if ok {
		return nil, NewWriteErrorMsg(
			ErrConflictingUpdateOperators,
			fmt.Sprintf(
				"Updating the path '%[1]s' would create a conflict at '%[1]s'", duplicate,
			),
		)
	}

	return doc, nil
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

		case bool:
			continue

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

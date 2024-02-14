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
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument iterates through documents from iter and processes them sequentially based on param.
// Returns UpdateResult if all operations (update/upsert) are successful.
//
// In case of updating multiple documents, UpdateDocument returns an error immediately after one of the
// operation fails. The rest of the documents are not processed.
// TODO https://github.com/FerretDB/FerretDB/issues/2612
func UpdateDocument(ctx context.Context, c backends.Collection, cmd string, iter types.DocumentsIterator, param *Update) (*UpdateResult, error) { //nolint:lll // for readability
	result := new(UpdateResult)

	docs, err := iterator.ConsumeValues(iter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	result.Matched.Count = int32(len(docs))

	var upsert bool

	if len(docs) == 0 {
		if !param.Upsert {
			return result, nil
		}

		upsert = true
		docs = append(docs, must.NotFail(types.NewDocument()))
	}

	for _, doc := range docs {
		var modified bool

		result.Matched.Doc = doc.DeepCopy()

		if upsert {
			if err = processFilterEqualityCondition(doc, param.Filter); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		if !param.HasUpdateOperators {
			modified, err = processReplacementDoc(cmd, doc, param.Update)
		} else {
			modified, err = processUpdateOperator(cmd, doc, param.Update, upsert)
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !doc.Has("_id") {
			doc.Set("_id", types.NewObjectID())
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/3454
		if err = doc.ValidateData(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if upsert {
			_, err = c.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{doc}})
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			result.Upserted.Doc = doc
		} else if modified {
			_, err := c.UpdateAll(ctx, &backends.UpdateAllParams{Docs: []*types.Document{doc}})
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			result.Modified.Doc = doc
			result.Modified.Count++
		}
	}

	return result, nil
}

// processFilterEqualityCondition copies the fields with equality condition from filter to doc.
func processFilterEqualityCondition(doc, filter *types.Document) error {
	iter := filter.Iterator()
	defer iter.Close()

	for {
		key, val, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return nil
			}

			return lazyerrors.Error(err)
		}

		if key[0] == '$' {
			continue
		}

		if valDoc, ok := val.(*types.Document); ok {
			if valDoc.Len() == 1 {
				valEq, _ := valDoc.Get("$eq")
				if valEq == nil {
					continue
				}
				val = valEq
			} else {
				val = valDoc
			}
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return lazyerrors.Error(err)
		}

		err = doc.SetByPath(path, val)
		if err != nil {
			return lazyerrors.Error(err)
		}
	}
}

// processReplacementDoc replaces the given document with a new document while retaining its
// original _id.
// Returns true if the document is changed. Returns error when _id is attempted to be changed.
func processReplacementDoc(command string, doc, update *types.Document) (bool, error) {
	if types.Compare(doc, update) == types.Equal {
		return false, nil
	}

	docId, _ := doc.Get("_id")
	updatedId, _ := update.Get("_id")

	if docId != nil && updatedId != nil && types.Compare(docId, updatedId) != types.Equal {
		return false, newUpdateError(
			handlererrors.ErrImmutableField,
			"Performing an update on the path '_id' would modify the immutable field '_id'",
			command,
		)
	}

	var changed bool

	for _, key := range doc.Keys() {
		if key != "_id" {
			doc.Remove(key)
			changed = true
		}
	}

	for _, key := range update.Keys() {
		doc.Set(key, must.NotFail(update.Get(key)))
		changed = true
	}

	return changed, nil
}

// processUpdateOperator updates the given document with a series of update operators.
// Returns true if the document is changed.
// Returns CommandError if the command is findAndModify, otherwise returns WriteError.
// TODO https://github.com/FerretDB/FerretDB/issues/3044
func processUpdateOperator(command string, doc, update *types.Document, upsert bool) (bool, error) {
	var docUpdated bool
	var err error

	docId, _ := doc.Get("_id")

	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		var updated bool

		switch updateOp {
		case "$currentDate":
			updated, err = processCurrentDateFieldExpression(doc, updateV)
			if err != nil {
				return false, err
			}

		case "$set":
			updated, err = processSetFieldExpression(command, doc, updateV.(*types.Document), false)
			if err != nil {
				return false, err
			}

		case "$setOnInsert":
			if !upsert {
				continue
			}

			updated, err = processSetFieldExpression(command, doc, updateV.(*types.Document), true)
			if err != nil {
				return false, err
			}

		case "$unset":
			// updateV is document, checked in ValidateUpdateOperators.
			unsetDoc := updateV.(*types.Document)

			for _, key := range unsetDoc.Keys() {
				var path types.Path

				path, err = types.NewPathFromString(key)
				if err != nil {
					// ValidateUpdateOperators checked already $unset contains valid path.
					panic(err)
				}

				if doc.HasByPath(path) {
					doc.RemoveByPath(path)
					updated = true
				}
			}

		case "$inc":
			updated, err = processIncFieldExpression(command, doc, updateV)
			if err != nil {
				return false, err
			}

		case "$max":
			updated, err = processMaxFieldExpression(command, doc, updateV)
			if err != nil {
				return false, err
			}

		case "$min":
			updated, err = processMinFieldExpression(command, doc, updateV)
			if err != nil {
				return false, err
			}

		case "$mul":
			if updated, err = processMulFieldExpression(command, doc, updateV); err != nil {
				return false, err
			}

		case "$rename":
			updated, err = processRenameFieldExpression(command, doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$pop":
			updated, err = processPopArrayUpdateExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$push":
			updated, err = processPushArrayUpdateExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$addToSet":
			updated, err = processAddToSetArrayUpdateExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$pullAll":
			updated, err = processPullAllArrayUpdateExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$pull":
			updated, err = processPullArrayUpdateExpression(doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		case "$bit":
			updated, err = processBitFieldExpression(command, doc, updateV.(*types.Document))
			if err != nil {
				return false, err
			}

		default:
			if strings.HasPrefix(updateOp, "$") {
				return false, newUpdateError(
					handlererrors.ErrNotImplemented,
					fmt.Sprintf("UpdateDocument: unhandled operation %q", updateOp),
					command,
				)
			}
		}

		docUpdated = docUpdated || updated
	}

	updatedId, _ := doc.Get("_id")
	if types.Compare(docId, updatedId) != types.Equal {
		return false, newUpdateError(
			handlererrors.ErrImmutableField,
			"Performing an update on the path '_id' would modify the immutable field '_id'",
			command,
		)
	}

	return docUpdated, nil
}

// processSetFieldExpression changes document according to $set and $setOnInsert operators.
// If the document was changed it returns true.
func processSetFieldExpression(command string, doc, setDoc *types.Document, setOnInsert bool) (bool, error) {
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

		// setKey has valid path, checked in ValidateUpdateOperators.
		path := must.NotFail(types.NewPathFromString(setKey))

		if doc.HasByPath(path) {
			docValue := must.NotFail(doc.GetByPath(path))
			if types.Identical(setValue, docValue) {
				continue
			}
		}

		// we should insert the value if it's a regular key
		if setOnInsert && path.Len() > 1 {
			continue
		}

		if err := doc.SetByPath(path, setValue); err != nil {
			return false, newUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
		}

		changed = true
	}

	return changed, nil
}

// processRenameFieldExpression changes document according to $rename operator.
// If the document was changed it returns true.
func processRenameFieldExpression(command string, doc *types.Document, update *types.Document) (bool, error) {
	update.SortFieldsByKey()

	var changed bool

	for _, key := range update.Keys() {
		renameRawValue := must.NotFail(update.Get(key))

		if key == "" || renameRawValue == "" {
			return changed, newUpdateError(
				handlererrors.ErrEmptyName,
				"An empty update path is not valid.",
				command,
			)
		}

		// this is covered in validateRenameExpression
		renameValue := renameRawValue.(string)

		sourcePath, err := types.NewPathFromString(key)
		if err != nil {
			var pathErr *types.PathError
			if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
				return false, newUpdateError(
					handlererrors.ErrEmptyName,
					fmt.Sprintf(
						"The update path '%s' contains an empty field name, which is not allowed.",
						key,
					),
					command,
				)
			}
		}

		targetPath, err := types.NewPathFromString(renameValue)
		if err != nil {
			return changed, lazyerrors.Error(err)
		}

		// Get value to move
		val, err := doc.GetByPath(sourcePath)
		if err != nil {
			var dpe *types.PathError
			if !errors.As(err, &dpe) {
				panic("getByPath returned error with invalid type")
			}

			if dpe.Code() == types.ErrPathKeyNotFound || dpe.Code() == types.ErrPathIndexOutOfBound {
				continue
			}

			if dpe.Code() == types.ErrPathIndexInvalid {
				return false, newUpdateError(
					handlererrors.ErrUnsuitableValueType,
					fmt.Sprintf("cannot use path '%s' to traverse the document", sourcePath),
					command,
				)
			}

			return changed, newUpdateError(handlererrors.ErrUnsuitableValueType, dpe.Error(), command)
		}

		// Remove old document
		doc.RemoveByPath(sourcePath)

		// Set new path with old value
		if err := doc.SetByPath(targetPath, val); err != nil {
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

// processIncFieldExpression changes document according to $inc operator.
// If the document was changed it returns true.
func processIncFieldExpression(command string, doc *types.Document, updateV any) (bool, error) {
	// updateV is document, checked in ValidateUpdateOperators.
	incDoc := updateV.(*types.Document)

	var changed bool

	for _, incKey := range incDoc.Keys() {
		incValue := must.NotFail(incDoc.Get(incKey))

		// ensure incValue is a valid number type.
		switch incValue.(type) {
		case float64, int32, int64:
		default:
			return false, newUpdateError(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(`Cannot increment with non-numeric argument: {%s: %#v}`, incKey, incValue),
				command,
			)
		}

		var err error

		// incKey has valid path, checked in ValidateUpdateOperators.
		path := must.NotFail(types.NewPathFromString(incKey))

		if !doc.HasByPath(path) {
			// $inc sets the field if it does not exist.
			if err := doc.SetByPath(path, incValue); err != nil {
				return false, newUpdateError(
					handlererrors.ErrUnsuitableValueType,
					err.Error(),
					command,
				)
			}

			changed = true

			continue
		}

		path, err = types.NewPathFromString(incKey)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		docValue, err := doc.GetByPath(path)
		if err != nil {
			return false, err
		}

		incremented, err := addNumbers(incValue, docValue)
		if err == nil {
			if err = doc.SetByPath(path, incremented); err != nil {
				return false, lazyerrors.Error(err)
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

		switch {
		case errors.Is(err, handlerparams.ErrUnexpectedRightOpType):
			k := incKey
			if path.Len() > 1 {
				k = path.Suffix()
			}

			return false, newUpdateError(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(
					`Cannot apply $inc to a value of non-numeric type. `+
						`{_id: %s} has the field '%s' of non-numeric type %s`,
					types.FormatAnyValue(must.NotFail(doc.Get("_id"))),
					k,
					handlerparams.AliasFromType(docValue),
				),
				command,
			)
		case errors.Is(err, handlerparams.ErrLongExceededPositive), errors.Is(err, handlerparams.ErrLongExceededNegative):
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $inc operations to current value ((NumberLong)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
				command,
			)
		case errors.Is(err, handlerparams.ErrIntExceeded):
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $inc operations to current value ((NumberInt)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
				command,
			)
		default:
			return false, lazyerrors.Error(err)
		}
	}

	return changed, nil
}

// processMaxFieldExpression changes document according to $max operator.
// If the document was changed it returns true.
func processMaxFieldExpression(command string, doc *types.Document, updateV any) (bool, error) {
	maxExpression := updateV.(*types.Document)
	maxExpression.SortFieldsByKey()

	var changed bool

	iter := maxExpression.Iterator()
	defer iter.Close()

	for {
		maxKey, maxVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		// maxKey has valid path, checked in ValidateUpdateOperators.
		path := must.NotFail(types.NewPathFromString(maxKey))

		if !doc.HasByPath(path) {
			err = doc.SetByPath(path, maxVal)
			if err != nil {
				return false, newUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
			}

			changed = true
			continue
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, lazyerrors.Error(err)
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
				return changed, handlererrors.NewCommandErrorMsgWithArgument(
					handlererrors.ErrNotImplemented,
					"document comparison is not implemented",
					"$max",
				)
			}
		}

		if err = doc.SetByPath(path, maxVal); err != nil {
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

// processMinFieldExpression changes document according to $min operator.
// If the document was changed it returns true.
func processMinFieldExpression(command string, doc *types.Document, updateV any) (bool, error) {
	minExpression := updateV.(*types.Document)
	minExpression.SortFieldsByKey()

	var changed bool

	iter := minExpression.Iterator()
	defer iter.Close()

	for {
		minKey, minVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		// minKey has valid path, checked in ValidateUpdateOperators.
		path := must.NotFail(types.NewPathFromString(minKey))

		if !doc.HasByPath(path) {
			err = doc.SetByPath(path, minVal)
			if err != nil {
				return false, newUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
			}

			changed = true
			continue
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

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

		if err = doc.SetByPath(path, minVal); err != nil {
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

// processMulFieldExpression updates document according to $mul operator.
// If the document was changed it returns true.
func processMulFieldExpression(command string, doc *types.Document, updateV any) (bool, error) {
	// updateV is document, checked in ValidateUpdateOperators.
	mulDoc := updateV.(*types.Document)

	var changed bool

	for _, mulKey := range mulDoc.Keys() {
		mulValue := must.NotFail(mulDoc.Get(mulKey))

		var path types.Path
		var err error

		path, err = types.NewPathFromString(mulKey)
		if err != nil {
			// ValidateUpdateOperators checked already $mul contains valid path.
			panic(err)
		}

		if !doc.HasByPath(path) {
			// $mul sets the field to zero if the field does not exist.
			switch mulValue.(type) {
			case float64:
				mulValue = float64(0)
			case int32:
				mulValue = int32(0)
			case int64:
				mulValue = int64(0)
			default:
				return false, newUpdateError(
					handlererrors.ErrTypeMismatch,
					fmt.Sprintf(`Cannot multiply with non-numeric argument: {%s: %#v}`, mulKey, mulValue),
					command,
				)
			}

			err := doc.SetByPath(path, mulValue)
			if err != nil {
				return false, newUpdateError(
					handlererrors.ErrUnsuitableValueType,
					err.Error(),
					command,
				)
			}

			changed = true

			continue
		}

		docValue, err := doc.GetByPath(path)
		if err != nil {
			return false, err
		}

		var multiplied any
		multiplied, err = multiplyNumbers(mulValue, docValue)

		switch {
		case err == nil:
			if multiplied, ok := multiplied.(float64); ok && math.IsInf(multiplied, 0) {
				return false, handlererrors.NewCommandErrorMsg(
					handlererrors.ErrBadValue,
					fmt.Sprintf("update produces invalid value: { %q: %f } "+
						"(update operations that produce infinity values are not allowed)", path, multiplied,
					),
				)
			}

			err = doc.SetByPath(path, multiplied)
			if err != nil {
				// after successfully getting value from path, setting it back cannot fail.
				panic(err)
			}

			// A change from int32(0) to int64(0) is considered changed.
			// Hence, do not use types.Compare(docValue, multiplied) because
			// it will equate int32(0) == int64(0).
			if docValue == multiplied {
				continue
			}

			changed = true

			continue

		case errors.Is(err, handlerparams.ErrUnexpectedLeftOpType):
			return false, newUpdateError(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(
					`Cannot multiply with non-numeric argument: {%s: %#v}`,
					mulKey,
					mulValue,
				),
				command,
			)
		case errors.Is(err, handlerparams.ErrUnexpectedRightOpType):
			k := mulKey
			if path.Len() > 1 {
				k = path.Suffix()
			}

			return false, newUpdateError(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(
					`Cannot apply $mul to a value of non-numeric type. `+
						`{_id: %s} has the field '%s' of non-numeric type %s`,
					types.FormatAnyValue(must.NotFail(doc.Get("_id"))),
					k,
					handlerparams.AliasFromType(docValue),
				),
				command,
			)
		case errors.Is(err, handlerparams.ErrLongExceededPositive), errors.Is(err, handlerparams.ErrLongExceededNegative):
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $mul operations to current value ((NumberLong)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
				command,
			)
		case errors.Is(err, handlerparams.ErrIntExceeded):
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					`Failed to apply $mul operations to current value ((NumberInt)%d) for document {_id: "%s"}`,
					docValue,
					must.NotFail(doc.Get("_id")),
				),
				command,
			)
		default:
			return false, err
		}
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

// processBitFieldExpression updates document according to $bit operator.
// If document was changed, it returns true.
func processBitFieldExpression(command string, doc *types.Document, updateV any) (bool, error) {
	var changed bool

	bitDoc := updateV.(*types.Document)
	for _, bitKey := range bitDoc.Keys() {
		bitValue := must.NotFail(bitDoc.Get(bitKey))

		nestedDoc, ok := bitValue.(*types.Document)
		if !ok {
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					`The $bit modifier is not compatible with a %s. `+
						`You must pass in an embedded document: {$bit: {field: {and/or/xor: #}}`,
					handlerparams.AliasFromType(bitValue),
				),
				command,
			)
		}

		if nestedDoc.Len() == 0 {
			return false, newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(
					"You must pass in at least one bitwise operation. "+
						`The format is: {$bit: {field: {and/or/xor: #}}`,
				),
				command,
			)
		}

		// bitKey has valid path, checked in ValidateUpdateOperators
		path := must.NotFail(types.NewPathFromString(bitKey))

		// $bit sets the field if it does not exist by applying bitwise operat on 0 and operand value.
		var docValue any = int32(0)

		hasPath := doc.HasByPath(path)
		if hasPath {
			docValue = must.NotFail(doc.GetByPath(path))
		}

		for _, bitOp := range nestedDoc.Keys() {
			bitOpValue := must.NotFail(nestedDoc.Get(bitOp))

			bitOpResult, err := performBitLogic(bitOp, bitOpValue, docValue)

			switch {
			case err == nil:
				if err = doc.SetByPath(path, bitOpResult); err != nil {
					return false, newUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
				}

				if docValue == bitOpResult && hasPath {
					continue
				}

				changed = true

				continue

			case errors.Is(err, handlerparams.ErrUnexpectedLeftOpType):
				return false, newUpdateError(
					handlererrors.ErrBadValue,
					fmt.Sprintf(
						`The $bit modifier field must be an Integer(32/64 bit); a `+
							`'%s' is not supported here: {%s: %s}`,
						handlerparams.AliasFromType(bitOpValue),
						bitOp,
						types.FormatAnyValue(bitOpValue),
					),
					command,
				)

			case errors.Is(err, handlerparams.ErrUnexpectedRightOpType):
				return false, newUpdateError(
					handlererrors.ErrBadValue,
					fmt.Sprintf(
						`Cannot apply $bit to a value of non-integral type.`+
							`_id: %s has the field %s of non-integer type %s`,
						types.FormatAnyValue(must.NotFail(doc.Get("_id"))),
						path.Suffix(),
						handlerparams.AliasFromType(docValue),
					),
					command,
				)

			default:
				return false, newUpdateError(handlererrors.ErrBadValue, err.Error(), command)
			}
		}
	}

	return changed, nil
}

// ValidateUpdateOperators validates update statement.
// ValidateUpdateOperators returns CommandError for findAndModify case-insensitive command name,
// WriteError for other commands.
func ValidateUpdateOperators(command string, update *types.Document) error {
	var err error

	currentDate, err := extractValueFromUpdateOperator(command, "$currentDate", update)
	if err != nil {
		return err
	}

	inc, err := extractValueFromUpdateOperator(command, "$inc", update)
	if err != nil {
		return err
	}

	max, err := extractValueFromUpdateOperator(command, "$max", update)
	if err != nil {
		return err
	}

	min, err := extractValueFromUpdateOperator(command, "$min", update)
	if err != nil {
		return err
	}

	mul, err := extractValueFromUpdateOperator(command, "$mul", update)
	if err != nil {
		return err
	}

	set, err := extractValueFromUpdateOperator(command, "$set", update)
	if err != nil {
		return err
	}

	unset, err := extractValueFromUpdateOperator(command, "$unset", update)
	if err != nil {
		return err
	}

	setOnInsert, err := extractValueFromUpdateOperator(command, "$setOnInsert", update)
	if err != nil {
		return err
	}

	_, err = extractValueFromUpdateOperator(command, "$rename", update)
	if err != nil {
		return err
	}

	pop, err := extractValueFromUpdateOperator(command, "$pop", update)
	if err != nil {
		return err
	}

	push, err := extractValueFromUpdateOperator(command, "$push", update)
	if err != nil {
		return err
	}

	addToSet, err := extractValueFromUpdateOperator(command, "$addToSet", update)
	if err != nil {
		return err
	}

	pullAll, err := extractValueFromUpdateOperator(command, "$pullAll", update)
	if err != nil {
		return err
	}

	pull, err := extractValueFromUpdateOperator(command, "$pull", update)
	if err != nil {
		return err
	}

	bit, err := extractValueFromUpdateOperator(command, "$bit", update)
	if err != nil {
		return err
	}

	if err = validateOperatorKeys(
		command,
		addToSet,
		currentDate,
		inc,
		min,
		max,
		mul,
		pop,
		pull,
		pullAll,
		push,
		set,
		setOnInsert,
		unset,
		bit,
	); err != nil {
		return err
	}

	if err = validateCurrentDateExpression(command, update); err != nil {
		return err
	}

	if err = validateRenameExpression(command, update); err != nil {
		return err
	}

	return nil
}

// HasSupportedUpdateModifiers checks if the update document contains supported update operators.
//
// Returns false when no update operators are found. Returns an error when the document contains
// unsupported operators or when there is a mix of operators and non-$-prefixed fields.
func HasSupportedUpdateModifiers(command string, update *types.Document) (bool, error) {
	var updateOps int

	for _, operator := range update.Keys() {
		switch operator {
		case // field update operators:
			"$currentDate",
			"$inc", "$min", "$max", "$mul",
			"$rename",
			"$set", "$setOnInsert", "$unset",
			"$bit",

			// array update operators:
			"$pop", "$push", "$addToSet", "$pullAll", "$pull":
			updateOps++
		default:
			if strings.HasPrefix(operator, "$") {
				return false, newUpdateError(
					handlererrors.ErrFailedToParse,
					fmt.Sprintf(
						"Unknown modifier: %s. Expected a valid update modifier or pipeline-style "+
							"update specified as an array", operator,
					),
					command,
				)
			}

			// In case the operator doesn't start with $, treat the update as a replacement document
		}
	}

	if updateOps > 0 && updateOps != update.Len() {
		// update contains a mix of non-$-prefixed fields (replacement document) and operators
		return false, newUpdateError(
			handlererrors.ErrDollarPrefixedFieldName,
			"The dollar ($) prefixed field is not allowed in the context of an update's replacement document.",
			command,
		)
	}

	return (updateOps > 0), nil
}

// newUpdateError returns CommandError for findAndModify command, WriteError for other commands.
func newUpdateError(code handlererrors.ErrorCode, msg, command string) error {
	// Depending on the driver, the command may be camel case or lower case.
	if strings.ToLower(command) == "findandmodify" {
		return handlererrors.NewCommandErrorMsgWithArgument(code, msg, command)
	}

	return handlererrors.NewWriteErrorMsg(code, msg)
}

// validateOperatorKeys returns error if any key contains empty path or
// the same path prefix exists in other key or other document.
func validateOperatorKeys(command string, docs ...*types.Document) error {
	var visitedPaths []types.Path

	for _, doc := range docs {
		for _, key := range doc.Keys() {
			nextPath, err := types.NewPathFromString(key)
			if err != nil {
				return newUpdateError(
					handlererrors.ErrEmptyName,
					fmt.Sprintf(
						"The update path '%s' contains an empty field name, which is not allowed.",
						key,
					),
					command,
				)
			}

			err = types.IsConflictPath(visitedPaths, nextPath)
			var pathErr *types.PathError

			if errors.As(err, &pathErr) {
				if pathErr.Code() == types.ErrPathConflictOverwrite ||
					pathErr.Code() == types.ErrPathConflictCollision {
					return newUpdateError(
						handlererrors.ErrConflictingUpdateOperators,
						fmt.Sprintf(
							"Updating the path '%[1]s' would create a conflict at '%[1]s'", key,
						),
						command,
					)
				}
			}

			if err != nil {
				return lazyerrors.Error(err)
			}

			visitedPaths = append(visitedPaths, nextPath)
		}
	}

	return nil
}

// extractValueFromUpdateOperator gets operator "op" value and returns CommandError for `findAndModify`
// WriteError error other commands if it is not a document.
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
//
// It also checks for path collisions and returns the error if there's any.
func extractValueFromUpdateOperator(command, op string, update *types.Document) (*types.Document, error) {
	if !update.Has(op) {
		return nil, nil
	}
	updateExpression := must.NotFail(update.Get(op))

	doc, ok := updateExpression.(*types.Document)
	if !ok {
		return nil, newUpdateError(
			handlererrors.ErrFailedToParse,
			fmt.Sprintf(`Modifiers operate on fields but we found type %[1]s instead. `+
				`For example: {$mod: {<field>: ...}} not {%s: %s}`,
				handlerparams.AliasFromType(updateExpression),
				op,
				types.FormatAnyValue(updateExpression),
			),
			command,
		)
	}

	duplicate, ok := doc.FindDuplicateKey()
	if ok {
		return nil, newUpdateError(
			handlererrors.ErrConflictingUpdateOperators,
			fmt.Sprintf(
				"Updating the path '%[1]s' would create a conflict at '%[1]s'", duplicate,
			),
			command,
		)
	}

	return doc, nil
}

// validateRenameExpression validates $rename input on correctness.
func validateRenameExpression(command string, update *types.Document) error {
	if !update.Has("$rename") {
		return nil
	}

	updateExpression := must.NotFail(update.Get("$rename"))

	// updateExpression is document, checked in ValidateUpdateOperators.
	doc := updateExpression.(*types.Document)

	iter := doc.Iterator()
	defer iter.Close()

	keys := map[string]struct{}{}

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return lazyerrors.Error(err)
		}

		var vStr string

		vStr, ok := v.(string)
		if !ok {
			return newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf("The 'to' field for $rename must be a string: %s: %v", k, v),
				command,
			)
		}

		// disallow fields where key is equal to the target
		if k == vStr {
			return newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(`The source and target field for $rename must differ: %s: "%[1]s"`, k, vStr),
				command,
			)
		}

		if _, ok = keys[k]; ok {
			return newUpdateError(
				handlererrors.ErrConflictingUpdateOperators,
				fmt.Sprintf("Updating the path '%s' would create a conflict at '%s'", k, k),
				command,
			)
		}

		keys[k] = struct{}{}

		if _, ok = keys[vStr]; ok {
			return newUpdateError(
				handlererrors.ErrConflictingUpdateOperators,
				fmt.Sprintf("Updating the path '%s' would create a conflict at '%s'", vStr, vStr),
				command,
			)
		}

		keys[vStr] = struct{}{}
	}

	return nil
}

// validateCurrentDateExpression validates $currentDate input on correctness.
func validateCurrentDateExpression(command string, update *types.Document) error {
	currentDateTopField, err := update.Get("$currentDate")
	if err != nil {
		return nil // it is ok: key is absent
	}

	// currentDateExpression is document, checked in ValidateUpdateOperators.
	currentDateExpression := currentDateTopField.(*types.Document)

	for _, field := range currentDateExpression.Keys() {
		setValue := must.NotFail(currentDateExpression.Get(field))

		switch setValue := setValue.(type) {
		case *types.Document:
			for _, k := range setValue.Keys() {
				if k != "$type" {
					return newUpdateError(
						handlererrors.ErrBadValue,
						fmt.Sprintf("Unrecognized $currentDate option: %s", k),
						command,
					)
				}
			}
			currentDateType, err := setValue.Get("$type")
			if err != nil { // ok, default is date
				continue
			}

			currentDateTypeString, ok := currentDateType.(string)
			if !ok || !slices.Contains([]string{"date", "timestamp"}, currentDateTypeString) {
				return newUpdateError(
					handlererrors.ErrBadValue,
					"The '$type' string field is required to be 'date' or 'timestamp': "+
						"{$currentDate: {field : {$type: 'date'}}}",
					command,
				)
			}

		case bool:
			continue

		default:
			return newUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf("%s is not valid type for $currentDate. Please use a boolean ('true') "+
					"or a $type expression ({$type: 'timestamp/date'}).", handlerparams.AliasFromType(setValue),
				),
				command,
			)
		}
	}

	return nil
}

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

// kvOp represents key-value pair and its associated update operator.
type kvOp struct {
	Key      string
	Value    any
	Operator string
}

// UpdateDocument iterates through documents from iter and processes them sequentially based on param.
// Returns UpdateResult if all operations (update/upsert) are successful.
//
// In case of updating multiple documents, UpdateDocument returns an error immediately after one of the
// operation fails. The rest of the documents are not processed.
// TODO https://github.com/FerretDB/FerretDB/issues/2612
func UpdateDocument(ctx context.Context, c backends.Collection, cmd string, iter types.DocumentsIterator, param *Update) (*UpdateResult, error) { //nolint:lll // for readability
	result := new(UpdateResult)

	isFindAndModify := (strings.ToLower(cmd) == "findandmodify")

	for {
		var upsert, modified bool

		_, doc, err := iter.Next()
		if err != nil {
			if !errors.Is(err, iterator.ErrIteratorDone) {
				return nil, lazyerrors.Error(err)
			}

			if result.Matched.Count == 0 && param.Upsert {
				upsert = true
				doc = must.NotFail(types.NewDocument())
			}

			if !upsert {
				return result, nil
			}
		}

		if upsert {
			if err = processFilterEqualityCondition(doc, param.Filter); err != nil {
				return nil, lazyerrors.Error(err)
			}
		} else {
			result.Matched.Count++
			if isFindAndModify {
				result.Matched.Doc = doc.DeepCopy()
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

			// upsert happens only once, no need to iterate further
			return result, nil
		} else if modified {
			_, err := c.UpdateAll(ctx, &backends.UpdateAllParams{Docs: []*types.Document{doc}})
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			result.Modified.Count++
			if isFindAndModify {
				result.Modified.Doc = doc
			}
		}
	}
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

		if key[0] == '$' { // logical operators like $and, $or, $not
			continue
		}

		if valDoc, ok := val.(*types.Document); ok {
			if valDoc.Len() == 1 {
				innerKey := valDoc.Keys()[0]

				if innerKey[0] != '$' {
					// valDoc does not contain an operator
					val = valDoc
				} else if innerKey == "$eq" {
					// valDoc has $eq operator, so extract the inner value
					val, _ = valDoc.Get("$eq")
				} else {
					// valDoc has operators like $lt, $gt, $ne, $in, $exists, $regex
					continue
				}
			} else {
				// valDoc is a sub-document with many key/val pairs, without any operators
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
		return false, NewUpdateError(
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

	for _, kvOp := range getSortedKVOps(update) {
		var updated bool

		key, value := kvOp.Key, kvOp.Value

		switch kvOp.Operator {
		case "$currentDate":
			updated, err = processCurrentDateFieldExpression(doc, key, value)
			if err != nil {
				return false, err
			}

		case "$set":
			updated, err = processSetFieldExpression(command, doc, key, value, false)
			if err != nil {
				return false, err
			}

		case "$setOnInsert":
			if !upsert {
				continue
			}

			updated, err = processSetFieldExpression(command, doc, key, value, true)
			if err != nil {
				return false, err
			}

		case "$unset":
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

		case "$inc":
			updated, err = processIncFieldExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$max":
			updated, err = processMaxFieldExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$min":
			updated, err = processMinFieldExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$mul":
			if updated, err = processMulFieldExpression(command, doc, key, value); err != nil {
				return false, err
			}

		case "$rename":
			updated, err = processRenameFieldExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$pop":
			updated, err = processPopArrayUpdateExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$push":
			updated, err = processPushArrayUpdateExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$addToSet":
			updated, err = processAddToSetArrayUpdateExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$pull":
			updated, err = processPullArrayUpdateExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$pullAll":
			updated, err = processPullAllArrayUpdateExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		case "$bit":
			updated, err = processBitFieldExpression(command, doc, key, value)
			if err != nil {
				return false, err
			}

		default:
			if strings.HasPrefix(kvOp.Operator, "$") {
				return false, NewUpdateError(
					handlererrors.ErrNotImplemented,
					fmt.Sprintf("UpdateDocument: unhandled operation %q", kvOp.Operator),
					command,
				)
			}
		}

		docUpdated = docUpdated || updated
	}

	updatedId, _ := doc.Get("_id")
	if docId != nil && (updatedId == nil || types.Compare(docId, updatedId) != types.Equal) {
		return false, NewUpdateError(
			handlererrors.ErrImmutableField,
			"Performing an update on the path '_id' would modify the immutable field '_id'",
			command,
		)
	}

	return docUpdated, nil
}

// getSortedKVOps extracts key-value pairs and associated operators from update document
// and sorts them based on lexicographic order of keys.
func getSortedKVOps(update *types.Document) []*kvOp {
	kvOps := []*kvOp{}

	iter := update.Iterator()
	defer iter.Close()

	for {
		operator, opVal, err := iter.Next()
		if err == iterator.ErrIteratorDone { //nolint:errorlint // only ErrIteratorDone could be returned
			break
		}

		opDoc := opVal.(*types.Document)

		opDocIter := opDoc.Iterator()

		for {
			var key string
			var val any

			key, val, err = opDocIter.Next()
			if err == iterator.ErrIteratorDone { //nolint:errorlint // only ErrIteratorDone could be returned
				opDocIter.Close()
				break
			}

			kvOps = append(kvOps, &kvOp{
				Key:      key,
				Value:    val,
				Operator: operator,
			})
		}

		opDocIter.Close()
	}

	slices.SortFunc(kvOps, func(a *kvOp, b *kvOp) int {
		return strings.Compare(a.Key, b.Key)
	})

	return kvOps
}

// processSetFieldExpression changes document according to $set and $setOnInsert operators.
// If the document was changed it returns true.
func processSetFieldExpression(command string, doc *types.Document, setKey string, setValue any, setOnInsert bool) (bool, error) {
	if setOnInsert {
		// $setOnInsert do not set null and empty array value.
		if _, ok := setValue.(types.NullType); ok {
			return false, nil
		}

		if arr, ok := setValue.(*types.Array); ok && arr.Len() == 0 {
			return false, nil
		}
	}

	// setKey has valid path, checked in ValidateUpdateOperators.
	path := must.NotFail(types.NewPathFromString(setKey))

	if doc.HasByPath(path) {
		docValue := must.NotFail(doc.GetByPath(path))
		if types.Identical(setValue, docValue) {
			return false, nil
		}
	}

	if err := doc.SetByPath(path, setValue); err != nil {
		return false, NewUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
	}

	return true, nil
}

// processRenameFieldExpression changes document according to $rename operator.
// If the document was changed it returns true.
func processRenameFieldExpression(command string, doc *types.Document, key string, value any) (bool, error) {
	var changed bool

	if key == "" || value == "" {
		return changed, NewUpdateError(
			handlererrors.ErrEmptyName,
			"An empty update path is not valid.",
			command,
		)
	}

	// this is covered in validateRenameExpression
	newKey := value.(string)

	sourcePath, err := types.NewPathFromString(key)
	if err != nil {
		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
			return false, NewUpdateError(
				handlererrors.ErrEmptyName,
				fmt.Sprintf(
					"The update path '%s' contains an empty field name, which is not allowed.",
					key,
				),
				command,
			)
		}
	}

	targetPath, err := types.NewPathFromString(newKey)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	// Get value to move
	val, err := doc.GetByPath(sourcePath)
	if err != nil {
		var dpe *types.PathError
		if !errors.As(err, &dpe) {
			panic("getByPath returned error with invalid type")
		}

		if dpe.Code() == types.ErrPathKeyNotFound || dpe.Code() == types.ErrPathIndexOutOfBound {
			return false, nil
		}

		if dpe.Code() == types.ErrPathIndexInvalid {
			return false, NewUpdateError(
				handlererrors.ErrUnsuitableValueType,
				fmt.Sprintf("cannot use path '%s' to traverse the document", sourcePath),
				command,
			)
		}

		return false, NewUpdateError(handlererrors.ErrUnsuitableValueType, dpe.Error(), command)
	}

	// Remove old document
	doc.RemoveByPath(sourcePath)

	// Set new path with old value
	if err = doc.SetByPath(targetPath, val); err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// processIncFieldExpression changes document according to $inc operator.
// If the document was changed it returns true.
func processIncFieldExpression(command string, doc *types.Document, incKey string, incValue any) (bool, error) {
	// ensure incValue is a valid number type.
	switch incValue.(type) {
	case float64, int32, int64:
	default:
		return false, NewUpdateError(
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
		if err = doc.SetByPath(path, incValue); err != nil {
			return false, NewUpdateError(
				handlererrors.ErrUnsuitableValueType,
				err.Error(),
				command,
			)
		}

		return true, nil
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
			return false, nil
		}

		return true, nil
	}

	switch {
	case errors.Is(err, handlerparams.ErrUnexpectedRightOpType):
		k := incKey
		if path.Len() > 1 {
			k = path.Suffix()
		}

		return false, NewUpdateError(
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
		return false, NewUpdateError(
			handlererrors.ErrBadValue,
			fmt.Sprintf(
				`Failed to apply $inc operations to current value ((NumberLong)%d) for document {_id: "%s"}`,
				docValue,
				must.NotFail(doc.Get("_id")),
			),
			command,
		)
	case errors.Is(err, handlerparams.ErrIntExceeded):
		return false, NewUpdateError(
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

// processMaxFieldExpression changes document according to $max operator.
// If the document was changed it returns true.
func processMaxFieldExpression(command string, doc *types.Document, maxKey string, maxValue any) (bool, error) {
	// maxKey has valid path, checked in ValidateUpdateOperators.
	path := must.NotFail(types.NewPathFromString(maxKey))

	if !doc.HasByPath(path) {
		err := doc.SetByPath(path, maxValue)
		if err != nil {
			return false, NewUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
		}

		return true, nil
	}

	val, err := doc.GetByPath(path)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	// if the document value was found, compare it with max value
	if val != nil {
		res := types.CompareOrder(val, maxValue, types.Ascending)
		switch res {
		case types.Equal, types.Greater:
			return false, nil
		case types.Less:
			// if document value is less than max value, update the value
		default:
			return false, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrNotImplemented,
				"document comparison is not implemented",
				"$max",
			)
		}
	}

	if err = doc.SetByPath(path, maxValue); err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// processMinFieldExpression changes document according to $min operator.
// If the document was changed it returns true.
func processMinFieldExpression(command string, doc *types.Document, minKey string, minValue any) (bool, error) {
	// minKey has valid path, checked in ValidateUpdateOperators.
	path := must.NotFail(types.NewPathFromString(minKey))

	if !doc.HasByPath(path) {
		err := doc.SetByPath(path, minValue)
		if err != nil {
			return false, NewUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
		}

		return true, nil
	}

	val, err := doc.GetByPath(path)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	// if the document value was found, compare it with min value
	if val != nil {
		res := types.CompareOrder(val, minValue, types.Ascending)
		switch res {
		case types.Equal, types.Less:
			return false, nil
		case types.Greater:
		}
	}

	if err = doc.SetByPath(path, minValue); err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// processMulFieldExpression updates document according to $mul operator.
// If the document was changed it returns true.
func processMulFieldExpression(command string, doc *types.Document, mulKey string, mulValue any) (bool, error) {
	// $mul contains valid path, checked in ValidateUpdateOperators.
	path := must.NotFail(types.NewPathFromString(mulKey))

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
			return false, NewUpdateError(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(`Cannot multiply with non-numeric argument: {%s: %#v}`, mulKey, mulValue),
				command,
			)
		}

		err := doc.SetByPath(path, mulValue)
		if err != nil {
			return false, NewUpdateError(
				handlererrors.ErrUnsuitableValueType,
				err.Error(),
				command,
			)
		}

		return true, nil
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

		// after successfully getting value from path, setting it back cannot fail.
		must.NoError(doc.SetByPath(path, multiplied))

		// A change from int32(0) to int64(0) is considered changed.
		// Hence, do not use types.Compare(docValue, multiplied) because
		// it will equate int32(0) == int64(0).
		if docValue == multiplied {
			return false, nil
		}

		return true, nil

	case errors.Is(err, handlerparams.ErrUnexpectedLeftOpType):
		return false, NewUpdateError(
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

		return false, NewUpdateError(
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
		return false, NewUpdateError(
			handlererrors.ErrBadValue,
			fmt.Sprintf(
				`Failed to apply $mul operations to current value ((NumberLong)%d) for document {_id: "%s"}`,
				docValue,
				must.NotFail(doc.Get("_id")),
			),
			command,
		)
	case errors.Is(err, handlerparams.ErrIntExceeded):
		return false, NewUpdateError(
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

// processCurrentDateFieldExpression changes document according to $currentDate operator.
// If the document was changed it returns true.
func processCurrentDateFieldExpression(doc *types.Document, field string, value any) (bool, error) {
	var changed bool

	now := time.Now().UTC()

	// refers to BSON types, either `Date` or `timestamp`
	var setValType any

	switch value := value.(type) {
	case *types.Document:
		// ignore error, error cases were validated in validateCurrentDateExpression
		setValType, _ = value.Get("$type")
	case bool:
		// if boolean value is passed, then `Date` type is used for setting current date.
		setValType = "date"
	}

	switch setValType {
	case "date":
		doc.Set(field, now)
		changed = true
	case "timestamp":
		doc.Set(field, types.NextTimestamp(now))
		changed = true
	}

	return changed, nil
}

// processBitFieldExpression updates document according to $bit operator.
// If document was changed, it returns true.
func processBitFieldExpression(command string, doc *types.Document, bitKey string, bitValue any) (bool, error) {
	bitDoc, ok := bitValue.(*types.Document)
	if !ok {
		return false, NewUpdateError(
			handlererrors.ErrBadValue,
			fmt.Sprintf(
				`The $bit modifier is not compatible with a %s. `+
					`You must pass in an embedded document: {$bit: {field: {and/or/xor: #}}`,
				handlerparams.AliasFromType(bitValue),
			),
			command,
		)
	}

	if bitDoc.Len() == 0 {
		return false, NewUpdateError(
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

	var changed bool

	for _, bitOp := range bitDoc.Keys() {
		bitOpValue := must.NotFail(bitDoc.Get(bitOp))

		bitOpResult, err := performBitLogic(bitOp, bitOpValue, docValue)

		switch {
		case err == nil:
			if err = doc.SetByPath(path, bitOpResult); err != nil {
				return false, NewUpdateError(handlererrors.ErrUnsuitableValueType, err.Error(), command)
			}

			if !hasPath || docValue != bitOpResult {
				changed = true
			}

			continue

		case errors.Is(err, handlerparams.ErrUnexpectedLeftOpType):
			return false, NewUpdateError(
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
			return false, NewUpdateError(
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
			return false, NewUpdateError(handlererrors.ErrBadValue, err.Error(), command)
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
				return false, NewUpdateError(
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
		return false, NewUpdateError(
			handlererrors.ErrDollarPrefixedFieldName,
			"The dollar ($) prefixed field is not allowed in the context of an update's replacement document.",
			command,
		)
	}

	return (updateOps > 0), nil
}

// NewUpdateError returns CommandError for findAndModify command, WriteError for other commands.
func NewUpdateError(code handlererrors.ErrorCode, msg, command string) error {
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
				return NewUpdateError(
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
					return NewUpdateError(
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
		return nil, NewUpdateError(
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
		return nil, NewUpdateError(
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
			return NewUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf("The 'to' field for $rename must be a string: %s: %v", k, v),
				command,
			)
		}

		// disallow fields where key is equal to the target
		if k == vStr {
			return NewUpdateError(
				handlererrors.ErrBadValue,
				fmt.Sprintf(`The source and target field for $rename must differ: %s: "%[1]s"`, k, vStr),
				command,
			)
		}

		if _, ok = keys[k]; ok {
			return NewUpdateError(
				handlererrors.ErrConflictingUpdateOperators,
				fmt.Sprintf("Updating the path '%s' would create a conflict at '%s'", k, k),
				command,
			)
		}

		keys[k] = struct{}{}

		if _, ok = keys[vStr]; ok {
			return NewUpdateError(
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
					return NewUpdateError(
						handlererrors.ErrBadValue,
						fmt.Sprintf("Unrecognized $currentDate option: %s", k),
						command,
					)
				}
			}

			currentDateType, _ := setValue.Get("$type")

			currentDateTypeString, ok := currentDateType.(string)
			if !ok || !slices.Contains([]string{"date", "timestamp"}, currentDateTypeString) {
				return NewUpdateError(
					handlererrors.ErrBadValue,
					"The '$type' string field is required to be 'date' or 'timestamp': "+
						"{$currentDate: {field : {$type: 'date'}}}",
					command,
				)
			}

		case bool:
			continue

		default:
			return NewUpdateError(
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

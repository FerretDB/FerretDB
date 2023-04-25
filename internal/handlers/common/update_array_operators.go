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

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// processPopArrayUpdateExpression changes document according to $pop operator.
// If the document was changed it returns true.
func processPopArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, popValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		popValue, err := commonparams.GetWholeNumberParam(popValueRaw)
		if err != nil {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf(`Expected a number in: %s: "%v"`, key, popValueRaw),
			)
		}

		if popValue != 1 && popValue != -1 {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf("$pop expects 1 or -1, found: %d", popValue),
			)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			// If any sub path exists in the doc, $pop returns ErrUnsuitableValueType.
			if err = checkUnsuitableValueError(doc, key, path); err != nil {
				return false, err
			}

			// doc does not have a path, nothing to do.
			continue
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf("Path '%s' contains an element of non-array type '%s'", key, commonparams.AliasFromType(val)),
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
			return false, lazyerrors.Error(err)
		}

		changed = true
	}

	return changed, nil
}

// checkUnsuitableValueError returns ErrUnsuitableValueType if path contains
// a non-document value. If no element exists on path, it returns nil.
// For example, if the path is "v.foo" and
// - doc is {v: 42}, it returns ErrUnsuitableValueType, v is used by unsuitable value type;
// - doc is {c: 10}, it returns no error since the path does not exist.
func checkUnsuitableValueError(doc *types.Document, key string, path types.Path) error {
	// return no error if path is suffix or key.
	if path.Len() == 1 {
		return nil
	}

	prefix := path.Prefix()

	// check if part of the path exists in the document.
	if doc.Has(prefix) {
		val := must.NotFail(doc.Get(prefix))
		if prefixDoc, ok := val.(*types.Document); ok {
			// recursively check if document contains part of the part.
			return checkUnsuitableValueError(prefixDoc, key, path.TrimPrefix())
		}

		// ErrUnsuitableValueType is returned if the document contains prefix.
		return commonerrors.NewWriteErrorMsg(
			commonerrors.ErrUnsuitableValueType,
			fmt.Sprintf(
				"Cannot use the part (%s) of (%s) to traverse the element ({%s: %v})",
				path.Slice()[1],
				key,
				prefix,
				types.FormatAnyValue(val),
			),
		)
	}

	// no part of the path exists in the doc.
	return nil
}

// processPushArrayUpdateExpression changes document according to $push array update operator.
// If the document was changed it returns true.
func processPushArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, pushValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		var each *types.Array

		if pushValue, ok := pushValueRaw.(*types.Document); ok {
			if pushValue.Has("$each") {
				eachRaw := must.NotFail(pushValue.Get("$each"))

				each, ok = eachRaw.(*types.Array)
				if !ok {
					return false, commonerrors.NewWriteErrorMsg(
						commonerrors.ErrBadValue,
						fmt.Sprintf(
							"The argument to $each in $push must be an array but it was of type: %s",
							commonparams.AliasFromType(eachRaw),
						),
					)
				}
			}
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		// If the path does not exist, create a new array and set it.
		if !doc.HasByPath(path) {
			changed = true

			if err = doc.SetByPath(path, types.MakeArray(1)); err != nil {
				return false, commonerrors.NewWriteErrorMsg(
					commonerrors.ErrUnsuitableValueType,
					err.Error(),
				)
			}
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, err
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"The field '%s' must be an array but is of type '%s' in document {_id: %s}",
					key, commonparams.AliasFromType(val), must.NotFail(doc.Get("_id")),
				),
			)
		}

		if each == nil {
			each = types.MakeArray(1)
			each.Append(pushValueRaw)
		}

		for i := 0; i < each.Len(); i++ {
			array.Append(must.NotFail(each.Get(i)))

			changed = true
		}

		if err = doc.SetByPath(path, array); err != nil {
			return false, lazyerrors.Error(err)
		}
	}

	return changed, nil
}

// processAddToSetArrayUpdateExpression changes document according to $addToSet array update operator.
// If the document was changed it returns true.
func processAddToSetArrayUpdateExpression(doc, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, addToSetValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		var each *types.Array

		if addToSetValue, ok := addToSetValueRaw.(*types.Document); ok {
			if addToSetValue.Has("$each") {
				eachRaw := must.NotFail(addToSetValue.Get("$each"))

				each, ok = eachRaw.(*types.Array)
				if !ok {
					return false, commonerrors.NewWriteErrorMsg(
						commonerrors.ErrTypeMismatch,
						fmt.Sprintf(
							"The argument to $each in $addToSet must be an array but it was of type %s",
							commonparams.AliasFromType(eachRaw),
						),
					)
				}
			}
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		// If the path does not exist, create a new array and set it.
		if !doc.HasByPath(path) {
			changed = true

			if err = doc.SetByPath(path, types.MakeArray(1)); err != nil {
				return false, commonerrors.NewWriteErrorMsg(
					commonerrors.ErrUnsuitableValueType,
					err.Error(),
				)
			}
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, err
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"The field '%s' must be an array but is of type '%s' in document {_id: %s}",
					key, commonparams.AliasFromType(val), must.NotFail(doc.Get("_id")),
				),
			)
		}

		if each == nil {
			each = types.MakeArray(1)
			each.Append(addToSetValueRaw)
		}

		for i := 0; i < each.Len(); i++ {
			value := must.NotFail(each.Get(i))

			if array.Contains(value) {
				continue
			}

			changed = true

			array.Append(value)
		}

		if err = doc.SetByPath(path, array); err != nil {
			return false, lazyerrors.Error(err)
		}
	}

	return changed, nil
}

// processPullAllArrayUpdateExpression changes document according to $pullAll array update operator.
// If the document was changed it returns true.
func processPullAllArrayUpdateExpression(doc, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, pullAllValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		// If the path does not exist, create a new array and set it.
		if !doc.HasByPath(path) {
			if err = doc.SetByPath(path, types.MakeArray(1)); err != nil {
				return false, commonerrors.NewWriteErrorMsg(
					commonerrors.ErrUnsuitableValueType,
					err.Error(),
				)
			}
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"The field '%s' must be an array but is of type '%s' in document {_id: %s}",
					key, commonparams.AliasFromType(val), must.NotFail(doc.Get("_id")),
				),
			)
		}

		pullAllArray, ok := pullAllValueRaw.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"The field '%s' must be an array but is of type '%s'",
					key, commonparams.AliasFromType(pullAllValueRaw),
				),
			)
		}

		for j := 0; j < pullAllArray.Len(); j++ {
			var valueToPull any

			valueToPull, err = pullAllArray.Get(j)
			if err != nil {
				return false, lazyerrors.Error(err)
			}

			// we remove all instances of valueToPull in array
			for i := array.Len() - 1; i >= 0; i-- {
				var value any

				value, err = array.Get(i)
				if err != nil {
					return false, lazyerrors.Error(err)
				}

				if types.Compare(value, valueToPull) == types.Equal {
					array.Remove(i)

					changed = true
				}
			}
		}

		if err = doc.SetByPath(path, array); err != nil {
			return false, lazyerrors.Error(err)
		}
	}

	return changed, nil
}

// processPullArrayUpdateExpression changes document according to $pull array update operator.
// If the document was changed it returns true.
func processPullArrayUpdateExpression(doc *types.Document, update *types.Document) (bool, error) {
	var changed bool

	iter := update.Iterator()
	defer iter.Close()

	for {
		key, pullValueRaw, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false, lazyerrors.Error(err)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		// If the path does not exist, try to set it.
		// If it is not possible to set the path return error.
		// This will deal with cases like $pull on a non-existing field or unset field.
		if !doc.HasByPath(path) {
			if err = doc.SetByPath(path, types.MakeArray(1)); err != nil {
				return false, commonerrors.NewWriteErrorMsg(
					commonerrors.ErrUnsuitableValueType,
					err.Error(),
				)
			}
		}

		val, err := doc.GetByPath(path)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		array, ok := val.(*types.Array)
		if !ok {
			return false, commonerrors.NewWriteErrorMsg(
				commonerrors.ErrBadValue,
				"Cannot apply $pull to a non-array value",
			)
		}

		for i := array.Len() - 1; i >= 0; i-- {
			value := must.NotFail(array.Get(i))

			if types.Compare(value, pullValueRaw) == types.Equal {
				array.Remove(i)

				changed = true
			}
		}

		if err = doc.SetByPath(path, array); err != nil {
			return false, lazyerrors.Error(err)
		}
	}

	return changed, nil
}

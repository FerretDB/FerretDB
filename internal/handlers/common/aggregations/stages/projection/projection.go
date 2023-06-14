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

// Package projection provides projection for aggregations.
package projection

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ValidateProjection check projection document.
// Document fields could be either included or excluded but not both.
// Exception is for the _id field that could be included or excluded.
//
// Command error codes:
//   - `ErrEmptyProject` when projection document is empty;
//   - `ErrProjectionExIn` when there is exclusion in inclusion projection;
//   - `ErrProjectionInEx` when there is inclusion in exclusion projection;
//   - `ErrEmptyFieldPath` when projection path is empty;
//   - `ErrPathContainsEmptyElement` when projection path contains empty key;
//   - `ErrFieldPathInvalidName` when `$` is at the prefix of a key in the path;
//   - `ErrWrongPositionalOperatorLocation` when there are multiple `$`;
//   - `ErrAggregatePositionalProject` when `$` is used in the suffix key;
//   - `ErrAggregatePositionalProject` when positional projection contains empty path;
//   - `ErrNotImplemented` when there is unimplemented projection operators and expressions.
func ValidateProjection(projection *types.Document) (*types.Document, bool, error) {
	validated := types.MakeDocument(0)

	if projection.Len() == 0 {
		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrEmptyProject,
			"Invalid $project :: caused by :: projection specification must have at least one field",
			"$project (stage)",
		)
	}

	var projectionVal *bool

	iter := projection.Iterator()
	defer iter.Close()

	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if key == "" {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrEmptyFieldPath,
				"Invalid $project :: caused by :: FieldPath cannot be constructed with empty string",
				"$project (stage)",
			)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			if strings.HasSuffix(key, "$") {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrAggregatePositionalProject,
					"Invalid $project :: caused by :: Cannot use positional projection in aggregation projection",
					"$project (stage)",
				)
			}

			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrPathContainsEmptyElement,
				"Invalid $project :: caused by :: FieldPath field names may not be empty strings.",
				"projection",
			)
		}

		if key == "$" {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFieldPathInvalidName,
				"Invalid $project :: caused by :: FieldPath field names may not start with '$'. "+
					"Consider using $getField or $setField.",
				"$project (stage)",
			)
		}

		if strings.HasSuffix(key, "$") {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrAggregatePositionalProject,
				"Invalid $project :: caused by :: Cannot use positional projection in aggregation projection",
				"$project (stage)",
			)
		}

		// if `$` is at prefix, it returns ErrFieldPathInvalidName error code instead.
		prefixTrimmed := strings.TrimPrefix(key, "$")
		if slices.Contains(strings.Split(prefixTrimmed, "."), "$") {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrWrongPositionalOperatorLocation,
				"Invalid $project :: caused by :: "+
					"Positional projection may only be used at the end, "+
					"for example: a.b.$. If the query previously used a form "+
					"like a.b.$.d, remove the parts following the '$' and "+
					"the results will be equivalent.",
				"$project (stage)",
			)
		}

		for _, k := range path.Slice() {
			if strings.HasPrefix(k, "$") {
				// arbitrary `$` cannot exist in the path
				// `v.$foo` is invalid, `v.$` and `v.foo$` are fine.
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFieldPathInvalidName,
					"Invalid $project :: caused by :: FieldPath field names may not start with '$'. "+
						"Consider using $getField or $setField.",
					"$project (stage)",
				)
			}
		}

		var result bool

		switch value := value.(type) {
		case *types.Document:
			// validate operators later
			validated.Set(key, value)
			result = true

		case *types.Array, string, types.Binary, types.ObjectID,
			time.Time, types.NullType, types.Regex, types.Timestamp: // all this types are treated as new fields value
			result = true

			validated.Set(key, value)
		case float64, int32, int64:
			// projection treats 0 as false and any other value as true
			comparison := types.Compare(value, int32(0))

			if comparison != types.Equal {
				result = true
			}

			// set the value with boolean result to omit type assertion when we will apply projection
			validated.Set(key, result)
		case bool:
			result = value

			// set the value with boolean result to omit type assertion when we will apply projection
			validated.Set(key, result)
		default:
			return nil, false, lazyerrors.Errorf("unsupported operation %s %value (%T)", key, value, value)
		}

		if projection.Len() == 1 && key == "_id" {
			return validated, result, nil
		}

		// if projectionVal is nil we are processing the first field
		if projectionVal == nil {
			if key == "_id" {
				continue
			}

			projectionVal = &result

			continue
		}

		if *projectionVal != result {
			if *projectionVal {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrProjectionExIn,
					fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", key),
					"projection",
				)
			}

			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrProjectionInEx,
				fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
				"projection",
			)
		}
	}

	return validated, *projectionVal, nil
}

// ProjectDocument applies projection to the copy of the document.
func ProjectDocument(doc, projection *types.Document, inclusion bool) (*types.Document, error) {
	projected, err := types.NewDocument("_id", must.NotFail(doc.Get("_id")))
	if err != nil {
		return nil, err
	}

	if projection.Has("_id") {
		idValue := must.NotFail(projection.Get("_id"))

		var set bool

		switch idValue := idValue.(type) {
		case *types.Document: // field: { $elemMatch: { field2: value }}
			var op operators.Operator
			var value any

			op, err = operators.NewOperator(idValue)
			if err != nil {
				return nil, processOperatorError(err)
			}

			value, err = op.Process(projected)
			if err != nil {
				return nil, err
			}

			projected.Set("_id", value)

		case *types.Array, string, types.Binary, types.ObjectID,
			time.Time, types.NullType, types.Regex, types.Timestamp: // all this types are treated as new fields value
			projected.Set("_id", idValue)

			set = true
		case bool:
			set = idValue

		default:
			return nil, lazyerrors.Errorf("unsupported operation %s %v (%T)", "_id", idValue, idValue)
		}

		if !set {
			projected.Remove("_id")
		}
	}

	projectedWithoutID, err := projectDocumentWithoutID(doc, projection, inclusion)
	if err != nil {
		// TODO: https://github.com/FerretDB/FerretDB/issues/2633
		return nil, err
	}

	for _, key := range projectedWithoutID.Keys() {
		projected.Set(key, must.NotFail(projectedWithoutID.Get(key)))
	}

	return projected, nil
}

// projectDocumentWithoutID applies projection to the copy of the document and returns projected document.
// It ignores _id field in the projection.
func projectDocumentWithoutID(doc *types.Document, projection *types.Document, inclusion bool) (*types.Document, error) {
	projectionWithoutID := projection.DeepCopy()
	projectionWithoutID.Remove("_id")

	docWithoutID := doc.DeepCopy()
	docWithoutID.Remove("_id")

	projected := types.MakeDocument(0)

	if !inclusion {
		projected = docWithoutID.DeepCopy()
	}

	iter := projectionWithoutID.Iterator()
	defer iter.Close()

	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch value := value.(type) { // found in the projection
		case *types.Document: // field: { $elemMatch: { field2: value }}
			var op operators.Operator
			var v any

			op, err = operators.NewOperator(value)
			if err != nil {
				return nil, processOperatorError(err)
			}

			v, err = op.Process(doc)
			if err != nil {
				return nil, err
			}

			projected.Set(key, v)

		case *types.Array, string, types.Binary, types.ObjectID,
			time.Time, types.NullType, types.Regex, types.Timestamp: // all these types are treated as new fields value
			projected.Set(key, value)

		case bool: // field: bool
			if inclusion {
				// inclusion projection copies the field on the path from docWithoutID to projected.
				if _, err = includeProjection(path, docWithoutID, projected); err != nil {
					return nil, err
				}

				continue
			}

			// exclusion projection removes the field on the path in projected.
			excludeProjection(path, projected)
		default:
			return nil, lazyerrors.Errorf("unsupported operation %s %v (%T)", key, value, value)
		}
	}

	return projected, nil
}

// includeProjection copies the field on the path from source to projected.
// When an array is on the path, it returns the array containing any document
// with the same key. Dot notation with array index path does not include
// the field unlike document.SetByPath(path).
// Inclusion projection with non-existent path creates an empty document
// or an empty array based on what source has.
// It returns iterator errors other than ErrIteratorDone.
// If the projected contains field that is not expected in source, it panics.
//
//	Example: "v.foo" path inclusion projection:
//	{v: {foo: 1, bar: 1}}               -> {v: {foo: 1}}
//	{v: {bar: 1}}                       -> {v: {}}
//	{v: [{bar: 1}]}                     -> {v: [{}]}
//	{v: [{foo: 1}, {foo: 2}, {bar: 1}]} -> {v: [{foo: 1}, {foo: 2}, {}]}
//
//	Example: "v.0.foo" path inclusion projection:
//	{v: [{foo: 1}, {foo: 2}, {bar: 1}]} -> {v: [{}, {}, {}]}
func includeProjection(path types.Path, source any, projected *types.Document) (*types.Array, error) {
	key := path.Prefix()

	switch source := source.(type) {
	case *types.Document:
		embeddedSource, err := source.Get(key)
		if err != nil {
			// key does not exist, nothing to set.
			return nil, nil
		}

		if path.Len() <= 1 {
			// path reached suffix, set field in projected.
			setBySourceOrder(key, embeddedSource, source, projected)
			return nil, nil
		}

		doc := new(types.Document)

		if projected.Has(key) {
			// set doc if projected has field from other projection field.
			v := must.NotFail(projected.Get(key))
			if d, ok := v.(*types.Document); ok {
				doc = d
			}

			if arr, ok := v.(*types.Array); ok {
				// use next prefix key with arr value, allowing array to parse existing
				// projection fields.
				doc = must.NotFail(types.NewDocument(path.TrimPrefix().Prefix(), arr))
			}
		}

		// when next prefix has an array use returned value arr,
		// if it has a document, field in the doc is set by includeProjection.
		arr, err := includeProjection(path.TrimPrefix(), embeddedSource, doc)
		if err != nil {
			return nil, err
		}

		switch embeddedSource.(type) {
		case *types.Document:
			setBySourceOrder(key, doc, source, projected)
		case *types.Array:
			projected.Set(key, arr)
		}

		return nil, nil
	case *types.Array:
		iter := source.Iterator()
		defer iter.Close()

		arr := new(types.Array)
		var inclusionExists bool

		if v, err := projected.Get(key); err == nil {
			projectedArr, ok := v.(*types.Array)
			if ok {
				arr = projectedArr
				inclusionExists = true
			}
		}

		i := 0

		for {
			_, arrElem, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return nil, lazyerrors.Error(err)
			}

			if _, ok := arrElem.(*types.Document); !ok {
				continue
			}

			doc := new(types.Document)

			if inclusionExists {
				// when there are multiple inclusion fields, first inclusion
				// inserts all documents from source to arr, they could be empty
				// if it did not match previous inclusion fields.
				// But number of documents in arr must be the same as number of documents
				// in source.
				var v any

				v, err = arr.Get(i)
				if err != nil {
					panic(err)
				}

				docVal, ok := v.(*types.Document)
				if !ok {
					panic("projected field must be a document")
				}

				doc = docVal
			} else {
				// first inclusion field, insert it to the doc.
				arr.Append(doc)
			}

			if _, err = includeProjection(path, arrElem, doc); err != nil {
				return nil, err
			}

			arr.Set(i, doc)
			i++
		}

		return arr, nil
	default:
		// field is not a document or an array, nothing to set.
		return nil, nil
	}
}

// excludeProjection removes the field on the path in projected.
// When an array is on the path, it checks if the array contains any document
// with the key to remove that document. This is not the case in document.Remove(key).
// Dot notation with array index path do not exclude unlike document.RemoveByPath(key).
//
//	Examples: "v.foo" path exclusion projection:
//	{v: {foo: 1}}                       -> {v: {}}
//	{v: {foo: 1, bar: 1}}               -> {v: {bar: 1}}
//	{v: [{foo: 1}, {foo: 2}]}           -> {v: [{}, {}]}
//	{v: [{foo: 1}, {foo: 2}, {bar: 1}]} -> {v: [{}, {}, {bar: 1}]}
//
//	Example: "v.0.foo" path exclusion projection:
//	{v: [{foo: 1}, {foo: 2}]}           -> {v: [{foo: 1}, {foo: 2}]}
func excludeProjection(path types.Path, projected any) {
	key := path.Prefix()

	switch projected := projected.(type) {
	case *types.Document:
		embeddedSource, err := projected.Get(key)
		if err != nil {
			// key does not exist, nothing to exclude.
			return
		}

		if path.Len() <= 1 {
			// path reached suffix, remove the field from the document.
			projected.Remove(key)
			return
		}

		// recursively remove field from the embeddedSource.
		excludeProjection(path.TrimPrefix(), embeddedSource)

		return
	case *types.Array:
		// modifies the field of projected, hence not using iterator.
		for i := 0; i < projected.Len(); i++ {
			arrElem := must.NotFail(projected.Get(i))

			if _, ok := arrElem.(*types.Document); !ok {
				// not a document, cannot possibly be part of path, do nothing.
				continue
			}

			excludeProjection(path, arrElem)
		}

		return
	default:
		// not a path, nothing to exclude.
		return
	}
}

// setBySourceOrder sets the key value field to projected in same field order as the source.
// Example:
//
//	key: foo
//	val: 1
//	source: {foo: 1, bar: 2}
//	projected: {bar: 2}
//
// setBySourceOrder sets projected to {foo: 1, bar: 2} rather than adding it to the last field.
func setBySourceOrder(key string, val any, source, projected *types.Document) {
	projectedKeys := projected.Keys()

	// newFieldIndex is where new field is to be inserted in projected document.
	newFieldIndex := 0

	for _, sourceKey := range source.Keys() {
		if sourceKey == key {
			break
		}

		if newFieldIndex >= len(projectedKeys) {
			break
		}

		if sourceKey == projectedKeys[newFieldIndex] {
			newFieldIndex++
		}
	}

	tmp := projected.DeepCopy()

	// remove fields of projected from newFieldIndex to the end
	for i := newFieldIndex; i < len(projectedKeys); i++ {
		projected.Remove(projectedKeys[i])
	}

	projected.Set(key, val)

	// copy newFieldIndex-th to the end from tmp to projected
	i := newFieldIndex
	for _, key := range tmp.Keys()[newFieldIndex:] {
		projected.Set(key, must.NotFail(tmp.Get(tmp.Keys()[i])))
		i++
	}
}

// processOperatorError takes internal error related to operator evaluation and
// returns proper CommandError that can be returned by $process aggregation stage.
func processOperatorError(err error) error {
	if err == nil {
		return nil
	}

	var opErr operators.OperatorError
	if !errors.As(err, &opErr) {
		return lazyerrors.Error(err)
	}

	switch opErr.Code() {
	case operators.ErrEmptyField:
		return commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrEmptySubProject,
			"Invalid $project :: caused by :: An empty sub-projection is not a valid value."+
				" Found empty object at path",
			"$project (stage)",
		)
	case operators.ErrTooManyFields:
		return commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFieldPathInvalidName,
			"Invalid $project :: caused by :: FieldPath field names may not start with '$'."+
				" Consider using $getField or $setField.",
			"$project (stage)",
		)
	case operators.ErrNotImplemented:
		return commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrEmptySubProject,
			"Invalid $project :: caused by :: An empty sub-projection is not a valid value."+
				" Found empty object at path",
			"$project (stage)",
		)
	case operators.ErrArgsInvalidLen:
		return commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrOperatorWrongLenOfArgs,
			"Invalid $project :: caused by :: "+opErr.Error(),
			"$project (stage)",
		)
	case operators.ErrInvalidExpression:
		return commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidPipelineOperator,
			"Invalid $project :: caused by :: "+opErr.Error(),
			"$project (stage)",
		)
	case operators.ErrNoOperator, operators.ErrWrongType:
		fallthrough
	default:
		return lazyerrors.Error(err)
	}
}

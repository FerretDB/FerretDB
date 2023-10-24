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
	"slices"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/commonpath"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FilterDocument returns true if given document satisfies given filter expression.
//
// Passed arguments must not be modified.
func FilterDocument(doc, filter *types.Document) (bool, error) {
	iter := filter.Iterator()
	defer iter.Close()

	for {
		filterKey, filterValue, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return true, nil
			}

			return false, lazyerrors.Error(err)
		}

		// top-level filters are ANDed together
		matches, err := filterDocumentPair(doc, filterKey, filterValue)
		if err != nil {
			return false, lazyerrors.Error(err)
		}
		if !matches {
			return false, nil
		}
	}
}

// HasQueryOperator recursively checks if filter document contains any operator prefixed with $.
func HasQueryOperator(filter *types.Document) (bool, error) {
	iter := filter.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return false, nil
			}

			return false, lazyerrors.Error(err)
		}

		if strings.HasPrefix(k, "$") {
			return true, nil
		}

		doc, ok := v.(*types.Document)
		if !ok {
			continue
		}

		hasOperator, err := HasQueryOperator(doc)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		if hasOperator {
			return true, nil
		}
	}
}

// filterDocumentPair handles a single filter element key/value pair {filterKey: filterValue}.
func filterDocumentPair(doc *types.Document, filterKey string, filterValue any) (bool, error) {
	var vals []any
	filterSuffix := filterKey

	if strings.ContainsRune(filterKey, '.') {
		path, err := types.NewPathFromString(filterKey)
		if err != nil {
			return false, lazyerrors.Error(err)
		}

		filterSuffix = path.Suffix()

		// filter using dot notation returns the value by valid array index
		// or values for the given key in array's document
		if vals, err = commonpath.FindValues(doc, path, &commonpath.FindValuesOpts{
			FindArrayIndex:     true,
			FindArrayDocuments: true,
		}); err != nil {
			return false, lazyerrors.Error(err)
		}
	} else {
		if val, _ := doc.Get(filterKey); val != nil {
			vals = []any{val}
		}
	}

	if strings.HasPrefix(filterKey, "$") {
		// {$operator: filterValue}
		return filterOperator(doc, filterKey, filterValue)
	}

	switch filterValue := filterValue.(type) {
	case *types.Document:
		var docs []*types.Document
		for _, val := range vals {
			docs = append(docs, must.NotFail(types.NewDocument(filterSuffix, val)))
		}

		if len(docs) == 0 {
			// operators like $nin uses empty document to filter non-existent field
			docs = append(docs, types.MakeDocument(0))
		}

		for _, doc := range docs {
			// {field: {expr}} or {field: {document}}
			ok, err := filterFieldExpr(doc, filterKey, filterSuffix, filterValue)
			if err != nil {
				return false, err
			}

			if ok {
				return true, nil
			}
		}
	case types.NullType:
		if len(vals) == 0 {
			// comparing non-existent field with null returns true
			return true, nil
		}

		for _, val := range vals {
			if result := types.Compare(val, filterValue); result == types.Equal {
				return true, nil
			}
		}
	case types.Regex:
		for _, val := range vals {
			ok, err := filterFieldRegex(val, filterValue)
			if err != nil {
				return false, err
			}

			if ok {
				return true, nil
			}
		}
	default:
		for _, val := range vals {
			if result := types.Compare(val, filterValue); result == types.Equal {
				return true, nil
			}
		}
	}

	// If we got here, it means that none of the documents matched the filter.
	return false, nil
}

// filterOperator handles a top-level operator filter {$operator: filterValue}.
func filterOperator(doc *types.Document, operator string, filterValue any) (bool, error) {
	switch operator {
	case "$and":
		// {$and: [{expr1}, {expr2}, ...]}
		exprs, ok := filterValue.(*types.Array)
		if !ok {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$and must be an array",
				operator,
			)
		}

		if exprs.Len() == 0 {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$and/$or/$nor must be a nonempty array",
				operator,
			)
		}

		for i := 0; i < exprs.Len(); i++ {
			_, ok := must.NotFail(exprs.Get(i)).(*types.Document)
			if !ok {
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"$or/$and/$nor entries need to be full objects",
					operator,
				)
			}
		}

		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)

			matches, err := FilterDocument(doc, expr)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}

		return true, nil

	case "$or":
		// {$or: [{expr1}, {expr2}, ...]}
		exprs, ok := filterValue.(*types.Array)
		if !ok {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$or must be an array",
				operator,
			)
		}

		if exprs.Len() == 0 {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$and/$or/$nor must be a nonempty array",
				operator,
			)
		}

		for i := 0; i < exprs.Len(); i++ {
			_, ok := must.NotFail(exprs.Get(i)).(*types.Document)
			if !ok {
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"$or/$and/$nor entries need to be full objects",
					operator,
				)
			}
		}

		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)

			matches, err := FilterDocument(doc, expr)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}

		return false, nil

	case "$nor":
		// {$nor: [{expr1}, {expr2}, ...]}
		exprs, ok := filterValue.(*types.Array)
		if !ok {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$nor must be an array",
				operator,
			)
		}

		if exprs.Len() == 0 {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$and/$or/$nor must be a nonempty array",
				operator,
			)
		}

		for i := 0; i < exprs.Len(); i++ {
			_, ok := must.NotFail(exprs.Get(i)).(*types.Document)
			if !ok {
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"$or/$and/$nor entries need to be full objects",
					operator,
				)
			}
		}

		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)

			matches, err := FilterDocument(doc, expr)
			if err != nil {
				return false, err
			}
			if matches {
				return false, nil
			}
		}

		return true, nil

	case "$comment":
		return true, nil

	case "$expr":
		return filterExprOperator(doc, must.NotFail(types.NewDocument(operator, filterValue)))
	default:
		msg := fmt.Sprintf(
			`unknown top level operator: %s. `+
				`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			operator,
		)

		return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, msg, "$operator")
	}
}

// filterExprOperator uses $expr operator to allow usage of aggregation expression.
// It returns boolean indicating filter has matched.
//
// $expr is primary used by operators such as $gt and $cond which return boolean result.
// However, if non-boolean result is returned from processing aggregation expression,
// it returns false for null or zero value and true for all other values.
func filterExprOperator(doc, filter *types.Document) (bool, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3170
	op, err := operators.NewExpr(filter, "$expr")
	if err != nil {
		return false, err
	}

	v, err := op.Process(doc)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	switch v := v.(type) {
	case *types.Document, *types.Array, string, types.Binary, types.ObjectID, time.Time, types.Regex, types.Timestamp:
		return true, nil
	case float64, int32, int64:
		return types.Compare(v, int32(0)) != types.Equal, nil
	case bool:
		return v, nil
	case types.NullType:
		return false, nil
	default:
		panic(fmt.Sprintf("common.filterExprOperator: unexpected type %[1]T (%#[1]v)", v))
	}
}

// filterFieldExpr handles {field: {expr}} or {field: {document}} filter.
func filterFieldExpr(doc *types.Document, filterKey, filterSuffix string, expr *types.Document) (bool, error) {
	// check if both documents are empty
	if expr.Len() == 0 {
		fieldValue, err := doc.Get(filterSuffix)
		if err != nil {
			return false, nil
		}
		if fieldValue, ok := fieldValue.(*types.Document); ok && fieldValue.Len() == 0 {
			return true, nil
		}
		return false, nil
	}

	for _, exprKey := range expr.Keys() {
		if exprKey == "$options" {
			// handled by $regex
			continue
		}

		exprValue := must.NotFail(expr.Get(exprKey))

		fieldValue, err := doc.Get(filterSuffix)
		if err != nil {
			switch exprKey {
			case "$exists", "$not", "$elemMatch":
			case "$type":
				if v, ok := exprValue.(string); ok && v == "null" {
					// null and unset are different for $type operator.
					return false, nil
				}
			default:
				// Set non-existent field to null for the operator
				// to compute result. The comparison treats non-existent
				// field on documents as equivalent.
				fieldValue = types.Null
			}
		}

		if !strings.HasPrefix(exprKey, "$") {
			if documentValue, ok := fieldValue.(*types.Document); ok {
				result := types.Compare(documentValue, expr)
				return result == types.Equal, nil
			}
			return false, nil
		}

		switch exprKey {
		case "$eq":
			// {field: {$eq: exprValue}}
			switch exprValue := exprValue.(type) {
			case *types.Document:
				if fieldValue, ok := fieldValue.(*types.Document); ok {
					result := types.Compare(exprValue, fieldValue)
					return result == types.Equal, nil
				}
				return false, nil
			default:
				result := types.Compare(fieldValue, exprValue)
				if result != types.Equal {
					return false, nil
				}
			}

		case "$ne":
			// {field: {$ne: exprValue}}
			switch exprValue := exprValue.(type) {
			case *types.Document:
				if fieldValue, ok := fieldValue.(*types.Document); ok {
					result := types.Compare(exprValue, fieldValue)
					return result != types.Equal, nil
				}

				return true, nil
			case types.Regex:
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"Can't have regex as arg to $ne.",
					exprKey,
				)
			default:
				result := types.Compare(fieldValue, exprValue)
				if result == types.Equal {
					return false, nil
				}
			}

		case "$gt":
			// {field: {$gt: exprValue}}
			if _, ok := exprValue.(types.Regex); ok {
				msg := fmt.Sprintf(`Can't have RegEx as arg to predicate over field '%s'.`, filterKey)
				return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, msg, exprKey)
			}

			// Array and non-array comparison with $gt compares the non-array
			// value against the maximum value of the same BSON type value of the array.
			// Filter the array by only keeping the same type as the non-array value,
			// then find the maximum value from the array.
			// If array does not contain same BSON type, returns false.
			// All numbers are treated as the same type.
			// Example:
			// expr {v: {$gt: 42}}
			// value [{v: 40}, {v: 41.5}, {v: "foo"}, {v: nil}]
			// Above compares the maximum number of array 41.5 to the filter 42,
			// and results in Less. Other values "foo" and nil which are
			// not number type are not considered for $gt comparison.

			result := types.CompareOrderForOperator(fieldValue, exprValue, types.Descending)
			if result != types.Greater {
				return false, nil
			}

		case "$gte":
			// {field: {$gte: exprValue}}
			if _, ok := exprValue.(types.Regex); ok {
				msg := fmt.Sprintf(`Can't have RegEx as arg to predicate over field '%s'.`, filterKey)
				return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, msg, exprKey)
			}

			// Array and non-array comparison with $gte compares the non-array
			// value against the maximum value of the same BSON type value of the array.
			// Filter the array by only keeping the same type as the non-array value,
			// then find the maximum value from the array.
			// If array does not contain same BSON type, returns false.
			// All numbers are treated as the same type.
			// Example:
			// expr {v: {$gte: 42}}
			// value [{v: 40}, {v: 41.5}, {v: "foo"}, {v: nil}]
			// Above compares the maximum number of array 41.5 to the filter 42,
			// and results in Less. Other values "foo" and nil which are
			// not number type are not considered for $gte comparison.
			result := types.CompareOrderForOperator(fieldValue, exprValue, types.Descending)
			if result != types.Equal && result != types.Greater {
				return false, nil
			}

		case "$lt":
			// {field: {$lt: exprValue}}
			if _, ok := exprValue.(types.Regex); ok {
				msg := fmt.Sprintf(`Can't have RegEx as arg to predicate over field '%s'.`, filterKey)
				return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, msg, exprKey)
			}

			// Array and non-array comparison with $lt compares the non-array
			// value against the minimum value of the same BSON type value of the array.
			// Filter the array by only keeping the same type as the non-array value,
			// then find the minimum value from the array.
			// If array does not contain same BSON type, returns false.
			// All numbers are treated as the same type.
			// Example:
			// expr {v: {$gte: 42}}
			// value [{v: 40}, {v: 41.5}, {v: "foo"}, {v: nil}]
			// Above compares the minimum number of array 40 to the filter 42,
			// and results in Less. Other values "foo" and nil which are
			// not number type are not considered for $lt comparison.

			result := types.CompareOrderForOperator(fieldValue, exprValue, types.Ascending)
			if result != types.Less {
				return false, nil
			}

		case "$lte":
			// {field: {$lte: exprValue}}
			if _, ok := exprValue.(types.Regex); ok {
				msg := fmt.Sprintf(`Can't have RegEx as arg to predicate over field '%s'.`, filterKey)
				return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, msg, exprKey)
			}

			// Array and non-array comparison with $lte compares the non-array
			// value against the minimum value of the same BSON type value of the array.
			// Filter the array by only keeping the same type as the non-array value,
			// then find the minimum value from the array.
			// If array does not contain same BSON type, returns false.
			// All numbers are treated as the same type.
			// Example:
			// expr {v: {$gte: 42}}
			// value [{v: 40}, {v: 41.5}, {v: "foo"}, {v: nil}]
			// Above compares the minimum number of array 40 to the filter 42,
			// and results in Less. Other values "foo" and nil which are
			// not number type are not considered for $lt comparison.

			result := types.CompareOrderForOperator(fieldValue, exprValue, types.Ascending)
			if result != types.Equal && result != types.Less {
				return false, nil
			}

		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			arr, ok := exprValue.(*types.Array)
			if !ok {
				return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, "$in needs an array", exprKey)
			}

			var found bool
			for i := 0; i < arr.Len(); i++ {
				if found {
					break
				}

				switch arrValue := must.NotFail(arr.Get(i)).(type) {
				case *types.Document:
					for _, key := range arrValue.Keys() {
						if strings.HasPrefix(key, "$") {
							return false, commonerrors.NewCommandErrorMsgWithArgument(
								commonerrors.ErrBadValue,
								"cannot nest $ under $in",
								exprKey,
							)
						}
					}

					if fieldValue, ok := fieldValue.(*types.Document); ok {
						if result := types.Compare(fieldValue, arrValue); result == types.Equal {
							found = true
						}
					}
				case types.Regex:
					match, err := filterFieldRegex(fieldValue, arrValue)
					switch {
					case err != nil:
						return false, err
					case match:
						found = true
					}
				default:
					result := types.Compare(fieldValue, arrValue)
					if result == types.Equal {
						found = true
					}
				}
			}

			if !found {
				return false, nil
			}

		case "$nin":
			// {field: {$nin: [value1, value2, ...]}}
			arr, ok := exprValue.(*types.Array)
			if !ok {
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"$nin needs an array",
					exprKey,
				)
			}

			var found bool
			for i := 0; i < arr.Len(); i++ {
				if found {
					break
				}

				switch arrValue := must.NotFail(arr.Get(i)).(type) {
				case *types.Document:
					for _, key := range arrValue.Keys() {
						if strings.HasPrefix(key, "$") {
							return false, commonerrors.NewCommandErrorMsgWithArgument(
								commonerrors.ErrBadValue,
								"cannot nest $ under $in",
								exprKey,
							)
						}
					}

					if fieldValue, ok := fieldValue.(*types.Document); ok {
						if result := types.Compare(fieldValue, arrValue); result == types.Equal {
							found = true
						}
					}
				case types.Regex:
					match, err := filterFieldRegex(fieldValue, arrValue)
					switch {
					case err != nil:
						return false, err
					case match:
						found = true
					}
				default:
					result := types.Compare(fieldValue, arrValue)
					if result == types.Equal {
						found = true
					}
				}
			}

			if found {
				return false, nil
			}

		case "$not":
			// {field: {$not: {expr}}}
			switch exprValue := exprValue.(type) {
			case *types.Document:
				res, err := filterFieldExpr(doc, filterKey, filterSuffix, exprValue)
				if res || err != nil {
					return false, err
				}
			case types.Regex:
				optionsAny, _ := expr.Get("$options")
				res, err := filterFieldExprRegex(fieldValue, exprValue, optionsAny)
				if res || err != nil {
					return false, err
				}
			default:
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"$not needs a regex or a document",
					exprKey,
				)
			}

		case "$regex":
			// {field: {$regex: exprValue}}
			optionsAny, _ := expr.Get("$options")
			res, err := filterFieldExprRegex(fieldValue, exprValue, optionsAny)
			if !res || err != nil {
				return false, err
			}

		case "$elemMatch":
			// {field: {$elemMatch: value}}
			res, err := filterFieldExprElemMatch(doc, filterKey, filterSuffix, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$size":
			// {field: {$size: value}}
			res, err := filterFieldExprSize(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$all":
			// {field: {$all: [value, another_value, ...]}}
			res, err := filterFieldExprAll(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$bitsAllClear":
			// {field: {$bitsAllClear: value}}
			res, err := filterFieldExprBitsAllClear(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$bitsAllSet":
			// {field: {$bitsAllSet: value}}
			res, err := filterFieldExprBitsAllSet(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$bitsAnyClear":
			// {field: {$bitsAnyClear: value}}
			res, err := filterFieldExprBitsAnyClear(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$bitsAnySet":
			// {field: {$bitsAnySet: value}}
			res, err := filterFieldExprBitsAnySet(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$mod":
			// {field: {$mod: [divisor, remainder]}}
			res, err := filterFieldMod(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$exists":
			// {field: {$exists: value}}
			res, err := filterFieldExprExists(fieldValue != nil, exprValue)
			if !res || err != nil {
				return false, err
			}

		case "$type":
			// {field: {$type: value}}
			res, err := filterFieldExprType(fieldValue, exprValue)
			if !res || err != nil {
				return false, err
			}

		default:
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("unknown operator: %s", exprKey),
				"$operator",
			)
		}
	}

	return true, nil
}

// filterFieldRegex handles {field: /regex/} filter. Provides regular expression capabilities
// for pattern matching strings in queries, even if the strings are in an array.
func filterFieldRegex(fieldValue any, regex types.Regex) (bool, error) {
	for _, option := range regex.Options {
		if !slices.Contains([]rune{'i', 'm', 's', 'x'}, option) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadRegexOption,
				fmt.Sprintf(" invalid flag in regex options: %c", option),
				"$options",
			)
		}
	}

	re, err := regex.Compile()
	if err != nil && err == types.ErrOptionNotImplemented {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			`option 'x' not implemented`,
			"$options",
		)
	}
	if err != nil {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrRegexMissingParen,
			err.Error(),
			"$regex",
		)
	}

	switch fieldValue := fieldValue.(type) {
	case *types.Array:
		for i := 0; i < fieldValue.Len(); i++ {
			arrValue := must.NotFail(fieldValue.Get(i))
			s, isString := arrValue.(string)
			if !isString {
				continue
			}
			if re.MatchString(s) {
				return true, nil
			}
		}

	case string:
		return re.MatchString(fieldValue), nil

	case types.Regex:
		result := types.Compare(fieldValue, regex)
		return result == types.Equal, nil
	}

	return false, nil
}

// filterFieldExprRegex handles {field: {$regex: regexValue, $options: optionsValue}} filter.
func filterFieldExprRegex(fieldValue any, regexValue, optionsValue any) (bool, error) {
	var options string
	if optionsValue != nil {
		var ok bool
		if options, ok = optionsValue.(string); !ok {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"$options has to be a string",
				"$options",
			)
		}
	}

	switch regexValue := regexValue.(type) {
	case string:
		regex := types.Regex{
			Pattern: regexValue,
			Options: options,
		}
		return filterFieldRegex(fieldValue, regex)

	case types.Regex:
		if options != "" {
			if regexValue.Options != "" {
				return false, commonerrors.NewCommandErrorMsg(
					commonerrors.ErrRegexOptions,
					"options set in both $regex and $options",
				)
			}
			regexValue.Options = options
		}
		return filterFieldRegex(fieldValue, regexValue)

	default:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"$regex has to be a string",
			"$regex",
		)
	}
}

// filterFieldExprSize handles {field: {$size: sizeValue}} filter.
func filterFieldExprSize(fieldValue any, sizeValue any) (bool, error) {
	size, err := commonparams.GetWholeNumberParam(sizeValue)
	if err != nil {
		switch err {
		case commonparams.ErrUnexpectedType:
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(`Failed to parse $size. Expected a number in: $size: %s`, types.FormatAnyValue(sizeValue)),
				"$size",
			)
		case commonparams.ErrNotWholeNumber:
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(`Failed to parse $size. Expected an integer: $size: %s`, types.FormatAnyValue(sizeValue)),
				"$size",
			)
		case commonparams.ErrInfinity:
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					`Failed to parse $size. Cannot represent as a 64-bit integer: $size: %s`,
					types.FormatAnyValue(sizeValue),
				),
				"$size",
			)
		default:
			return false, err
		}
	}

	if size < 0 {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(
				`Failed to parse $size. Expected a non-negative number in: $size: %s`,
				types.FormatAnyValue(sizeValue),
			),
			"$size",
		)
	}

	arr, ok := fieldValue.(*types.Array)
	if !ok {
		return false, nil
	}

	if arr.Len() != int(size) {
		return false, nil
	}

	return true, nil
}

// filterFieldExprAll handles {field: {$all: [value, another_value, ...]}} filter.
// The main purpose of $all is to filter arrays.
// It is possible to filter non-arrays: {field: {$all: [value]}}, but such statement is equivalent to {field: value}.
func filterFieldExprAll(fieldValue any, allValue any) (bool, error) {
	query, ok := allValue.(*types.Array)
	if !ok {
		return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, "$all needs an array", "$all")
	}

	if query.Len() == 0 {
		return false, nil
	}

	switch value := fieldValue.(type) {
	case *types.Document:
		// For documents we return false as $all doesn't work on documents.
		return false, nil

	case *types.Array:
		// For arrays we check that the array contains all the elements of the query.
		return value.ContainsAll(query), nil

	default:
		// For other types (scalars) we check that the value is equal to each scalar in the query.
		// Example: value: 42, query: [42, 42] should give us `true`
		for i := 0; i < query.Len(); i++ {
			res := types.Compare(value, must.NotFail(query.Get(i)))
			if res != types.Equal {
				return false, nil
			}
		}
		return true, nil
	}
}

// filterFieldExprBitsAllClear handles {field: {$bitsAllClear: value}} filter.
func filterFieldExprBitsAllClear(fieldValue, maskValue any) (bool, error) {
	bitmask, err := getBinaryMaskParam("$bitsAllClear", maskValue)
	if err != nil {
		return false, err
	}

	switch value := fieldValue.(type) {
	case float64:
		if isInvalidBitwiseValue(value) {
			return false, nil
		}

		return (^uint64(value) & bitmask) == bitmask, nil

	case types.Binary:
		// TODO https://github.com/FerretDB/FerretDB/issues/508
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"BinData() not supported yet",
			"$bitsAllClear",
		)

	case int32:
		return (^uint64(value) & bitmask) == bitmask, nil

	case int64:
		return (^uint64(value) & bitmask) == bitmask, nil

	default:
		return false, nil
	}
}

// filterFieldExprBitsAllSet handles {field: {$bitsAllSet: value}} filter.
func filterFieldExprBitsAllSet(fieldValue, maskValue any) (bool, error) {
	bitmask, err := getBinaryMaskParam("$bitsAllSet", maskValue)
	if err != nil {
		return false, err
	}

	switch value := fieldValue.(type) {
	case float64:
		if isInvalidBitwiseValue(value) {
			return false, nil
		}

		return (uint64(value) & bitmask) == bitmask, nil

	case types.Binary:
		// TODO https://github.com/FerretDB/FerretDB/issues/508
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"BinData() not supported yet",
			"$bitsAllSet",
		)

	case int32:
		return (uint64(value) & bitmask) == bitmask, nil

	case int64:
		return (uint64(value) & bitmask) == bitmask, nil

	default:
		return false, nil
	}
}

// filterFieldExprBitsAnyClear handles {field: {$bitsAnyClear: value}} filter.
func filterFieldExprBitsAnyClear(fieldValue, maskValue any) (bool, error) {
	bitmask, err := getBinaryMaskParam("$bitsAnyClear", maskValue)
	if err != nil {
		return false, err
	}

	switch value := fieldValue.(type) {
	case float64:
		if isInvalidBitwiseValue(value) {
			return false, nil
		}

		return (^uint64(value) & bitmask) != 0, nil

	case types.Binary:
		// TODO https://github.com/FerretDB/FerretDB/issues/508
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"BinData() not supported yet",
			"$bitsAnyClear",
		)

	case int32:
		return (^uint64(value) & bitmask) != 0, nil

	case int64:
		return (^uint64(value) & bitmask) != 0, nil

	default:
		return false, nil
	}
}

// filterFieldExprBitsAnySet handles {field: {$bitsAnySet: value}} filter.
func filterFieldExprBitsAnySet(fieldValue, maskValue any) (bool, error) {
	bitmask, err := getBinaryMaskParam("$bitsAnySet", maskValue)
	if err != nil {
		return false, err
	}

	switch value := fieldValue.(type) {
	case float64:
		if isInvalidBitwiseValue(value) {
			return false, nil
		}

		return (uint64(value) & bitmask) != 0, nil

	case types.Binary:
		// TODO https://github.com/FerretDB/FerretDB/issues/508
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"BinData() not supported yet",
			"$bitsAnySet",
		)

	case int32:
		return (uint64(value) & bitmask) != 0, nil

	case int64:
		return (uint64(value) & bitmask) != 0, nil

	default:
		return false, nil
	}
}

// isInvalidBitwiseValue returns true for an invalid value of float64
// use for bitwise operation.
// Non-integer float64, Nan, Inf are unsupported.
// The value less than math.MaxInt64,
// and greater than or equal to math.MinInt64 are unsupported.
func isInvalidBitwiseValue(value float64) bool {
	return value != math.Trunc(value) ||
		math.IsNaN(value) ||
		math.IsInf(value, 0) ||
		value >= math.MaxInt64 ||
		value < math.MinInt64
}

// filterFieldMod handles {field: {$mod: [divisor, remainder]}} filter.
func filterFieldMod(fieldValue, exprValue any) (bool, error) {
	arr := exprValue.(*types.Array)
	if arr.Len() < 2 {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			`malformed mod, not enough elements`,
			"$mod",
		)
	}
	if arr.Len() > 2 {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			`malformed mod, too many elements`,
			"$mod",
		)
	}

	var field, divisor, remainder int64
	switch d := must.NotFail(arr.Get(0)).(type) {
	case float64:
		if math.IsNaN(d) || math.IsInf(d, 0) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`malformed mod, divisor value is invalid :: caused by :: `+`Unable to coerce NaN/Inf to integral type`,
				"$mod",
			)
		}

		d = math.Trunc(d)
		if d >= float64(math.MaxInt64) || d < float64(math.MinInt64) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`malformed mod, divisor value is invalid :: caused by :: `+`Out of bounds coercing to integral value`,
				"$mod",
			)
		}

		divisor = int64(d)

	case int32:
		divisor = int64(d)

	case int64:
		divisor = d

	default:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			`malformed mod, divisor not a number`,
			"$mod",
		)
	}

	switch r := must.NotFail(arr.Get(1)).(type) {
	case float64:
		if math.IsNaN(r) || math.IsInf(r, 0) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`malformed mod, remainder value is invalid :: caused by :: `+
					`Unable to coerce NaN/Inf to integral type`, "$mod",
			)
		}

		r = math.Trunc(r)

		if r >= float64(math.MaxInt64) || r < float64(math.MinInt64) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`malformed mod, remainder value is invalid :: caused by :: `+
					`Out of bounds coercing to integral value`, "$mod",
			)
		}

		remainder = int64(r)

	case int32:
		remainder = int64(r)

	case int64:
		remainder = r

	default:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			`malformed mod, remainder not a number`,
			"$mod",
		)
	}

	if divisor == 0 {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			`divisor cannot be 0`,
			"$mod",
		)
	}

	switch f := fieldValue.(type) {
	case float64:
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return false, nil
		}
		f = math.Trunc(f)
		field = int64(f)

		if f != float64(field) {
			return false, nil
		}

	case int32:
		field = int64(f)

	case int64:
		field = f

	default:
		return false, nil
	}

	f := field % divisor
	if f != remainder {
		return false, nil
	}

	return true, nil
}

// filterFieldExprExists handles {field: {$exists: value}} filter.
func filterFieldExprExists(fieldExist bool, exprValue any) (bool, error) {
	expr, ok := exprValue.(bool)
	// return all documents if filter value is not bool type
	if !ok {
		return true, nil
	}

	switch {
	case fieldExist && expr:
		return true, nil
	case !fieldExist && !expr:
		return true, nil
	default:
		return false, nil
	}
}

// filterFieldExprType handles {field: {$type: value}} filter.
func filterFieldExprType(fieldValue, exprValue any) (bool, error) {
	switch exprValue := exprValue.(type) {
	case *types.Array:
		hasSameType := commonparams.HasSameTypeElements(exprValue)

		for i := 0; i < exprValue.Len(); i++ {
			exprValue := must.NotFail(exprValue.Get(i))

			switch exprValue := exprValue.(type) {
			case float64:
				if math.IsNaN(exprValue) || math.IsInf(exprValue, 0) {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrBadValue,
						`Invalid numerical type code: `+strings.Trim(strings.ToLower(fmt.Sprintf("%v", exprValue)), "+"),
						"$type",
					)
				}
				if exprValue != math.Trunc(exprValue) {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrBadValue,
						fmt.Sprintf(`Invalid numerical type code: %v`, exprValue),
						"$type",
					)
				}

				code, err := commonparams.NewTypeCode(int32(exprValue))
				if err != nil {
					return false, err
				}

				if !hasSameType {
					continue
				}

				res, err := filterFieldValueByTypeCode(fieldValue, code)
				if err != nil {
					return false, err
				}
				if res {
					return true, nil
				}

			case string:
				code, err := commonparams.ParseTypeCode(exprValue)
				if err != nil {
					return false, err
				}
				res, err := filterFieldValueByTypeCode(fieldValue, code)
				if err != nil {
					return false, err
				}
				if res {
					return true, nil
				}
			case int32:
				code, err := commonparams.NewTypeCode(exprValue)
				if err != nil {
					return false, err
				}

				if !hasSameType {
					continue
				}

				res, err := filterFieldValueByTypeCode(fieldValue, code)
				if err != nil {
					return false, err
				}
				if res {
					return true, nil
				}
			default:
				return false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					fmt.Sprintf(`Invalid numerical type code: %s`, exprValue),
					"$type",
				)
			}
		}
		return false, nil

	case float64:
		if math.IsNaN(exprValue) || math.IsInf(exprValue, 0) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`Invalid numerical type code: `+strings.Trim(strings.ToLower(fmt.Sprintf("%v", exprValue)), "+"),
				"$type",
			)
		}
		if exprValue != math.Trunc(exprValue) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(`Invalid numerical type code: %v`, exprValue),
				"$type",
			)
		}

		code, err := commonparams.NewTypeCode(int32(exprValue))
		if err != nil {
			return false, err
		}

		return filterFieldValueByTypeCode(fieldValue, code)

	case string:
		code, err := commonparams.ParseTypeCode(exprValue)
		if err != nil {
			return false, err
		}

		return filterFieldValueByTypeCode(fieldValue, code)

	case int32:
		code, err := commonparams.NewTypeCode(exprValue)
		if err != nil {
			return false, err
		}

		return filterFieldValueByTypeCode(fieldValue, code)

	default:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(`Invalid numerical type code: %v`, exprValue),
			"$type",
		)
	}
}

// filterFieldValueByTypeCode filters fieldValue by given type code.
func filterFieldValueByTypeCode(fieldValue any, code commonparams.TypeCode) (bool, error) {
	// check types.Array elements for match to given code.
	if array, ok := fieldValue.(*types.Array); ok && code != commonparams.TypeCodeArray {
		for i := 0; i < array.Len(); i++ {
			value, _ := array.Get(i)

			// Skip embedded arrays.
			if _, ok := value.(*types.Array); ok {
				continue
			}

			res, err := filterFieldValueByTypeCode(value, code)
			if err != nil {
				return false, err
			}

			if res {
				return true, nil
			}
		}
	}

	switch code {
	case commonparams.TypeCodeArray:
		if _, ok := fieldValue.(*types.Array); !ok {
			return false, nil
		}
	case commonparams.TypeCodeObject:
		if _, ok := fieldValue.(*types.Document); !ok {
			return false, nil
		}
	case commonparams.TypeCodeDouble:
		if _, ok := fieldValue.(float64); !ok {
			return false, nil
		}
	case commonparams.TypeCodeString:
		if _, ok := fieldValue.(string); !ok {
			return false, nil
		}
	case commonparams.TypeCodeBinData:
		if _, ok := fieldValue.(types.Binary); !ok {
			return false, nil
		}
	case commonparams.TypeCodeObjectID:
		if _, ok := fieldValue.(types.ObjectID); !ok {
			return false, nil
		}
	case commonparams.TypeCodeBool:
		if _, ok := fieldValue.(bool); !ok {
			return false, nil
		}
	case commonparams.TypeCodeDate:
		if _, ok := fieldValue.(time.Time); !ok {
			return false, nil
		}
	case commonparams.TypeCodeNull:
		if _, ok := fieldValue.(types.NullType); !ok {
			return false, nil
		}
	case commonparams.TypeCodeRegex:
		if _, ok := fieldValue.(types.Regex); !ok {
			return false, nil
		}
	case commonparams.TypeCodeInt:
		if _, ok := fieldValue.(int32); !ok {
			return false, nil
		}
	case commonparams.TypeCodeTimestamp:
		if _, ok := fieldValue.(types.Timestamp); !ok {
			return false, nil
		}
	case commonparams.TypeCodeLong:
		if _, ok := fieldValue.(int64); !ok {
			return false, nil
		}
	case commonparams.TypeCodeNumber:
		// TypeCodeNumber should match int32, int64 and float64 types
		switch fieldValue.(type) {
		case float64, int32, int64:
			return true, nil
		default:
			return false, nil
		}
	case commonparams.TypeCodeDecimal, commonparams.TypeCodeMinKey, commonparams.TypeCodeMaxKey:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf(`Type code %v not implemented`, code),
			"$type",
		)
	default:
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(`Unknown type name alias: %s`, code.String()),
			"$type",
		)
	}

	return true, nil
}

// filterFieldExprElemMatch handles {field: {$elemMatch: value}}.
// Returns false if doc value is not an array.
func filterFieldExprElemMatch(doc *types.Document, filterKey, filterSuffix string, exprValue any) (bool, error) {
	expr, ok := exprValue.(*types.Document)
	if !ok {
		return false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"$elemMatch needs an Object",
			"$elemMatch",
		)
	}

	for _, key := range expr.Keys() {
		if slices.Contains([]string{"$text", "$where"}, key) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("%s can only be applied to the top-level document", key),
				"$elemMatch",
			)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/730
		if slices.Contains([]string{"$and", "$or", "$nor"}, key) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("$elemMatch: support for %s not implemented yet", key),
				"$elemMatch",
			)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/731
		if slices.Contains([]string{"$ne", "$not"}, key) {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("$elemMatch: support for %s not implemented yet", key),
				"$elemMatch",
			)
		}

		if expr.Len() > 1 && !strings.HasPrefix(key, "$") {
			return false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("unknown operator: %s", key),
				"$elemMatch",
			)
		}
	}

	value, err := doc.Get(filterSuffix)
	if err != nil {
		return false, nil
	}

	if _, ok := value.(*types.Array); !ok {
		return false, nil
	}

	return filterFieldExpr(doc, filterKey, filterSuffix, expr)
}

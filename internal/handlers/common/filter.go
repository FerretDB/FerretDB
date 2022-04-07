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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FilterDocument returns true if given document satisfies given filter expression.
//
// Passed arguments must not be modified.
func FilterDocument(doc, filter *types.Document) (bool, error) {
	filterMap := filter.Map()
	if len(filterMap) == 0 {
		return true, nil
	}

	// top-level filters are ANDed together
	for _, filterKey := range filter.Keys() {
		filterValue := filterMap[filterKey]
		matches, err := filterDocumentPair(doc, filterKey, filterValue)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// filterDocumentPair handles a single filter element key/value pair {filterKey: filterValue}.
func filterDocumentPair(doc *types.Document, filterKey string, filterValue any) (bool, error) {
	if strings.HasPrefix(filterKey, "$") {
		// {$operator: filterValue}
		return filterOperator(doc, filterKey, filterValue)
	}

	docValue, err := doc.Get(filterKey)
	if err != nil {
		return false, nil // no error - the field is just not present
	}

	switch filterValue := filterValue.(type) {
	case *types.Document:
		// {field: {expr}}
		return filterFieldExpr(docValue, filterValue)

	case *types.Array:
		// {field: [array]}
		panic("not implemented")

	case types.Regex:
		// {field: /regex/}
		return filterFieldRegex(docValue, filterValue)

	default:
		// {field: value}
		switch docValue := docValue.(type) {
		case *types.Document, *types.Array:
			return compare(docValue, filterValue) == equal, nil
		}

		return compareScalars(docValue, filterValue) == equal, nil
	}
}

// filterOperator handles a top-level operator filter {$operator: filterValue}.
func filterOperator(doc *types.Document, operator string, filterValue any) (bool, error) {
	switch operator {
	case "$and":
		// {$and: [{expr1}, {expr2}, ...]}
		exprs, err := AssertType[*types.Array](filterValue)
		if err != nil {
			return false, err
		}
		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)
			matches, err := FilterDocument(doc, expr)
			if err != nil {
				panic(err)
			}
			if !matches {
				return false, nil
			}
		}
		return true, nil

	case "$or":
		// {$or: [{expr1}, {expr2}, ...]}
		exprs, err := AssertType[*types.Array](filterValue)
		if err != nil {
			return false, err
		}
		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)
			matches, err := FilterDocument(doc, expr)
			if err != nil {
				panic(err)
			}
			if matches {
				return true, nil
			}
		}
		return false, nil

	case "$nor":
		// {$nor: [{expr1}, {expr2}, ...]}
		exprs, err := AssertType[*types.Array](filterValue)
		if err != nil {
			return false, err
		}
		for i := 0; i < exprs.Len(); i++ {
			expr := must.NotFail(exprs.Get(i)).(*types.Document)
			matches, err := FilterDocument(doc, expr)
			if err != nil {
				panic(err)
			}
			if matches {
				return false, nil
			}
		}
		return true, nil

	default:
		msg := fmt.Sprintf(
			`unknown top level operator: %s. `+
				`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			operator,
		)
		return false, NewErrorMsg(ErrBadValue, msg)
	}
}

// filterFieldExpr handles {field: {expr}} filter.
func filterFieldExpr(fieldValue any, expr *types.Document) (bool, error) {
	for _, exprKey := range expr.Keys() {
		if exprKey == "$options" {
			// handled by $regex
			continue
		}

		exprValue := must.NotFail(expr.Get(exprKey))

		switch exprKey {
		case "$eq":
			// {field: {$eq: exprValue}}
			if compare(fieldValue, exprValue) != equal {
				return false, nil
			}

		case "$ne":
			// {field: {$ne: exprValue}}
			// TODO regex
			if compareScalars(fieldValue, exprValue) == equal {
				return false, nil
			}

		case "$gt":
			// {field: {$gt: exprValue}}
			if compare(fieldValue, exprValue) != greater {
				return false, nil
			}

		case "$gte":
			// {field: {$gte: exprValue}}
			if c := compare(fieldValue, exprValue); c != greater && c != equal {
				return false, nil
			}

		case "$lt":
			// {field: {$lt: exprValue}}
			if c := compare(fieldValue, exprValue); c != less {
				return false, nil
			}

		case "$lte":
			// {field: {$lte: exprValue}}
			if c := compare(fieldValue, exprValue); c != less && c != equal {
				return false, nil
			}

		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			arr := exprValue.(*types.Array)
			var found bool
			for i := 0; i < arr.Len(); i++ {
				arrValue := must.NotFail(arr.Get(i))
				if compareScalars(fieldValue, arrValue) == equal {
					found = true
					break
				}
			}
			if !found {
				return false, nil
			}

		case "$nin":
			// {field: {$nin: [value1, value2, ...]}}
			arr := exprValue.(*types.Array)
			var found bool
			for i := 0; i < arr.Len(); i++ {
				arrValue := must.NotFail(arr.Get(i))
				if compareScalars(fieldValue, arrValue) == equal {
					found = true
					break
				}
			}
			if found {
				return false, nil
			}

		case "$not":
			// {field: {$not: {expr}}}
			expr := exprValue.(*types.Document)
			res, err := filterFieldExpr(fieldValue, expr)
			if res || err != nil {
				return false, err
			}

		case "$regex":
			// {field: {$regex: exprValue}}
			optionsAny, _ := expr.Get("$options")
			res, err := filterFieldExprRegex(fieldValue, exprValue, optionsAny)
			if !res || err != nil {
				return false, err
			}

		case "$size":
			// {field: {$size: value}}
			res, err := filterFieldExprSize(fieldValue, exprValue)
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

		default:
			panic(fmt.Sprintf("filterFieldExpr: %q", exprKey))
		}
	}

	return true, nil
}

// filterFieldRegex handles {field: /regex/} filter. Provides regular expression capabilities
// for pattern matching strings in queries, even if the strings are in an array.
func filterFieldRegex(fieldValue any, regex types.Regex) (bool, error) {
	re, err := regex.Compile()
	if err != nil {
		return false, err
	}

	switch fieldValue := fieldValue.(type) {
	case string:
		return re.MatchString(fieldValue), nil

	case *types.Array:
		for i := 0; i < fieldValue.Len(); i++ {
			arrValue := must.NotFail(fieldValue.Get(i)).(any)
			s, isString := arrValue.(string)
			if !isString {
				continue
			}
			if re.MatchString(s) == true {
				return true, nil
			}
		}

	case types.Regex:
		return compareScalars(fieldValue, regex) == equal, nil
	}

	return false, nil
}

// filterFieldExprRegex handles {field: {$regex: regexValue, $options: optionsValue}} filter.
func filterFieldExprRegex(fieldValue any, regexValue, optionsValue any) (bool, error) {
	var options string
	if optionsValue != nil {
		var ok bool
		if options, ok = optionsValue.(string); !ok {
			return false, NewErrorMsg(ErrBadValue, "$options has to be a string")
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
				return false, NewErrorMsg(ErrRegexOptions, "options set in both $regex and $options")
			}
			regexValue.Options = options
		}
		return filterFieldRegex(fieldValue, regexValue)

	default:
		return false, NewErrorMsg(ErrBadValue, "$regex has to be a string")
	}
}

// filterFieldExprSize handles {field: {$size: sizeValue}} filter.
func filterFieldExprSize(fieldValue any, sizeValue any) (bool, error) {
	arr, ok := fieldValue.(*types.Array)
	if !ok {
		return false, nil
	}

	size, err := GetWholeNumberParam(sizeValue)
	if err != nil {
		switch err {
		case errUnexpectedType:
			return false, NewErrorMsg(ErrBadValue, "$size needs a number")
		case errNotWholeNumber:
			return false, NewErrorMsg(ErrBadValue, "$size must be a whole number")
		default:
			return false, err
		}
	}

	if size < 0 {
		return false, NewErrorMsg(ErrBadValue, "$size may not be negative")
	}

	if arr.Len() != int(size) {
		return false, nil
	}

	return true, nil
}

// filterFieldExprBitsAllClear handles {field: {$bitsAllClear: value}} filter.
func filterFieldExprBitsAllClear(fieldValue, maskValue any) (bool, error) {
	fieldBinary, maskBinary, err := getBinaryParams(fieldValue, maskValue)
	if err != nil {
		return false, formatBitwiseOperatorErr(err, "$bitsAllClear", maskValue)
	}

	for i := 0; i < len(fieldBinary.B); i++ {
		if (fieldBinary.B[i] & maskBinary.B[i]) != 0 {
			return false, nil
		}
	}

	return true, nil
}

// filterFieldExprBitsAllSet handles {field: {$bitsAllSet: value}} filter.
func filterFieldExprBitsAllSet(fieldValue, maskValue any) (bool, error) {
	fieldBinary, maskBinary, err := getBinaryParams(fieldValue, maskValue)
	if err != nil {
		return false, formatBitwiseOperatorErr(err, "$bitsAllSet", maskValue)
	}

	for i := 0; i < len(fieldBinary.B); i++ {
		if (fieldBinary.B[i] & maskBinary.B[i]) != maskBinary.B[i] {
			return false, nil
		}
	}

	return true, nil
}

// filterFieldExprBitsAnyClear handles {field: {$bitsAnyClear: value}} filter.
func filterFieldExprBitsAnyClear(fieldValue, maskValue any) (bool, error) {
	fieldBinary, maskBinary, err := getBinaryParams(fieldValue, maskValue)
	if err != nil {
		return false, formatBitwiseOperatorErr(err, "$bitsAnyClear", maskValue)
	}

	for i := 0; i < len(fieldBinary.B); i++ {
		if fieldBinary.B[i] == 0 {
			continue
		}

		mask := maskBinary.B[i] ^ 0b1111_1111
		if (fieldBinary.B[i] & mask) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// filterFieldExprBitsAnySet handles {field: {$bitsAnySet: value}} filter.
func filterFieldExprBitsAnySet(fieldValue, maskValue any) (bool, error) {
	fieldBinary, maskBinary, err := getBinaryParams(fieldValue, maskValue)
	if err != nil {
		return false, formatBitwiseOperatorErr(err, "$bitsAnySet", maskValue)
	}

	for i := 0; i < len(fieldBinary.B); i++ {
		if fieldBinary.B[i] == 0 {
			continue
		}

		mask := maskBinary.B[i] ^ 0b0000_0000
		if (fieldBinary.B[i] & mask) == 0 {
			return false, nil
		}
	}

	return true, nil
}

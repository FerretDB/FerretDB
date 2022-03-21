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
	"math"
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
	// {$operator: [expr1, expr2, ...]}
	if strings.HasPrefix(filterKey, "$") {
		exprs, ok := filterValue.(*types.Array)
		if !ok {
			msg := fmt.Sprintf(
				`unknown top level operator: %s. `+
					`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
				filterKey,
			)
			return false, NewErrorMsg(ErrBadValue, msg)
		}

		switch filterKey {
		case "$and":
			// {$and: [{expr1}, {expr2}, ...]}
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
		}

		panic(fmt.Sprintf("filterDocumentPair: %q %v", filterKey, filterValue))
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
		return compareScalars(docValue, filterValue) == equal, nil
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
			// TODO regex
			if compareScalars(fieldValue, exprValue) != equal {
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
			if c := compareScalars(fieldValue, exprValue); c != greater {
				return false, nil
			}

		case "$gte":
			// {field: {$gte: exprValue}}
			if c := compareScalars(fieldValue, exprValue); c != greater && c != equal {
				return false, nil
			}

		case "$lt":
			// {field: {$lt: exprValue}}
			if c := compareScalars(fieldValue, exprValue); c != less {
				return false, nil
			}

		case "$lte":
			// {field: {$lte: exprValue}}
			if c := compareScalars(fieldValue, exprValue); c != less && c != equal {
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

		default:
			panic(fmt.Sprintf("filterFieldExpr: %q", exprKey))
		}
	}

	return true, nil
}

// filterFieldRegex handles {field: /regex/} filter.
func filterFieldRegex(fieldValue any, regex types.Regex) (bool, error) {
	s, ok := fieldValue.(string)
	if !ok {
		return false, nil
	}

	re, err := regex.Compile()
	if err != nil {
		return false, err
	}

	return re.MatchString(s), nil
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

	var size int
	switch sizeValue := sizeValue.(type) {
	case float64:
		if sizeValue != math.Trunc(sizeValue) || math.IsNaN(sizeValue) || math.IsInf(sizeValue, 0) {
			return false, NewErrorMsg(ErrBadValue, "$size must be a whole number")
		}
		size = int(sizeValue)
	case int32:
		size = int(sizeValue)
	case int64:
		size = int(sizeValue)
	default:
		return false, NewErrorMsg(ErrBadValue, "$size needs a number")
	}

	// TODO check float negative zero

	if size < 0 {
		return false, NewErrorMsg(ErrBadValue, "$size may not be negative")
	}

	if arr.Len() != size {
		return false, nil
	}

	return true, nil
}

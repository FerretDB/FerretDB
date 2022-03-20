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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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
		matches, err := filterDocumentFoo(doc, filterKey, filterValue)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

func filterDocumentFoo(doc *types.Document, filterKey string, filterValue any) (bool, error) {
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
			panic("$nor")
		}

		panic(lazyerrors.Errorf("lala1 key %q, value %v", filterKey, filterValue))
	}

	docValue := must.NotFail(doc.Get(filterKey))

	switch filterValue := filterValue.(type) {
	case *types.Document:
		// {field: {expr}}
		return filterFieldExpr(docValue, filterValue)

	case *types.Array:
		panic("oops array")

	case types.Regex:
		// {field: /regex/}
		return filterFieldRegex(docValue, filterValue)

	default:
		// {field: value}
		return filterCompareScalars(docValue, filterValue) == equal, nil
	}
}

// filterFieldExpr returns true if given value satisfies given expression.
//
// It handles `{field: {expr}}`.
func filterFieldExpr(docValue any, expr *types.Document) (bool, error) {
	for _, exprKey := range expr.Keys() {
		if exprKey == "$options" {
			// handled by $regex
			continue
		}

		exprValue := must.NotFail(expr.Get(exprKey))

		switch exprKey {
		case "$eq":
			// {field: {$eq: value}}
			// TODO regex
			if filterCompareScalars(docValue, exprValue) != equal {
				return false, nil
			}

		case "$ne":
			// {field: {$ne: value}}
			// TODO regex
			if filterCompareScalars(docValue, exprValue) == equal {
				return false, nil
			}

		case "$gt":
			panic("$gt")
		case "$gte":
			panic("$gte")
		case "$lt":
			panic("$lt")
		case "$lte":
			panic("$lte")

		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			arr := exprValue.(*types.Array)
			var found bool
			for i := 0; i < arr.Len(); i++ {
				arrValue := must.NotFail(arr.Get(i))
				if filterCompareScalars(docValue, arrValue) == equal {
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
				if filterCompareScalars(docValue, arrValue) == equal {
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
			res, err := filterFieldExpr(docValue, expr)
			if !res || err != nil {
				return res, err
			}

		case "$regex":
			// {field: {$regex: value}}
			optionsAny, _ := expr.Get("$options")
			res, err := filterFieldExprRegex(docValue, exprValue, optionsAny)
			if !res || err != nil {
				return res, err
			}

		case "$size":
			// {field: {$size: value}}
			res, err := filterFieldExprSize(docValue, exprValue)
			if !res || err != nil {
				return res, err
			}

		default:
			panic(fmt.Sprintf("filterFieldExpr: %q", exprKey))
		}
	}

	return true, nil
}

// {field: /regex/}
func filterFieldRegex(docValue any, regex types.Regex) (bool, error) {
	docString, ok := docValue.(string)
	if !ok {
		return false, nil
	}

	re, err := regex.Compile()
	if err != nil {
		return false, err
	}

	return re.MatchString(docString), nil
}

func filterFieldExprRegex(docValue any, exprValue, optionsAny any) (bool, error) {
	var options string
	if optionsAny != nil {
		var ok bool
		if options, ok = optionsAny.(string); !ok {
			return false, NewErrorMsg(ErrBadValue, "$options has to be a string")
		}
	}

	switch exprValue := exprValue.(type) {
	case string:
		regex := types.Regex{
			Pattern: exprValue,
			Options: options,
		}
		return filterFieldRegex(docValue, regex)

	case types.Regex:
		if options != "" {
			if exprValue.Options != "" {
				return false, NewErrorMsg(ErrRegexOptions, "options set in both $regex and $options")
			}
			exprValue.Options = options
		}
		return filterFieldRegex(docValue, exprValue)

	default:
		return false, NewErrorMsg(ErrBadValue, "$regex has to be a string")
	}
}

// {field: {$size: value}}
func filterFieldExprSize(docValue any, exprValue any) (bool, error) {
	arr, ok := docValue.(*types.Array)
	if !ok {
		return false, nil
	}

	var value int
	switch exprValue := exprValue.(type) {
	case float64:
		if exprValue != math.Trunc(exprValue) || math.IsNaN(exprValue) || math.IsInf(exprValue, 0) {
			return false, NewErrorMsg(ErrBadValue, "$size must be a whole number")
		}
		value = int(exprValue)
	case int32:
		value = int(exprValue)
	case int64:
		value = int(exprValue)
	default:
		return false, NewErrorMsg(ErrBadValue, "$size needs a number")
	}

	// TODO check float negative zero

	if value < 0 {
		return false, NewErrorMsg(ErrBadValue, "$size may not be negative")
	}

	if arr.Len() != value {
		return false, nil
	}

	return true, nil
}

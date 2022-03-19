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

package jsonb1

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// filterDocument returns true if given document matches given filter condition.
//
// Passed arguments must not be modified.
func filterDocument(doc, filter *types.Document) (bool, error) {
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
			return false, common.NewErrorMsg(common.ErrBadValue, msg)
		}

		switch filterKey {
		case "$and":
			for i := 0; i < exprs.Len(); i++ {
				expr := must.NotFail(exprs.Get(i)).(*types.Document)

				matches, err := filterDocument(doc, expr)
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

				matches, err := filterDocument(doc, expr)
				if err != nil {
					panic(err)
				}
				if matches {
					return true, nil
				}
			}
			return false, nil
		}

		panic(lazyerrors.Errorf("lala1 key %q, value %v", filterKey, filterValue))
	}

	docValue := must.NotFail(doc.Get(filterKey))

	switch filterValue := filterValue.(type) {
	case *types.Document:
		// {field: {expr}}
		return filterFieldExpr(docValue, filterValue), nil

	case *types.Array:
		panic("oops array")

	case types.Regex:
		panic("oops regex")

	default:
		// {field: value}
		return filterScalarEqual(docValue, filterValue), nil
	}
}

func filterFieldExpr(docValue any, expr *types.Document) bool {
	for _, key := range expr.Keys() {
		switch key {
		case "$not":
			// {field: {$not: {expr}}}
			expr := must.NotFail(expr.Get(key)).(*types.Document)
			if filterFieldExpr(docValue, expr) {
				return false
			}

		case "$eq":
			v := must.NotFail(expr.Get(key))
			if !filterScalarEqual(docValue, v) {
				return false
			}
		default:
			panic(key)
		}
	}

	return true
}

// filterScalarEqual returns true if given scalar values are equal as used by filters.
func filterScalarEqual(a, b any) bool {
	if a == nil {
		panic("a is nil")
	}
	if b == nil {
		panic("b is nil")
	}

	switch a := a.(type) {
	case float64:
		return a == b.(float64)
	case string:
		return a == b.(string)
	case types.Binary:
		b := b.(types.Binary)
		return a.Subtype == b.Subtype && bytes.Equal(a.B, b.B)
	case types.ObjectID:
		return a == b.(types.ObjectID)
	case bool:
		return a == b.(bool)
	case time.Time:
		return a.Equal(b.(time.Time))
	case types.NullType:
		_ = b.(types.NullType)
		return true
	case types.Regex:
		return a == b.(types.Regex)
	case int32:
		return a == b.(int32)
	case types.Timestamp:
		return a == b.(types.Timestamp)
	case int64:
		return a == b.(int64)
	default:
		panic(fmt.Sprintf("unhandled type %T", a))
	}
}

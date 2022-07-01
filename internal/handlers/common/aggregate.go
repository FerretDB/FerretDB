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

type parseContext struct {
	parents []string
	values  *[]interface{}
}

func FormatFieldWithAncestor(field string, parents []string, ancestor string) string {
	newParents := make([]string, len(parents)+1)
	copy(newParents[1:], parents)
	newParents[0] = ancestor
	return FormatField(field, newParents)
}

func FormatField(field string, parents []string) string {
	return FormatFieldWithSeparators(field, parents, "->", "->>")
}

func FormatFieldWithSeparators(field string, parents []string, initSep string, otherSep string) string {
	if len(parents) == 0 {
		return field
	}
	res := ""
	fields := parents
	if field != "" {
		fields = append(fields, field)
	}
	for i, p := range fields {
		sep := ""
		if i < len(fields)-1 {
			sep = initSep
			if i > 0 {
				sep = otherSep
			}
		}
		fmtParent := p
		if i > 0 {
			fmtParent = `'` + p + `'`
		}
		res += fmt.Sprintf("%s%s", fmtParent, sep)
	}
	return res
}

func HandleJoin(ctx *parseContext, oper string, arr *types.Array) (string, error) {
	sql := "("
	for i := 0; i < arr.Len(); i++ {
		v := must.NotFail(arr.Get(i))
		if i > 0 {
			sql += " " + oper + " "
		}
		s, err := MatchToSql(ctx, "", v)
		if err != nil {
			return "", err
		}
		sql += *s
	}
	sql += ")"

	return sql, nil
}

func AddOperator(ctx *parseContext, format string, value interface{}) string {
	*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
	field := FormatField("", ctx.parents)
	return fmt.Sprintf(format, field, `$`+fmt.Sprintf("%v", len(*ctx.values)))
}

func MatchToSql(ctx *parseContext, key string, value interface{}) (*string, error) {
	var sql string

	switch key {
	case "$or":
		arr, ok := value.(*types.Array)
		if !ok {
			return nil, NewErrorMsg(ErrBadValue, "$or must be an array")
		}
		r, err := HandleJoin(ctx, "OR", arr)
		if err != nil {
			return nil, err
		}
		sql = r

	case "$and":
		arr, ok := value.(*types.Array)
		if !ok {
			return nil, NewErrorMsg(ErrBadValue, "$and must be an array")
		}
		r, err := HandleJoin(ctx, "AND", arr)
		if err != nil {
			return nil, err
		}
		sql = r

	case "$gt":
		sql = AddOperator(ctx, `%v > %v`, value)

	case "$gte":
		sql = AddOperator(ctx, `%v >= %v`, value)

	case "$lt":
		sql = AddOperator(ctx, `%v < %v`, value)

	case "$lte":
		sql = AddOperator(ctx, `%v <= %v`, value)

	case "$ne":
		sql = AddOperator(ctx, `%v <> %v`, value)

	case "$exists":
		parentValue := ctx.parents[len(ctx.parents)-1]
		*ctx.values = append(*ctx.values, fmt.Sprintf("%v", parentValue))

		parents := ctx.parents[:len(ctx.parents)-1]
		field := strings.Replace(FormatField("", parents), "_jsonb", "_jsonb::jsonb", -1)
		sql = field + ` ? $` + fmt.Sprintf("%v", len(*ctx.values))
		if value == false {
			sql = "NOT (" + sql + ")"
		}

	case "$in":
		arr, ok := value.(*types.Array)
		if !ok {
			return nil, NewErrorMsg(ErrBadValue, "$in must be an array")
		}

		arrVals := []string{}
		for i := 0; i < arr.Len(); i++ {
			arrVals = append(arrVals, fmt.Sprintf("%v", must.NotFail(arr.Get(i))))
		}

		*ctx.values = append(*ctx.values, arrVals)
		field := FormatField("", ctx.parents)
		sql = field + ` = ANY($` + fmt.Sprintf("%v", len(*ctx.values)) + `)`

	case "$nin":
		arr, ok := value.(*types.Array)
		if !ok {
			return nil, NewErrorMsg(ErrBadValue, "$in must be an array")
		}

		arrVals := []string{}
		for i := 0; i < arr.Len(); i++ {
			arrVals = append(arrVals, fmt.Sprintf("%v", must.NotFail(arr.Get(i))))
		}

		*ctx.values = append(*ctx.values, arrVals)
		field := FormatField("", ctx.parents)
		sql = field + ` <> ALL($` + fmt.Sprintf("%v", len(*ctx.values)) + `)`

	case "$all":
		arr, ok := value.(*types.Array)
		if !ok {
			return nil, NewErrorMsg(ErrBadValue, "$all must be an array")
		}

		arrVals := []string{}
		for i := 0; i < arr.Len(); i++ {
			arrVals = append(arrVals, fmt.Sprintf("%v", must.NotFail(arr.Get(i))))
		}

		*ctx.values = append(*ctx.values, arrVals)
		field := FormatField("", ctx.parents)
		sql = field + ` @> ($` + fmt.Sprintf("%v", len(*ctx.values)) + `)`

	case "$not":
		sql = "NOT ("
		s, err := MatchToSql(ctx, "", value)
		if err != nil {
			return nil, err
		}
		sql += *s
		sql += ")"

	case "$regex":
		*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
		field := FormatFieldWithSeparators("", ctx.parents, "->>", "->>")
		sql = field + ` ~ $` + fmt.Sprintf("%v", len(*ctx.values))

	default:
		if strings.HasPrefix(key, "$") {
			return nil, NewWriteErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf(
					"Unknown top level operator: %s. Expected a valid aggregate modifier", key,
				),
			)
		}

		switch v := value.(type) {
		case *types.Document:
			if key != "" {
				ctx.parents = append(ctx.parents, key)
			}
			sql = "("
			for i, key := range v.Keys() {
				value := must.NotFail(v.Get(key))
				s, err := MatchToSql(ctx, key, value)
				if err != nil {
					return nil, err
				}
				if i > 0 {
					sql += " AND "
				}
				sql += *s
			}
			sql += ")"

		default:
			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` = $` + fmt.Sprintf("%v", len(*ctx.values))
		}
	}

	return &sql, nil
}

func AggregateMatch(match *types.Document, parent string) (*string, []interface{}, error) {
	ctx := parseContext{
		parents: []string{},
		values:  &[]interface{}{},
	}

	sql, err := MatchToSql(&ctx, parent, match)
	return sql, *ctx.values, err
}

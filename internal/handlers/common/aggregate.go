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

func FormatField(field string, parents []string) string {
	if len(parents) == 0 {
		return field
	}
	res := ""
	for i, p := range parents {
		sep := "->"
		if i > 0 {
			sep = "->>"
		}
		res += fmt.Sprintf("'%s'%s", p, sep)
	}
	return res + "'" + field + "'"
}

func MatchToSql(ctx *parseContext, key string, value interface{}) (*string, error) {
	var sql string

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
		switch key {
		case "$or":
			arr, ok := value.(*types.Array)
			if !ok {
				return nil, NewErrorMsg(ErrBadValue, "$or must be an array")
			}
			sql = "("
			for i := 0; i < arr.Len(); i++ {
				v := must.NotFail(arr.Get(i))
				if i > 0 {
					sql += " OR "
				}
				s, err := MatchToSql(ctx, "", v)
				if err != nil {
					return nil, err
				}
				sql += *s
			}
			sql += ")"

		case "$gt":
			fmt.Printf("  *** Key:   %v\n", key)
			fmt.Printf("  *** Value: %v\n", value)
			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` > $` + fmt.Sprintf("%v", len(*ctx.values))

		case "$gte":
			fmt.Printf("  *** Key:   %v\n", key)
			fmt.Printf("  *** Value: %v\n", value)
			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` >= $` + fmt.Sprintf("%v", len(*ctx.values))

		case "$lt":
			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` < $` + fmt.Sprintf("%v", len(*ctx.values))

		case "$lte":
			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` <= $` + fmt.Sprintf("%v", len(*ctx.values))

		default:
			if strings.HasPrefix(key, "$") {
				return nil, NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf(
						"Unknown top level operator: %s. Expected a valid aggregate modifier", key,
					),
				)
			}

			*ctx.values = append(*ctx.values, fmt.Sprintf("%v", value))
			field := FormatField(key, ctx.parents)
			sql = field + ` = $` + fmt.Sprintf("%v", len(*ctx.values))
		}
	}

	return &sql, nil
}

func AggregateMatch(match *types.Document) (*string, []interface{}, error) {
	ctx := parseContext{
		parents: []string{},
		values:  &[]interface{}{},
	}

	sql, err := MatchToSql(&ctx, "_jsonb", match)
	return sql, *ctx.values, err
}

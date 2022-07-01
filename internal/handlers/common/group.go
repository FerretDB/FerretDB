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

type GroupContext struct {
	parents   []string
	fields    []interface{}
	groups    []string
	subFields []string
	subGroups []string
	distinct  string
}

func NewGroupContext() GroupContext {
	ctx := GroupContext{
		parents:   []string{},
		fields:    []interface{}{},
		subFields: []string{},
		groups:    []string{},
		subGroups: []string{},
	}

	return ctx
}

func (c *GroupContext) AddField(name string, value interface{}) {
	c.fields = append(c.fields, name)
	c.fields = append(c.fields, value)
}

func (c *GroupContext) AddSubField(value string) {
	c.subFields = append(c.subFields, value)
}

func (c *GroupContext) AddGroup(name string) {
	c.groups = append(c.groups, name)
}

func (c *GroupContext) AddSubGroup(name string) {
	c.subGroups = append(c.subGroups, name)
}

func (c *GroupContext) GetParent() string {
	return c.parents[len(c.parents)-1]
}

func (c *GroupContext) FieldAsString() string {
	prefix := ""
	if c.distinct != "" {
		prefix = "DISTINCT ON (" + c.distinct + ") "
	}
	str := fmt.Sprintf("%sjson_build_object('$k', jsonb_build_array(", prefix)

	for i, key := range c.fields {
		if i%2 == 0 {
			str += fmt.Sprintf("'%s', ", key)
		}
	}

	str = strings.TrimSuffix(str, ", ") + "), "

	for i, v := range c.fields {
		if i%2 == 0 {
			str += fmt.Sprintf("'%s', ", v)
		} else {
			str += fmt.Sprintf("%v, ", v)
		}
	}
	str = strings.TrimSuffix(str, ", ") + ") AS _jsonb"

	return str
}

func (c *GroupContext) GetSubQuery() string {
	if len(c.subFields) == 0 {
		return ""
	}

	sql := "SELECT " + strings.Join(c.subFields, ", ") + " FROM %s"

	if len(c.subGroups) > 0 {
		sql += " GROUP BY " + strings.Join(c.subGroups, ", ") + ", "
	}

	return sql
}

func GetNumericValue(field string) string {
	return fmt.Sprintf(`(CASE WHEN (%s ? '$f') THEN (%s->>'$f')::numeric ELSE (%s)::numeric END)`, field, field, field)
}

func ParseOperators(ctx *GroupContext, parentKey string, doc *types.Document) error {
	for _, key := range doc.Keys() {
		value := must.NotFail(doc.Get(key))
		switch key {
		case "$dateToString":
			params := value.(*types.Document)
			date, err := params.Get("date")
			if err != nil {
				return fmt.Errorf("Error getting 'date': %w", err)
			}
			// FIXME support multiple formats - https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToString/
			// TODO format := params.Get("format")
			if date == nil {
				return NewWriteErrorMsg(
					ErrFailedToParse,
					"Missing 'date' parameter to $dateToString",
				)
			}

			field := strings.TrimPrefix(date.(string), "$")
			res := FormatFieldWithAncestor(field, ctx.parents, "_jsonb") + "->>'$d'"
			ctx.AddField(parentKey, parentKey)
			ctx.AddSubField(fmt.Sprintf("TO_CHAR(TO_TIMESTAMP((%s)::numeric / 1000), 'YYYY-MM-DD') AS %s", res, parentKey))
			if parentKey == "_id" {
				ctx.AddSubGroup("_id")
			}

			return nil

		case "$multiply":
			params := value.(*types.Array)
			fields := ""

			for i := 0; i < params.Len(); i++ {
				field := must.NotFail(params.Get(i)).(string)
				if strings.HasPrefix(field, "$") {
					res := FormatFieldWithAncestor(strings.TrimPrefix(field, "$"), ctx.parents, "_jsonb")
					fields += GetNumericValue(res)
				} else {
					fields += field
				}
				fields += " * "
			}
			fields = "(" + strings.TrimSuffix(fields, " * ") + ") AS " + parentKey

			ctx.AddField(parentKey, parentKey)
			ctx.AddSubField(fields)
		}
	}

	return nil
}

func ParseGroup(ctx *GroupContext, key string, value interface{}) error {
	switch key {
	case "_id":
		switch v := value.(type) {
		case *types.Document:
			err := ParseOperators(ctx, key, v)
			if err != nil {
				return err
			}

		default:
			if strings.HasPrefix(value.(string), "$") {
				field := strings.TrimPrefix(value.(string), "$")
				field = FormatFieldWithAncestor(field, ctx.parents, "_jsonb")
				ctx.distinct = field
				ctx.AddField("_id", field)
			} else {
				ctx.AddField("_id", value)
			}
		}

	case "$count":
		ctx.AddField(ctx.GetParent(), "COUNT(*)")

	case "$sum":
		// FIXME Support array of fields to sum

		// FIXME we are always casting the avg to a float64, check if we can find a way
		//       to dynamically detect int vs. float
		ctx.AddField(ctx.GetParent(), "json_build_object('$f', "+ctx.GetParent()+")")

		switch param := value.(type) {
		case string:
			if !strings.HasPrefix(param, "$") {
				return NewWriteErrorMsg(
					ErrFailedToParse,
					fmt.Sprintf("Invalid '%s' parameter to $sum: must start with $ (temporarily)", param),
				)
			}

			res := FormatFieldWithAncestor(strings.TrimPrefix(param, "$"), []string{}, "_jsonb")
			ctx.AddSubField(fmt.Sprintf("SUM(%s) AS %s", GetNumericValue(res), ctx.GetParent()))

		case int32, int64, float64:
			ctx.AddSubField(fmt.Sprintf("SUM(%v) AS %s", param, ctx.GetParent()))

		case *types.Array:
			return NewWriteErrorMsg(
				ErrFailedToParse,
				"The $sum accumulator is a unary operator",
			)
		}

	case "$avg":
		param := value.(string)
		if !strings.HasPrefix(param, "$") {
			// FIXME handle constant expressions (?)
			return NewWriteErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf("Invalid '%s' parameter to $avg: must start with $ (temporarily)", param),
			)
		}

		res := FormatFieldWithAncestor(strings.TrimPrefix(param, "$"), []string{}, "_jsonb")
		// FIXME we are always casting the avg to a float64, check if we can find a way
		//       to dynamically detect int vs. float
		ctx.AddField(ctx.GetParent(), "json_build_object('$f', "+ctx.GetParent()+")")
		ctx.AddSubField(fmt.Sprintf("AVG(%s) AS %s", GetNumericValue(res), ctx.GetParent()))

	default:
		if strings.HasPrefix(key, "$") {
			return NewWriteErrorMsg(
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
			for _, key := range v.Keys() {
				value := must.NotFail(v.Get(key))
				err := ParseGroup(ctx, key, value)
				if err != nil {
					return err
				}
			}
		default:
			fmt.Printf("  *** BOTTOM %#v", v)

		}
	}

	return nil
}

type AggregateResult struct {
	Fields   string
	Groups   string
	SubQuery string
}

func AggregateGroup(group *types.Document) (*AggregateResult, error) {
	ctx := NewGroupContext()

	err := ParseGroup(&ctx, "", group)
	if err != nil {
		return nil, err
	}

	fields := ctx.FieldAsString()
	groups := ""
	subQuery := ctx.GetSubQuery()

	res := AggregateResult{fields, groups, subQuery}

	return &res, nil
}

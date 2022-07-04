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

package aggregate

import (
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type Field struct {
	name     string
	type_    string
	contents string
}

func (f *Field) ToSql() string {
	return f.contents + " AS " + f.name
}

type GroupParser struct {
	fields  []Field
	groups  []string
	parents []string
	// distinct string
}

func (gp *GroupParser) AddField(name, type_ string, contents string) {
	gp.fields = append(gp.fields, Field{name, type_, contents})
}

func (gp *GroupParser) AddGroup(name string) {
	gp.groups = append(gp.groups, name)
}

func (gp *GroupParser) AddParent(name string) {
	gp.parents = append(gp.parents, name)
}

func (mp *GroupParser) parse(key string, value interface{}) error {
	return mp.parseWithParent(key, value, "")
}

func (mp *GroupParser) parseWithParent(key string, value interface{}, parent string) error {
	switch key {
	case "_id":
		switch v := value.(type) {
		case string:
			if strings.HasPrefix(value.(string), "$") {
				field := strings.TrimPrefix(v, "$")
				mp.AddField("_id", "", FormatField(field, []string{"_jsonb"}))
				mp.AddGroup("_id")
			} else {
				mp.AddField("_id", "", v)
			}

		case *types.Document:
			group, err := mp.parseOperators(value.(*types.Document), key)
			if err != nil {
				return err
			}
			mp.AddField("_id", "", *group)
			mp.AddGroup("_id")

		default:
			mp.AddField("_id", "", fmt.Sprintf("%v", value))
			mp.AddGroup("_id")

		}

	case "$sum":
		// FIXME Support document with an operation on fields to sum, like $multiply
		switch param := value.(type) {
		case string:
			if !strings.HasPrefix(param, "$") {
				return common.NewWriteErrorMsg(
					common.ErrFailedToParse,
					fmt.Sprintf("Invalid '%s' parameter to $sum: must start with $ (temporarily)", param),
				)
			}

			field := strings.TrimPrefix(param, "$")
			field = FormatField(field, []string{"_jsonb"})
			contents := fmt.Sprintf("SUM(%s)", GetNumericValue(field))
			mp.AddField(parent, "float", contents)

		case int32, int64, float64:
			contents := fmt.Sprintf("SUM(%v)", param)
			mp.AddField(parent, "float", contents)

		case *types.Document:
			res, err := mp.parseOperators(param, key)
			if err != nil {
				return err
			}

			contents := "SUM(" + *res + ")"
			mp.AddField(parent, "float", contents)

		case *types.Array:
			return common.NewWriteErrorMsg(
				common.ErrFailedToParse,
				"The $sum accumulator is a unary operator",
			)
		}

	case "$avg":
		switch param := value.(type) {
		case string:
			if !strings.HasPrefix(param, "$") {
				// FIXME handle constant expressions (?)
				return common.NewWriteErrorMsg(
					common.ErrFailedToParse,
					fmt.Sprintf("Invalid '%s' parameter to $avg: must start with $ (temporarily)", param),
				)
			}

			field := strings.TrimPrefix(param, "$")
			field = FormatField(field, []string{"_jsonb"})
			contents := fmt.Sprintf("AVG(%s)", GetNumericValue(field))
			mp.AddField(parent, "float", contents)

		case int32, int64, float64:
			contents := fmt.Sprintf("AVG(%v)", param)
			mp.AddField(parent, "float", contents)

		case *types.Document:
			res, err := mp.parseOperators(param, key)
			if err != nil {
				return err
			}

			contents := "AVG(" + *res + ")"
			mp.AddField(parent, "float", contents)

		case *types.Array:
			return common.NewWriteErrorMsg(
				common.ErrFailedToParse,
				"The $avg accumulator is a unary operator",
			)
		}

	default:
		if strings.HasPrefix(key, "$") {
			return common.NewWriteErrorMsg(
				common.ErrFailedToParse,
				fmt.Sprintf(
					"Unknown top level operator: %s. Expected a valid aggregate modifier", key,
				),
			)
		}

		switch v := value.(type) {
		case *types.Document:
			for _, k := range v.Keys() {
				value := must.NotFail(v.Get(k))
				err := mp.parseWithParent(k, value, key)
				if err != nil {
					return err
				}
			}

		default:
			return common.NewErrorMsg(common.ErrFailedToParse, fmt.Sprintf("Could not parse $group of type: %T", v))
		}
	}

	return nil
}

func (mp *GroupParser) parseOperators(doc *types.Document, parent string) (*string, error) {
	for _, key := range doc.Keys() {
		value := must.NotFail(doc.Get(key))
		switch key {
		case "$dateToString":
			params := value.(*types.Document)
			date, err := params.Get("date")
			if err != nil {
				return nil, fmt.Errorf("error getting 'date': %w", err)
			}
			// FIXME support multiple formats - https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToString/
			// TODO format := params.Get("format")
			if date == nil {
				return nil, common.NewWriteErrorMsg(
					common.ErrFailedToParse,
					"Missing 'date' parameter to $dateToString",
				)
			}

			name := strings.TrimPrefix(date.(string), "$")
			field := FormatField(name, []string{"_jsonb"}) + "->>'$d'"
			contents := fmt.Sprintf("TO_CHAR(TO_TIMESTAMP((%s)::numeric / 1000), 'YYYY-MM-DD')", field)
			return &contents, nil

		case "$add", "$subtract", "$multiply", "$divide":
			params := value.(*types.Array)
			fields := ""

			var oper string
			switch key {
			case "$add":
				oper = "+"
			case "$subtract":
				oper = "-"
			case "$multiply":
				oper = "*"
			case "$divide":
				oper = "/"
			}

			for i := 0; i < params.Len(); i++ {
				field := must.NotFail(params.Get(i)).(string)
				if strings.HasPrefix(field, "$") {
					// FIXME we might need to consider parents here
					// res := FormatFieldWithAncestor(strings.TrimPrefix(field, "$"), ctx.parents, "_jsonb")
					res := FormatFieldWithAncestor(strings.TrimPrefix(field, "$"), []string{}, "_jsonb")
					fields += GetNumericValue(res)
				} else {
					fields += field
				}
				fields += " " + oper + " "
			}

			res := strings.TrimSuffix(fields, " "+oper+" ")
			return &res, nil

		default:
			return nil, common.NewWriteErrorMsg(
				common.ErrFailedToParse,
				fmt.Sprintf("Unsupported operator: %s", key),
			)
		}
	}

	return nil, nil
}

func ParseGroupStage(group *types.Document) (*Stage, error) {
	gp := GroupParser{}

	err := gp.parse("", group)
	if err != nil {
		return nil, err
	}

	fmt.Printf("  *** GROUP: %#v\n", gp)

	stage := NewStage("group", gp.groups, nil)
	for _, field := range gp.fields {
		stage.AddField(field.name, field.type_, field.ToSql())
	}

	return &stage, nil
}

func GetNumericValue(field string) string {
	if !strings.Contains(field, "->") && !strings.Contains(field, "->>") {
		parts := strings.Split(field, ".")
		field = "_jsonb"
		for _, part := range parts {
			field += "->'" + part + "'"
		}
	}
	return fmt.Sprintf(`(CASE WHEN (%s ? '$f') THEN (%s->>'$f')::numeric ELSE (%s)::numeric END)`, field, field, field)
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

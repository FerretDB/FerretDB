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
)

func FieldToSql(field string, raw bool) string {
	if field == "" {
		return "_jsonb"
	}

	sql := "_jsonb"
	sep := "->"
	parts := strings.Split(field, ".")
	if len(parts) == 1 {
		if raw {
			sep = "->>"
		}
		sql += sep + "'" + parts[0] + "'"
	} else {
		for i, f := range parts {
			if i == len(parts)-1 {
				sql += `->>'` + f + `'`
			} else {
				sql += sep + `'` + f + `'`
			}
		}
	}
	return sql
}

func ParseField(field string) (string, string) {
	parts := strings.Split(field, ".")
	return parts[len(parts)-1], strings.Join(parts[:len(parts)-1], ".")
}

type FilterNode struct {
	index    int
	op       string
	field    string
	value    interface{}
	parent   *FilterNode
	children []*FilterNode
	unary    bool
	raw      bool
}

func NewRootNode() *FilterNode {
	return &FilterNode{0, "AND", "", nil, nil, []*FilterNode{}, false, false}
}

func NewFieldFilterNode(index int, field string, op string, value interface{}, parent *FilterNode, raw bool) FilterNode {
	return FilterNode{index, op, field, value, parent, []*FilterNode{}, false, raw}
}

func NewOpFilterNode(op string, parent *FilterNode) FilterNode {
	return FilterNode{0, op, "", nil, parent, []*FilterNode{}, false, false}
}

func NewUnaryOpFilterNode(op string, parent *FilterNode) FilterNode {
	return FilterNode{0, op, "", nil, parent, []*FilterNode{}, true, false}
}

func (node *FilterNode) ToSql(json bool) string {
	if len(node.children) > 0 {
		if node.unary {
			if len(node.children) > 1 {
				// FIXME re-evaluate this method of handling unary op errors
				panic("unary operator with multiple children: " + node.op)
			}
			return node.op + " (" + node.children[0].ToSql(json) + ")"
		}
		strs := make([]string, len(node.children))
		for i, child := range node.children {
			str := child.ToSql(json)
			strs[i] = str
		}
		return "(" + strings.Join(strs, " "+node.op+" ") + ")"
	}

	field := node.field
	if json {
		switch node.value.(type) {
		case int32, float64:
			field = GetNumericValue(FieldToSql(node.field, node.raw))
		default:
			field = FieldToSql(node.field, node.raw)
		}
	}
	opValPlaceholder := fmt.Sprintf("%s $%v", node.op, node.index)
	if strings.Contains(node.op, "%s") {
		opValPlaceholder = fmt.Sprintf(node.op, fmt.Sprintf("$%v", node.index))
	}
	return fmt.Sprintf("%s %s", field, opValPlaceholder)
}

func (node *FilterNode) AddRawFilter(index int, field string, op string, value interface{}) *FilterNode {
	child := NewFieldFilterNode(index, field, op, value, node, true)
	node.children = append(node.children, &child)
	return &child
}

func (node *FilterNode) AddFilter(index int, parent string, field string, op string, value interface{}) *FilterNode {
	if parent != "" {
		field = parent + "." + field
	}
	child := NewFieldFilterNode(index, field, op, value, node, false)
	node.children = append(node.children, &child)
	return &child
}

func (node *FilterNode) AddOp(op string) *FilterNode {
	child := NewOpFilterNode(op, node)
	node.children = append(node.children, &child)
	return &child
}

func (node *FilterNode) AddUnaryOp(op string) *FilterNode {
	child := NewUnaryOpFilterNode(op, node)
	node.children = append(node.children, &child)
	return &child
}

func (node *FilterNode) GetValues() []interface{} {
	values := []interface{}{}
	if node.value != nil {
		switch node.value.(type) {
		case float64:
			values = append(values, GetNumericValue(fmt.Sprintf("%v", node.value)))

		case int32:
			values = append(values, fmt.Sprintf("%v", node.value))

		default:
			values = append(values, node.value)
		}
	}
	for _, child := range node.children {
		values = append(values, child.GetValues()...)
	}
	return values
}

type StageField struct {
	name     string
	type_    string
	contents string
}

type StageSortFieldDir int32

const (
	SortAsc StageSortFieldDir = iota
	SortDesc
)

type StageSortField struct {
	name string
	dir  StageSortFieldDir
}

type Stage struct {
	fields     []StageField
	groups     []string
	sortFields []StageSortField
	root       *FilterNode
}

func NewEmptyStage() Stage {
	return Stage{[]StageField{}, []string{}, []StageSortField{}, nil}
}

func NewStage(groups []string, filterTree *FilterNode) Stage {
	return Stage{[]StageField{}, groups, []StageSortField{}, filterTree}
}

func (stage *Stage) AddField(name, type_, contents string) {
	stage.fields = append(stage.fields, StageField{name, type_, contents})
}

func (stage *Stage) AddSortField(name string, direction StageSortFieldDir) {
	stage.sortFields = append(stage.sortFields, StageSortField{name, direction})
}

func (stage *Stage) FieldContents() []string {
	contents := []string{}
	for _, f := range stage.fields {
		contents = append(contents, f.contents)
	}
	return contents
}

func (stage *Stage) FiltersToSql(json bool) string {
	if stage.root == nil {
		return ""
	}
	return stage.root.ToSql(json)
}

func (stage *Stage) SortToSql(json bool) string {
	fields := make([]string, len(stage.sortFields))
	for i, field := range stage.sortFields {
		if json {
			fields[i] = FieldToSql(field.name, false)
		} else {
			fields[i] = `"` + field.name + `"`
		}
		if field.dir == SortDesc {
			fields[i] += " DESC"
		}
	}

	return strings.Join(fields, ", ")
}

func (stage *Stage) ToSql(table string, json bool) string {
	fields := "*"
	if len(stage.fields) > 0 {
		fields = strings.Join(stage.FieldContents(), ", ")
	}
	where := stage.FiltersToSql(json)
	if where != "" {
		where = " WHERE " + where
	}
	groupBy := ""
	if len(stage.groups) > 0 {
		groupBy = " GROUP BY " + strings.Join(stage.groups, ", ")
	}
	orderBy := ""
	if len(stage.sortFields) > 0 {
		// FIXME this is not working when we group then sort, json is true but the field is no longer json
		orderBy = " ORDER BY " + stage.SortToSql(json)
	}
	sql := "SELECT " + fields + " FROM " + table + where + groupBy + orderBy

	return sql
}

func (stage *Stage) GetValues() []interface{} {
	if stage.root == nil {
		return []interface{}{}
	}
	return stage.root.GetValues()
}

func (stage *Stage) FieldAsJsonBuilder() string {
	prefix := ""
	// if c.distinct != "" {
	// 	prefix = "DISTINCT ON (" + c.distinct + ") "
	// }
	str := fmt.Sprintf("%sjson_build_object('$k', jsonb_build_array(", prefix)

	for _, field := range stage.fields {
		str += fmt.Sprintf("'%s', ", field.name)
	}
	str = strings.TrimSuffix(str, ", ") + "), "

	for _, field := range stage.fields {
		if field.type_ == "float" {
			str += fmt.Sprintf("'%s', json_build_object('$f', %s), ", field.name, field.name)
		} else {
			str += fmt.Sprintf("'%s', %s, ", field.name, field.name)
		}
	}
	str = strings.TrimSuffix(str, ", ") + ") AS _jsonb"

	return str
}

func Wrap(table string, stages []*Stage) (string, []interface{}) {
	sql := ""
	queryValues := []interface{}{}
	for i, stage := range stages {
		queryValues = append(queryValues, stage.GetValues()...)
		from := table
		if sql != "" {
			from = fmt.Sprintf("("+sql+") AS query%v", i)
		}
		sql = stage.ToSql(from, i < 1)
	}

	var stage *Stage
	for _, s := range stages {
		if len(s.fields) > 0 {
			stage = s
		}
	}

	if stage == nil {
		return sql, queryValues
	}
	sql = "SELECT " + stage.FieldAsJsonBuilder() + " FROM (" + sql + ") AS wrapped"

	return sql, queryValues
}

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
)

func FieldToSql(field string, raw bool) string {
	if field == "" {
		return "_jsonb"
	}

	sql := "_jsonb"
	sep := "->"
	if raw {
		sep = "->>"
	}
	parts := strings.Split(field, ".")
	for i, f := range parts {
		if i == len(parts)-1 {
			sql += sep + `'` + f + `'`
		} else {
			sql += `->>'` + f + `'`
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
	return &FilterNode{}
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

func (node *FilterNode) ToSql() string {
	if len(node.children) > 0 {
		if node.unary {
			if len(node.children) > 1 {
				// FIXME re-evaluate this method of handling unary op errors
				panic("unary operator with multiple children: " + node.op)
			}
			return node.op + " (" + node.children[0].ToSql() + ")"
		}
		strs := make([]string, len(node.children))
		for i, child := range node.children {
			str := child.ToSql()
			strs[i] = str
		}
		return "(" + strings.Join(strs, " "+node.op+" ") + ")"
	}

	field := FieldToSql(node.field, node.raw)
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

func (node *FilterNode) AddFilter(index int, field string, op string, value interface{}) *FilterNode {
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
		values = append(values, node.value)
	}
	for _, child := range node.children {
		values = append(values, child.GetValues()...)
	}
	return values
}

type Stage struct {
	fields []string
	root   *FilterNode
}

func NewStage() Stage {
	return Stage{}
}

func (s *Stage) AddField(name string) {
	s.fields = append(s.fields, name)
}

func (s *Stage) GetValues() []interface{} {
	return s.root.GetValues()
}

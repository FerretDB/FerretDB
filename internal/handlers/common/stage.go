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

func FieldToSql(field string) string {
	sql := "_jsonb"
	parts := strings.Split(field, ".")
	for i, f := range parts {
		if i == len(parts)-1 {
			sql += `->'` + f + `'`
		} else {
			sql += `->>'` + f + `'`
		}
	}
	return sql
}

type StageFilter struct {
	index int
	field string
	op    string
	value interface{}
}

func (sf *StageFilter) ToSql() (string, error) {
	field := fmt.Sprintf("_jsonb->'%s'", sf.field)
	return fmt.Sprintf("%s %s $%v", field, sf.op, sf.index+1), nil
}

type Stage struct {
	fields  []string
	filters []StageFilter
}

func NewStage() Stage {
	return Stage{}
}

func (s *Stage) AddField(name string) {
	s.fields = append(s.fields, name)
}

func (s *Stage) AddFilter(field string, op string, value interface{}) {
	index := len(s.filters)
	s.filters = append(s.filters, StageFilter{index, field, op, value})
}

func (s *Stage) SetLastFilter(op string, value interface{}) {
	s.filters[len(s.filters)-1].op = op
	s.filters[len(s.filters)-1].value = value
}

func (s *Stage) GetFilters() []StageFilter {
	return s.filters
}

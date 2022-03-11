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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"golang.org/x/exp/slices"
)

// elemMatch
// SELECT json_build_object('$k', array['value'], 'value'::text, _jsonb->$3) FROM "values"."values" WHERE (_jsonb->'name' = to_jsonb('array-embedded'::text))
// Filter fields to return.
func (s *storage) projection(projection *types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	if projection == nil {
		sql = "_jsonb"
		return
	}
	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "_jsonb"
		return
	}

	ks, arg := s.buildProjectionKeys(projection.Keys(), projectionMap, p)
	args = append(args, arg...)

	// build json object
	sql = "json_build_object('$k', array[" + ks + "],"
	for i, k := range projection.Keys() { // value

		doc, isDoc := projectionMap[k].(*types.Document)
		// { field: 1}
		if !isDoc {
			if i != 0 {
				sql += ", "
			}
			sql += p.Next() + "::text, _jsonb->" + p.Next()
			args = append(args, k, k)
			continue
		}

		// field: { field: value } is not supported
		// field: { $elemMatch: { }}
		supportedKeys := []string{"$elemMatch"}
		for _, fieldKey := range doc.Keys() {
			if !slices.Contains(supportedKeys, fieldKey) {
				s.l.Sugar().Warnf("%s not supported", fieldKey)
				continue
			}

			fieldAny, err := doc.Get(fieldKey)
			if err != nil {
				panic("impossible code " + k + " > " + fieldKey)
			}

			fieldDoc, ok := fieldAny.(*types.Document)
			if !ok {
				panic("document expected " + k + "." + fieldKey)
			}

			switch fieldKey {
			case "$elemMatch":
				elemMatchSQL, arg := s.elemMatchProjection(k, fieldKey, fieldDoc, p)
				sql += elemMatchSQL
				args = append(args, arg...)
			default:
				panic("unsupported projection " + k + "." + fieldKey)
			}
		}

	}
	sql += ")"

	return
}

func (s *storage) elemMatchProjection(k, fieldKey string, elemMatchDoc *types.Document, p *pg.Placeholder) (elemMatchSQL string, arg []any) {
	elemMatchSQL = `
				CASE
				WHEN  jsonb_typeof(_jsonb->` + p.Next() + `) != 'array' THEN null
				ELSE
					(
						SELECT tempTable.value result
						FROM jsonb_array_elements(_jsonb->` + p.Next() + `) tempTable
						WHERE %s
						LIMIT 1
					)
				END val`
	arg = append(arg, k, k)

	// where part
	elemMatchWhere := ""
	elemMatchMap := elemMatchDoc.Map()
	for elemMatchKey, elemMatchVal := range elemMatchMap { // elemMatch field
		if elemMatchWhere != "" {
			elemMatchWhere += " AND "
		}

		filter, isDoc := elemMatchMap[elemMatchKey].(*types.Document)
		// field: scalar value
		if !isDoc {
			elemMatchWhere += "tempTable.value @? '$." + p.Next() + "[*] ? (@ == " + p.Next() + ")'"
			arg = append(arg, fieldKey, elemMatchVal)
			continue
		}

		// field: { $gt: scalar value}
		filterMap := filter.Map()
		for op, val := range filterMap {
			var operand string
			switch op {
			case "$eq":
				// {field: {$eq: value}}
				operand = "=="
			case "$ne":
				// {field: {$ne: value}}
				operand = "<>"
			case "$lt":
				// {field: {$lt: value}}
				operand = "<"
			case "$lte":
				// {field: {$lte: value}}
				operand = "<="
			case "$gt":
				// {field: {$gt: value}}
				operand = ">"
			case "$gte":
				// {field: {$gte: value}}
				operand = ">="
			}

			elemMatchWhere += "tempTable.value @? '$." + p.Next() + "[*] ? (@ " + operand + " " + p.Next() + ")'"
			arg = append(arg, k, elemMatchKey, val)
			s.l.Sugar().Debugf("$elemMatch field [%s] in %s", elemMatchKey, k)
		}

	}
	elemMatchSQL = fmt.Sprintf(elemMatchSQL, elemMatchWhere)
	return
}

// buildProjectionKeys prepares a key list with placeholders
func (s *storage) buildProjectionKeys(projectionKeys []string, projectionMap map[string]any, p *pg.Placeholder) (ks string, arg []any) {
	for i, k := range projectionKeys {
		doc, isDoc := projectionMap[k].(*types.Document)
		if isDoc {
			if elemMatchAny, err := doc.Get("$elemMatch"); err == nil {
				elemMatchDoc, ok := elemMatchAny.(*types.Document)
				if !ok {
					panic("expected $elemMatch to be doc")
				}
				for _, filterField := range elemMatchDoc.Keys() { // elemMatch field
					if i != 0 {
						ks += ", "
					}
					ks += p.Next()
					arg = append(arg, filterField)
					s.l.Sugar().Debugf("$elemMatch field [%s] in %s", filterField, k)
				}
				continue
			}
		}

		if i != 0 {
			ks += ", "
		}
		ks += p.Next()
		arg = append(arg, k)
	}
	return
}

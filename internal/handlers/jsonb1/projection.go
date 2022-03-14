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

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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

	// create a list of keys for document
	ks, arg, err := s.buildProjectionKeys(projection.Keys(), projectionMap, p)
	if err != nil {
		err = lazyerrors.Errorf("buildProjectionKeys: %w", err)
		return
	}
	args = append(args, arg...)
	sql = "json_build_object('$k', array[" + ks + "], "

	// _id and _id value
	sql += p.Next() + "::text, _jsonb->" + p.Next() + " "
	args = append(args, "_id", "_id")

	// build json object
	for _, k := range projection.Keys() { // value
		s.l.Sugar().Debugf("projection key %s", k)

		doc, isDoc := projectionMap[k].(*types.Document)
		// { field: 1 } to value->field
		if !isDoc {
			sql += ", "
			sql += p.Next() + "::text, _jsonb->" + p.Next()
			args = append(args, k, k)
			s.l.Sugar().Debugf("%s->%s", k, k)
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

			var fieldAny any
			fieldAny, err = doc.Get(fieldKey)
			if err != nil {
				err = lazyerrors.Errorf("impossible code %s.%s", k, fieldKey)
				return
			}

			fieldDoc, ok := fieldAny.(*types.Document)
			if !ok {
				err = lazyerrors.Errorf("projection: document expected at %s.%s", k, fieldKey)
				return
			}

			switch fieldKey {
			case "$elemMatch":
				var elemMatchSQL string
				elemMatchSQL, arg, err = s.buildProjectionQueryElemMatch(k, fieldDoc, p)
				if err != nil {
					err = lazyerrors.Errorf("buildProjectionQueryELemMatch: %s.%s %w", k, fieldKey, err)
					return
				}
				sql += ", " + elemMatchSQL
				args = append(args, arg...)
			default:
				err = lazyerrors.Errorf("unsupported projection %s.%s", k, fieldKey)
				return
			}
		}
	}
	sql += ")"

	return
}

func (s *storage) buildProjectionQueryElemMatch(k string, elemMatchDoc *types.Document, p *pg.Placeholder) (
	elemMatchSQL string, arg []any, err error,
) {
	s.l.Sugar().Debugf("field %s -> $elemMatch", k)

	elemMatchSQL = p.Next() + "::text, CASE WHEN jsonb_typeof(_jsonb->" + p.Next() + ") != 'array' THEN null " +
		"ELSE jsonb_build_array(( SELECT tempTable.value result FROM jsonb_array_elements(_jsonb->" + p.Next() +
		") tempTable WHERE %s LIMIT 1 )) END "
	arg = append(arg, k, k, k)

	// where part
	elemMatchWhere := ""
	elemMatchMap := elemMatchDoc.Map()
	for elemMatchKey, elemMatchVal := range elemMatchMap { // elemMatch field
		s.l.Sugar().Debugf("field %s -> $elemMatch -> %s", k, elemMatchKey)

		if elemMatchWhere != "" {
			elemMatchWhere += " AND "
		}

		filter, isDoc := elemMatchMap[elemMatchKey].(*types.Document)
		// field: scalar value
		if !isDoc {
			var val string
			val, err = pg.Sanitize(elemMatchVal)
			if err != nil {
				err = lazyerrors.Errorf("pg.Sanitize: %w", err)
				return
			}
			elemMatchWhere += "tempTable.value @? " + "'$." + elemMatchKey + "[*] ? (@ == " + val[1:len(val)-1] + ")'"
			s.l.Sugar().Debugf("field %s -> $elemMatch -> { %s: %v }", k, elemMatchKey, elemMatchVal)
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
			var conditionVal string
			conditionVal, err = pg.Sanitize(conditionVal)
			if err != nil {
				err = lazyerrors.Errorf("pg.Sanitize: %w", err)
				return
			}
			elemMatchWhere += "tempTable.value @? '$." + p.Next() + "[*] ? (@ " + operand + " " + conditionVal + ")'"
			arg = append(arg, elemMatchKey)
			s.l.Sugar().Debugf("field %s -> $elemMatch -> { %s %s %v }", k, elemMatchKey, operand, val)
		}
	}
	elemMatchSQL = fmt.Sprintf(elemMatchSQL, elemMatchWhere)
	return
}

// buildProjectionKeys prepares a key list with placeholders.
func (s *storage) buildProjectionKeys(projectionKeys []string, projectionMap map[string]any, p *pg.Placeholder) (
	ks string, arg []any, err error,
) {
	ks += p.Next()
	arg = append(arg, "_id")

	for _, k := range projectionKeys {
		doc, isDoc := projectionMap[k].(*types.Document)
		if !isDoc {
			ks += ", "
			ks += p.Next()
			arg = append(arg, k)
			continue
		}

		var elemMatchAny any
		if elemMatchAny, err = doc.Get("$elemMatch"); err == nil {
			elemMatchDoc, ok := elemMatchAny.(*types.Document)
			if !ok {
				err = fmt.Errorf("$elemMatch condition is not doc")
				return
			}
			for _, filterField := range elemMatchDoc.Keys() { // elemMatch field
				s.l.Sugar().Debugf("$elemMatch field [%s] in %s", filterField, k)
				if ks != "" {
					ks += ", "
				}
				ks += p.Next()
				arg = append(arg, k)
			}
			continue
		}
	}
	return
}

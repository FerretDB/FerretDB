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
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate executes an aggreagtion pipeline on documents in a collection
func (h *storage) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	pipeline := m["pipeline"].(*types.Array)
	db := m["$db"].(string)
	collection := m["aggregate"].(string)

	sql, args, err := aggregateToSQL(pipeline, db, collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	rows, err := h.pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var docs types.Array

	for {
		doc, err := nextRow(rows)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		if doc == nil {
			break
		}

		docs.Append(doc)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"cursor", types.MustNewDocument(
				"firstBatch", &docs,
				"id", int64(0), // TODO
				"ns", db+"."+collection,
			),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

func aggregateToSQL(pipeline *types.Array, db, collection string) (string, []any, error) {
	filterPairs := []any{}
	groupByFields := map[string]string{}
	sortByFields := map[string]any{}
	aggregations := map[string]string{}
	for i := 0; i < pipeline.Len(); i++ {
		item, err := pipeline.Get(i)
		if err != nil {
			return "", nil, lazyerrors.Error(err)
		}
		if p, ok := item.(*types.Document); ok {
			cmd := p.Command()
			switch cmd {
			case "$match":
				tfilter := p.Map()["$match"].(*types.Document)
				for k, v := range tfilter.Map() {
					filterPairs = append(filterPairs, k)
					filterPairs = append(filterPairs, v)
				}
			case "$sort":
				tsort := p.Map()["$sort"].(*types.Document)
				for k, v := range tsort.Map() {
					sortByFields[k] = v
				}
			case "$group":
				tgroup := p.Map()["$group"].(*types.Document)
				for k, v := range tgroup.Map() {
					if fieldAgg, ok := v.(*types.Document); ok {
						for agg, val := range fieldAgg.Map() {
							aggFormat := "count(%s)"
							aggValue := "*"
							if agg == "$sum" {
								aggFormat = "coalesce(sum(%s),0)"
								aggValue = "1"
							}
							if agg == "$sortByCount" {
								aggFormat = "count(%s)"
							}
							if sval, ok := val.(string); ok {
								aggValue = fmt.Sprintf("_jsonb->'%s')", sval)
							}
							aggregations[k] = fmt.Sprintf(aggFormat, aggValue)
						}
					}
					if fieldName, ok := v.(string); ok {
						groupByFields[k] = fmt.Sprintf("_jsonb->'%s'", strings.TrimPrefix(fieldName, "$"))
					}
				}
			}
		}
	}

	filter, err := types.NewDocument(filterPairs...)
	if err != nil {
		return "", nil, lazyerrors.Error(err)
	}

	groupByFieldItems := make([]string, 0, len(groupByFields))
	groupByClauseItems := make([]string, 0, len(groupByFields))
	for k, v := range groupByFields {
		groupByFieldItems = append(groupByFieldItems, fmt.Sprintf("%s as %s", v, k))
		groupByClauseItems = append(groupByClauseItems, v)
	}
	groupByFieldClause := strings.Join(groupByFieldItems, ",")
	groupByClause := strings.Join(groupByClauseItems, ",")

	aggItems := make([]string, 0, len(aggregations))
	for k, v := range aggregations {
		aggItems = append(aggItems, fmt.Sprintf("%s as %s", v, k))
	}
	aggClause := strings.Join(aggItems, ",")

	sortItems := make([]string, 0, len(sortByFields))
	for k := range sortByFields {
		sortItems = append(sortItems, k)
	}
	sortByClause := strings.Join(sortItems, ",")

	var args []any
	var placeholder pg.Placeholder

	whereSQL, args, err := where(filter, &placeholder)
	if err != nil {
		return "", nil, lazyerrors.Error(err)
	}

	internalSql := fmt.Sprintf(`SELECT %s FROM %s %s`, aggClause, pgx.Identifier{db, collection}.Sanitize(), whereSQL)
	if len(groupByFieldClause) > 0 {
		internalSql = fmt.Sprintf(`SELECT %s,%s FROM %s %s GROUP BY %s`, groupByFieldClause, aggClause, pgx.Identifier{db, collection}.Sanitize(), whereSQL, groupByClause)
	}
	if len(sortByClause) > 0 {
		internalSql += fmt.Sprintf(` ORDER BY %s`, sortByClause)
	}

	allFields := make([]string, 0, len(groupByFields)+len(aggregations))
	allKeyFields := make([]string, 0, len(groupByFields)+len(aggregations))
	for k := range groupByFields {
		allFields = append(allFields, fmt.Sprintf("'%s', agg.%s", k, k))
		allKeyFields = append(allKeyFields, fmt.Sprintf("'%s'", k))
	}
	for k := range aggregations {
		allFields = append(allFields, fmt.Sprintf("'%s', agg.%s", k, k))
		allKeyFields = append(allKeyFields, fmt.Sprintf("'%s'", k))
	}

	return fmt.Sprintf(`SELECT jsonb_build_object(%s,'$k',jsonb_build_array(%s)) FROM (%s) as agg`, strings.Join(allFields, ","), strings.Join(allKeyFields, ","), internalSql), args, nil
}

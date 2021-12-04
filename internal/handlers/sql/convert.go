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

package sql

import (
	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
)

type rowInfo struct {
	names []string
}

func extractRowInfo(rows pgx.Rows) *rowInfo {
	fields := rows.FieldDescriptions()
	ri := &rowInfo{
		names: make([]string, len(fields)),
	}

	// TODO get table info field.TableOID, check constraints, return a single-column PK as _id

	for i, field := range fields {
		ri.names[i] = string(field.Name)
	}

	return ri
}

func nextRow(rows pgx.Rows, rowInfo *rowInfo) (*types.Document, error) {
	if !rows.Next() {
		return nil, rows.Err()
	}

	values, err := rows.Values()
	if err != nil {
		return nil, err
	}

	pairs := make([]interface{}, len(values)*2)
	for i, v := range values {
		pairs[i*2] = rowInfo.names[i]
		pairs[i*2+1] = v
	}

	doc := types.MustMakeDocument(pairs...)
	return &doc, nil
}

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

package main

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Extract returns rows of routines and parameters for the given schema.
// Returned routines are sorted by the function/procedure name and its parameter position.
func Extract(ctx context.Context, uri string, schema string) []map[string]any {
	conn, err := pgx.Connect(ctx, uri)
	must.NoError(err)

	defer func() {
		must.NoError(conn.Close(ctx))
	}()

	// https://www.postgresql.org/docs/current/infoschema-routines.html
	// https://www.postgresql.org/docs/current/infoschema-parameters.html
	q := `
	SELECT
		specific_schema,
		specific_name,
		routine_name,
		routine_type,
		parameter_name,
		parameter_mode,
		parameter_default,
		p.data_type,
		p.udt_schema,
		p.udt_name,
		r.data_type AS routine_data_type,
		r.type_udt_schema AS routine_udt_schema,
		r.type_udt_name AS routine_udt_name
	FROM information_schema.routines AS r
		LEFT JOIN information_schema.parameters AS p USING (specific_schema, specific_name)
	WHERE specific_schema = $1
	ORDER BY specific_schema, specific_name, ordinal_position
	`

	return must.NotFail(pgx.CollectRows(must.NotFail(conn.Query(ctx, q, schema)), pgx.RowToMap))
}

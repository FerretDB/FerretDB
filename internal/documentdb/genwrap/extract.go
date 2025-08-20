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
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// Extract returns routines and parameters for given schemas.
// Keys are specific routines name (specific_schema.specific_name).
// Values are sorted by parameter position.
func Extract(ctx context.Context, uri string, schemas []string) (map[string][]map[string]any, error) {
	conn, err := pgx.Connect(ctx, uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer conn.Close(ctx)

	slices.Sort(schemas)
	schemas = slices.Compact(schemas)

	placeholders := make([]string, len(schemas))
	args := make([]any, len(schemas))

	for i, schema := range schemas {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = schema
	}

	// https://www.postgresql.org/docs/current/infoschema-routines.html
	// https://www.postgresql.org/docs/current/infoschema-parameters.html
	q := fmt.Sprintf(`
	SELECT
		specific_schema,
		specific_name,
		routine_name,
		routine_type,
		routines.data_type AS routine_data_type,
		routines.type_udt_schema AS routine_udt_schema,
		routines.type_udt_name AS routine_udt_name,
		parameter_name,
		parameter_mode,
		parameter_default,
		parameters.data_type,
		parameters.udt_schema,
		parameters.udt_name
	FROM information_schema.routines
		LEFT JOIN information_schema.parameters USING (specific_schema, specific_name)
	WHERE specific_schema IN (%s)
	ORDER BY specific_schema, specific_name, ordinal_position
	`,
		strings.Join(placeholders, ", "),
	)

	rows, err := conn.Query(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	rowsMap, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := map[string][]map[string]any{}
	var name string
	var params []map[string]any

	for _, r := range rowsMap {
		n := fmt.Sprintf("%s.%s", r["specific_schema"], r["specific_name"])
		if n != name {
			if name != "" {
				res[name] = params
			}

			name = n
			params = nil
		}

		params = append(params, r)
	}

	res[name] = params

	return res, nil
}

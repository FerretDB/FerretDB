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

// Package main contains code generator for DocumentDB APIs.
package main

import (
	"cmp"
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1148
func main() {
	opts := &logging.NewHandlerOpts{
		Base:          "console",
		Level:         slog.LevelDebug,
		CheckMessages: true,
	}
	logging.Setup(opts, "")

	l := slog.Default()
	ctx := context.Background()

	schemasF := flag.String("schemas", "", "comma-separated list of schemas")
	flag.Parse()

	if *schemasF == "" {
		l.Log(ctx, logging.LevelFatal, "-schemas flag is empty.")
	}

	// DOCUMENTDB_GEN_URL=postgres://username:password@127.0.0.1:5432/postgres
	connString := os.Getenv("DOCUMENTDB_GEN_URL")
	if connString == "" {
		l.InfoContext(ctx, "DOCUMENTDB_GEN_URL not set, skipping code generation.")
		os.Exit(0)
	}

	schemas := map[string]struct{}{}

	for _, schema := range strings.Split(*schemasF, ",") {
		schema = strings.TrimSpace(schema)
		if schema == "" {
			continue
		}

		must.NoError(os.RemoveAll(schema))
		must.NoError(os.MkdirAll(schema, 0o777))

		schemas[schema] = struct{}{}
	}

	conn, err := pgx.Connect(ctx, connString)
	must.NoError(err)

	defer conn.Close(ctx)

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
	ORDER BY specific_schema, specific_name, ordinal_position
	`

	rowsData := must.NotFail(pgx.CollectRows(must.NotFail(conn.Query(ctx, q)), pgx.RowToMap))

	for schema := range schemas {
		vs := Extract(rowsData, schema)

		fs := make([]*convertedRoutine, len(vs))
		for i, v := range vs {
			fs[i] = Convert(v)
		}

		out := must.NotFail(os.Create(filepath.Join(schema, schema+".go")))
		defer out.Close() //nolint:errcheck // ignore for now, but it should be checked

		h := headerData{
			Cmd:     "genwrap " + strings.Join(os.Args[1:], " "),
			Package: schema,
		}
		must.NoError(headerTemplate.Execute(out, &h))

		slices.SortFunc(fs, func(a, b *convertedRoutine) int {
			return cmp.Compare(a.Name, b.Name)
		})

		for _, f := range fs {
			must.NoError(Generate(out, f))
		}
	}
}

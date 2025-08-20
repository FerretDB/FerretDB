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
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestCase(t *testing.T) {
	t.Parallel()

	c := new(converter)

	assert.Equal(t, "cursorGetMore", c.camelCase("cursor_get_more"))
	assert.Equal(t, "CursorGetMore", c.pascalCase("cursor_get_more"))
}

func TestParameterName(t *testing.T) {
	t.Parallel()

	c := new(converter)

	assert.Equal(t, "validateSpec", c.parameterName("validatespec"))
}

func TestConvert2(t *testing.T) {
	t.Parallel()

	l := testutil.Logger(t)

	t.Skip("FIXME") // FIXME

	rows := map[string][]map[string]any{
		"count_query_19116": {
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "count_query_19116",
				"routine_name":       "count_query",
				"routine_type":       "FUNCTION",
				"parameter_name":     "database",
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "text",
				"udt_schema":         "pg_catalog",
				"udt_name":           "text",
				"routine_data_type":  "USER-DEFINED",
				"routine_udt_schema": "documentdb_core",
				"routine_udt_name":   "bson",
			},
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "count_query_19116",
				"routine_name":       "count_query",
				"routine_type":       "FUNCTION",
				"parameter_name":     "countspec",
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bson",
				"routine_data_type":  "USER-DEFINED",
				"routine_udt_schema": "documentdb_core",
				"routine_udt_name":   "bson",
			},
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "count_query_19116",
				"routine_name":       "count_query",
				"routine_type":       "FUNCTION",
				"parameter_name":     "document",
				"parameter_mode":     "OUT",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bson",
				"routine_data_type":  "USER-DEFINED",
				"routine_udt_schema": "documentdb_core",
				"routine_udt_name":   "bson",
			},
		},
		"drop_indexes_19097": {
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "drop_indexes_19097",
				"routine_name":       "drop_indexes",
				"routine_type":       "PROCEDURE",
				"parameter_name":     "p_database_name",
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "text",
				"udt_schema":         "pg_catalog",
				"udt_name":           "text",
				"routine_data_type":  nil,
				"routine_udt_schema": nil,
				"routine_udt_name":   nil,
			},
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "drop_indexes_19097",
				"routine_name":       "drop_indexes",
				"routine_type":       "PROCEDURE",
				"parameter_name":     "p_arg",
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bson",
				"routine_data_type":  nil,
				"routine_udt_schema": nil,
				"routine_udt_name":   nil,
			},
			{
				"specific_schema":    "documentdb_api",
				"specific_name":      "drop_indexes_19097",
				"routine_name":       "drop_indexes",
				"routine_type":       "PROCEDURE",
				"parameter_name":     "retval",
				"parameter_mode":     "INOUT",
				"parameter_default":  "NULL::documentdb_core.bson",
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bson",
				"routine_data_type":  nil,
				"routine_udt_schema": nil,
				"routine_udt_name":   nil,
			},
		},
		"bsonquery_compare_16444": {
			{
				"specific_schema":    "documentdb_core",
				"specific_name":      "bsonquery_compare_16444",
				"routine_name":       "bsonquery_compare",
				"routine_type":       "FUNCTION",
				"parameter_name":     nil,
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bsonquery",
				"routine_data_type":  "integer",
				"routine_udt_schema": "pg_catalog",
				"routine_udt_name":   "int4",
			},
			{
				"specific_schema":    "documentdb_core",
				"specific_name":      "bsonquery_compare_16444",
				"routine_name":       "bsonquery_compare",
				"routine_type":       "FUNCTION",
				"parameter_name":     nil,
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bsonquery",
				"routine_data_type":  "integer",
				"routine_udt_schema": "pg_catalog",
				"routine_udt_name":   "int4",
			},
		},
		"bsonquery_compare_16445": {
			{
				"specific_schema":    "documentdb_core",
				"specific_name":      "bsonquery_compare_16445",
				"routine_name":       "bsonquery_compare",
				"routine_type":       "FUNCTION",
				"parameter_name":     nil,
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bson",
				"routine_data_type":  "integer",
				"routine_udt_schema": "pg_catalog",
				"routine_udt_name":   "int4",
			},
			{
				"specific_schema":    "documentdb_core",
				"specific_name":      "bsonquery_compare_16445",
				"routine_name":       "bsonquery_compare",
				"routine_type":       "FUNCTION",
				"parameter_name":     nil,
				"parameter_mode":     "IN",
				"parameter_default":  nil,
				"data_type":          "USER-DEFINED",
				"udt_schema":         "documentdb_core",
				"udt_name":           "bsonquery",
				"routine_data_type":  "integer",
				"routine_udt_schema": "pg_catalog",
				"routine_udt_name":   "int4",
			},
		},
	}

	expected := map[string]map[string]templateData{
		"documentdb_api": {
			"binary_extended_version": {
				FuncName:    "BinaryExtendedVersion",
				SQLName:     "documentdb_api.binary_extended_version",
				SQLArgs:     "",
				SQLReturns:  "binary_extended_version",
				Comment:     `documentdb_api.binary_extended_version(OUT binary_extended_version text)`,
				FuncReturns: "outBinaryExtendedVersion string",
				ScanArgs:    "&outBinaryExtendedVersion",
			},
			"count_query": {
				FuncName:   "CountQuery",
				SQLName:    "documentdb_api.count_query",
				SQLArgs:    "$1, $2::bytea",
				SQLReturns: "document::bytea",
				Comment: `documentdb_api.count_query(database text, countspec documentdb_core.bson, ` +
					`OUT document documentdb_core.bson)`,
				FuncParams:   "database string, countSpec wirebson.RawDocument",
				FuncReturns:  "outDocument wirebson.RawDocument",
				QueryRowArgs: "database, countSpec",
				ScanArgs:     "&outDocument",
			},
		},
		"documentdb_core": {
			"bsonquery_compare": {
				FuncName:    "BsonqueryCompare",
				SQLName:     "documentdb_core.bsonquery_compare",
				IsProcedure: false,
				SQLArgs:     "$1, $2",
				SQLReturns:  "bsonquery_compare",
				Comment: `documentdb_core.bsonquery_compare(anonymous documentdb_core.bsonquery, ` +
					`anonymous1 documentdb_core.bsonquery, OUT bsonquery_compare integer)`,
				FuncParams:   "anonymous struct{}, anonymous1 struct{}",
				FuncReturns:  "outBsonqueryCompare int32",
				ScanArgs:     "&outBsonqueryCompare",
				QueryRowArgs: "anonymous, anonymous1",
			},
			"bsonquery_compare1": {
				FuncName:    "BsonqueryCompare1",
				SQLName:     "documentdb_core.bsonquery_compare",
				IsProcedure: false,
				SQLArgs:     "$1::bytea, $2",
				SQLReturns:  "bsonquery_compare",
				Comment: `documentdb_core.bsonquery_compare(anonymous documentdb_core.bson, ` +
					`anonymous1 documentdb_core.bsonquery, OUT bsonquery_compare integer)`,
				FuncParams:   "anonymous wirebson.RawDocument, anonymous1 struct{}",
				FuncReturns:  "outBsonqueryCompare int32",
				ScanArgs:     "&outBsonqueryCompare",
				QueryRowArgs: "anonymous, anonymous1",
			},
		},
	}

	res := Convert(rows, l)
	require.Equal(t, expected, res)
}

func TestGenerateGoFunction(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct { //nolint:vet // use only for testing
		data templateData

		res string
	}{
		"DropIndexes": {
			data: templateData{
				FuncName:    "DropIndexes",
				SQLName:     "documentdb_api.drop_indexes",
				IsProcedure: true,
				Comment: `documentdb_api.drop_indexes(p_database_name text, p_arg documentdb_core.bson, ` +
					`INOUT retval documentdb_core.bson DEFAULT NULL)`,
				SQLArgs:      "$1, $2::bytea, $3::bytea",
				SQLReturns:   "retval::bytea",
				FuncParams:   "databaseName string, arg wirebson.RawDocument, retValue wirebson.RawDocument",
				FuncReturns:  "outRetValue wirebson.RawDocument",
				QueryRowArgs: "databaseName, arg, retValue",
				ScanArgs:     "&outRetValue",
			},
			//nolint:lll // generated function is too long
			res: `
// DropIndexes is a wrapper for
//
//	documentdb_api.drop_indexes(p_database_name text, p_arg documentdb_core.bson, INOUT retval documentdb_core.bson DEFAULT NULL).
func DropIndexes(ctx context.Context, conn *pgx.Conn, l *slog.Logger, databaseName string, arg wirebson.RawDocument, retValue wirebson.RawDocument) (outRetValue wirebson.RawDocument, err error) {
	ctx, span := otel.Tracer("").Start(
		ctx,
		"DropIndexes",
		oteltrace.WithAttributes(
			otelsemconv.DBStoredProcedureName("documentdb_api.drop_indexes"),
			// TODO DBQuerySummaryKey
		),
	)
	defer span.End()

	row := conn.QueryRow(ctx, "CALL documentdb_api.drop_indexes($1, $2::bytea, $3::bytea)", databaseName, arg, retValue)
	if err = row.Scan(&outRetValue); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api.drop_indexes", l)
	}
	return
}
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			w := bufio.NewWriter(&b)
			err := generateGoFunction(w, &tc.data)
			require.NoError(t, err)

			err = w.Flush()
			require.NoError(t, err)
			require.Equal(t, tc.res, b.String())
		})
	}
}

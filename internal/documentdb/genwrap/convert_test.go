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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCamelCase(t *testing.T) {
	t.Parallel()

	c := new(Converter)

	require.Equal(t, "cursorGetMore", c.camelCase("cursor_get_more"))
}

func TestConvert(t *testing.T) {
	t.Parallel()

	rows := []map[string]any{
		{
			"specific_schema":    "documentdb_api",
			"specific_name":      "binary_extended_version_19132",
			"routine_name":       "binary_extended_version",
			"routine_type":       "FUNCTION",
			"parameter_name":     nil,
			"parameter_mode":     nil,
			"parameter_default":  nil,
			"data_type":          nil,
			"udt_schema":         nil,
			"udt_name":           nil,
			"routine_data_type":  "text",
			"routine_udt_schema": "pg_catalog",
			"routine_udt_name":   "text",
		},
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
	}

	expected := map[string]convertedRoutine{
		"binary_extended_version": {
			Name:         "BinaryExtendedVersion",
			SQLFuncName:  "documentdb_api.binary_extended_version",
			QueryArgs:    "",
			QueryReturns: "binary_extended_version",
			Comment:      `documentdb_api.binary_extended_version(OUT binary_extended_version text)`,
			GoReturns: []convertedRoutineParam{
				{
					Name: "outBinaryExtendedVersion",
					Type: "string",
				},
			},
		},
		"count_query": {
			Name:         "CountQuery",
			SQLFuncName:  "documentdb_api.count_query",
			QueryArgs:    "$1, $2::bytea",
			QueryReturns: "document::bytea",
			Comment: `documentdb_api.count_query(database text, countspec documentdb_core.bson, ` +
				`OUT document documentdb_core.bson)`,
			GoParams: []convertedRoutineParam{
				{
					Name: "database",
					Type: "string",
				},
				{
					Name: "countSpec",
					Type: "wirebson.RawDocument",
				},
			},
			GoReturns: []convertedRoutineParam{
				{
					Name: "outDocument",
					Type: "wirebson.RawDocument",
				},
			},
		},
		"drop_indexes": {
			Name:         "DropIndexes",
			SQLFuncName:  "documentdb_api.drop_indexes",
			IsProcedure:  true,
			QueryArgs:    "$1, $2::bytea, $3::bytea",
			QueryReturns: "retval::bytea",
			Comment: `documentdb_api.drop_indexes(p_database_name text, p_arg documentdb_core.bson, ` +
				`INOUT retval documentdb_core.bson DEFAULT NULL)`,
			GoParams: []convertedRoutineParam{
				{
					Name: "databaseName",
					Type: "string",
				},
				{
					Name: "arg",
					Type: "wirebson.RawDocument",
				},
				{
					Name: "retValue",
					Type: "wirebson.RawDocument",
				},
			},
			GoReturns: []convertedRoutineParam{
				{
					Name: "outRetValue",
					Type: "wirebson.RawDocument",
				},
			},
		},
	}

	res := Convert(rows)
	require.Equal(t, expected, res)
}

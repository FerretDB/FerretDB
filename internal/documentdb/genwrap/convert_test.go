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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param extractedRoutine
		res   convertedRoutine
	}{
		"CountQuery": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "count_query_18525",
				RoutineName:    "count_query",
				Params: []extractedRoutineParam{
					{
						ParameterName: pointer.ToString("database"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("text"),
						UDTSchema:     pointer.ToString("pg_catalog"),
						UDTName:       pointer.ToString("text"),
					},
					{
						ParameterName: pointer.ToString("countspec"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName: pointer.ToString("document"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
				},
			},
			res: convertedRoutine{
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
						Name: "document",
						Type: "wirebson.RawDocument",
					},
				},
			},
		},
		"BinaryExtendedVersion": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "binary_extended_version_18551",
				RoutineName:    "binary_extended_version",
				DataType:       pointer.ToString("text"),
				Params:         []extractedRoutineParam{},
			},
			res: convertedRoutine{
				Name:         "BinaryExtendedVersion",
				SQLFuncName:  "documentdb_api.binary_extended_version",
				QueryArgs:    "",
				QueryReturns: "binary_extended_version",
				Comment:      `documentdb_api.binary_extended_version(OUT binary_extended_version text)`,
				GoReturns: []convertedRoutineParam{
					{
						Name: "binaryExtendedVersion",
						Type: "string",
					},
				},
			},
		},
		"Insert": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "insert_18510",
				RoutineName:    "insert",
				Params: []extractedRoutineParam{
					{
						ParameterName: pointer.ToString("p_database_name"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("text"),
						UDTSchema:     pointer.ToString("pg_catalog"),
						UDTName:       pointer.ToString("text"),
					},
					{
						ParameterName: pointer.ToString("p_insert"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName:    pointer.ToString("p_insert_documents"),
						ParameterMode:    pointer.ToString("IN"),
						ParameterDefault: pointer.ToString("NULL::documentdb_core.bsonsequence"),
						DataType:         pointer.ToString("USER-DEFINED"),
						UDTSchema:        pointer.ToString("documentdb_core"),
						UDTName:          pointer.ToString("bsonsequence"),
					},
					{
						ParameterName:    pointer.ToString("p_transaction_id"),
						ParameterMode:    pointer.ToString("IN"),
						ParameterDefault: pointer.ToString("NULL::text"),
						DataType:         pointer.ToString("text"),
						UDTSchema:        pointer.ToString("pg_catalog"),
						UDTName:          pointer.ToString("text"),
					},
					{
						ParameterName: pointer.ToString("p_result"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName: pointer.ToString("p_success"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("boolean"),
						UDTSchema:     pointer.ToString("pg_catalog"),
						UDTName:       pointer.ToString("bool"),
					},
				},
			},
			res: convertedRoutine{
				Name:         "Insert",
				SQLFuncName:  "documentdb_api.insert",
				QueryArgs:    "$1, $2::bytea, $3::bytea",
				QueryReturns: "p_result::bytea, p_success",
				Comment: `documentdb_api.insert(p_database_name text, p_insert documentdb_core.bson, ` +
					`p_insert_documents documentdb_core.bsonsequence DEFAULT NULL, OUT p_result documentdb_core.bson, ` +
					`OUT p_success boolean)`,
				GoParams: []convertedRoutineParam{
					{
						Name: "databaseName",
						Type: "string",
					},
					{
						Name: "insert",
						Type: "wirebson.RawDocument",
					},
					{
						Name: "insertDocuments",
						Type: "[]byte",
					},
				},
				GoReturns: []convertedRoutineParam{
					{
						Name: "result",
						Type: "wirebson.RawDocument",
					},
					{
						Name: "success",
						Type: "bool",
					},
				},
			},
		},
		"CursorState": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "cursor_state_18585",
				RoutineName:    "cursor_state",
				DataType:       pointer.ToString("boolean"),
				Params: []extractedRoutineParam{
					{
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
				},
			},
			res: convertedRoutine{
				Name:         "CursorState",
				SQLFuncName:  "documentdb_api.cursor_state",
				QueryArgs:    "$1::bytea, $2::bytea",
				QueryReturns: "cursor_state",
				Comment: `documentdb_api.cursor_state(anonymous documentdb_core.bson, anonymous1 documentdb_core.bson, ` +
					`OUT cursor_state boolean)`,
				GoParams: []convertedRoutineParam{
					{
						Name: "anonymous",
						Type: "wirebson.RawDocument",
					},
					{
						Name: "anonymous1",
						Type: "wirebson.RawDocument",
					},
				},
				GoReturns: []convertedRoutineParam{
					{
						Name: "cursorState",
						Type: "bool",
					},
				},
			},
		},
		"EmptyDataTable": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "empty_data_table_18465",
				RoutineName:    "empty_data_table",
				Params: []extractedRoutineParam{
					{
						ParameterName: pointer.ToString("shard_key_value"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("bigint"),
						UDTSchema:     pointer.ToString("pg_catalog"),
						UDTName:       pointer.ToString("int8"),
					},
					{
						ParameterName: pointer.ToString("object_id"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName: pointer.ToString("document"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName: pointer.ToString("creation_time"),
						ParameterMode: pointer.ToString("OUT"),
						DataType:      pointer.ToString("timestamp with time zone"),
						UDTSchema:     pointer.ToString("pg_catalog"),
						UDTName:       pointer.ToString("timestamptz"),
					},
				},
			},
			res: convertedRoutine{
				Name:         "EmptyDataTable",
				SQLFuncName:  "documentdb_api.empty_data_table",
				QueryArgs:    "",
				QueryReturns: "shard_key_value, object_id::bytea, document::bytea, creation_time",
				Comment: `documentdb_api.empty_data_table(OUT shard_key_value bigint, OUT object_id documentdb_core.bson, ` +
					`OUT document documentdb_core.bson, OUT creation_time timestamp with time zone)`,
				GoReturns: []convertedRoutineParam{
					{
						Name: "shardKeyValue",
						Type: "int64",
					},
					{
						Name: "objectID",
						Type: "wirebson.RawDocument",
					},
					{
						Name: "document",
						Type: "wirebson.RawDocument",
					},
					{
						Name: "creationTime",
						Type: "[]byte",
					},
				},
			},
		},
		"CreateCollection": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "create_collection_18502",
				RoutineName:    "create_collection",
				DataType:       pointer.ToString("boolean"),
				Params: []extractedRoutineParam{
					{
						ParameterName: pointer.ToString("p_database_name"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("text"),
					},
					{
						ParameterName: pointer.ToString("p_collection_name"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("text"),
					},
				},
			},
			res: convertedRoutine{
				Name:         "CreateCollection",
				SQLFuncName:  "documentdb_api.create_collection",
				QueryArgs:    "$1, $2",
				QueryReturns: "create_collection",
				Comment: `documentdb_api.create_collection(p_database_name text, p_collection_name text, ` +
					`OUT create_collection boolean)`,
				GoParams: []convertedRoutineParam{
					{
						Name: "databaseName",
						Type: "string",
					},
					{
						Name: "collectionName",
						Type: "string",
					},
				},
				GoReturns: []convertedRoutineParam{
					{
						Name: "createCollection",
						Type: "bool",
					},
				},
			},
		},
		"DropIndexes": {
			param: extractedRoutine{
				SpecificSchema: "documentdb_api",
				SpecificName:   "drop_indexes_18587",
				RoutineName:    "drop_indexes",
				RoutineType:    "PROCEDURE",
				Params: []extractedRoutineParam{
					{
						ParameterName: pointer.ToString("p_database_name"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("text"),
					},
					{
						ParameterName: pointer.ToString("p_arg"),
						ParameterMode: pointer.ToString("IN"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
					{
						ParameterName: pointer.ToString("retval"),
						ParameterMode: pointer.ToString("INOUT"),
						DataType:      pointer.ToString("USER-DEFINED"),
						UDTSchema:     pointer.ToString("documentdb_core"),
						UDTName:       pointer.ToString("bson"),
					},
				},
			},
			res: convertedRoutine{
				Name:         "DropIndexes",
				SQLFuncName:  "documentdb_api.drop_indexes",
				IsProcedure:  true,
				QueryArgs:    "$1, $2::bytea, $3::bytea",
				QueryReturns: "",
				Comment: `documentdb_api.drop_indexes(p_database_name text, p_arg documentdb_core.bson, ` +
					`INOUT retval documentdb_core.bson)`,
				GoParams: []convertedRoutineParam{
					{
						Name: "databaseName",
						Type: "string",
					},
					{
						Name: "arg",
						Type: "wirebson.RawDocument",
					},
				},
				GoInOut: []convertedRoutineParam{
					{
						Name: "retValue",
						Type: "wirebson.RawDocument",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := Convert(&tc.param)
			require.Equal(t, &tc.res, res)
		})
	}
}

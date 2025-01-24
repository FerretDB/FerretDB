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

func TestExtract(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		schema string
		rows   []map[string]any

		res []*extractedRoutine
	}{
		"Multiple": {
			schema: "documentdb_api",
			rows: []map[string]any{
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "count_query_18525",
					"routine_name":    "count_query",
					"routine_type":    "FUNCTION",
					"parameter_name":  "database",
					"parameter_mode":  "IN",
					"data_type":       "USER-DEFINED",
				},
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "count_query_18525",
					"routine_name":    "count_query",
					"routine_type":    "FUNCTION",
					"parameter_name":  "countspec",
					"parameter_mode":  "IN",
					"data_type":       "USER-DEFINED",
				},
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "count_query_18525",
					"routine_name":    "count_query",
					"routine_type":    "FUNCTION",
					"parameter_name":  "document",
					"parameter_mode":  "OUT",
					"data_type":       "USER-DEFINED",
				},
				{
					"specific_schema":   "documentdb_api",
					"specific_name":     "binary_extended_version_18551",
					"routine_type":      "FUNCTION",
					"routine_name":      "binary_extended_version",
					"routine_data_type": "text",
				},
			},
			res: []*extractedRoutine{
				{
					SpecificSchema: "documentdb_api",
					SpecificName:   "count_query_18525",
					RoutineName:    "count_query",
					RoutineType:    "FUNCTION",
					Params: []extractedRoutineParam{
						{
							ParameterName: pointer.ToString("database"),
							ParameterMode: pointer.ToString("IN"),
							DataType:      pointer.ToString("USER-DEFINED"),
						},
						{
							ParameterName: pointer.ToString("countspec"),
							ParameterMode: pointer.ToString("IN"),
							DataType:      pointer.ToString("USER-DEFINED"),
						},
						{
							ParameterName: pointer.ToString("document"),
							ParameterMode: pointer.ToString("OUT"),
							DataType:      pointer.ToString("USER-DEFINED"),
						},
					},
				},
				{
					SpecificSchema: "documentdb_api",
					SpecificName:   "binary_extended_version_18551",
					RoutineName:    "binary_extended_version",
					DataType:       pointer.ToString("text"),
					RoutineType:    "FUNCTION",
					Params: []extractedRoutineParam{
						{},
					},
				},
			},
		},
		"DropIndexes": {
			schema: "documentdb_api",
			rows: []map[string]any{
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "drop_indexes_18587",
					"routine_name":    "drop_indexes",
					"routine_type":    "PROCEDURE",
					"parameter_name":  "p_database_name",
					"parameter_mode":  "IN",
					"data_type":       "text",
				},
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "drop_indexes_18587",
					"routine_name":    "drop_indexes",
					"routine_type":    "PROCEDURE",
					"parameter_name":  "p_arg",
					"parameter_mode":  "IN",
					"data_type":       "USER-DEFINED",
				},
				{
					"specific_schema": "documentdb_api",
					"specific_name":   "drop_indexes_18587",
					"routine_name":    "drop_indexes",
					"routine_type":    "PROCEDURE",
					"parameter_name":  "retval",
					"parameter_mode":  "INOUT",
					"data_type":       "USER-DEFINED",
				},
			},
			res: []*extractedRoutine{
				{
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
						},
						{
							ParameterName: pointer.ToString("retval"),
							ParameterMode: pointer.ToString("INOUT"),
							DataType:      pointer.ToString("USER-DEFINED"),
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := Extract(tc.rows, tc.schema)
			require.Equal(t, tc.res, res)
		})
	}
}

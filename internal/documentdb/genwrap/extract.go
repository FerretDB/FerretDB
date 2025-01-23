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
	"log"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// extractedRoutine is SQL function or procedure information fetched from the database.
type extractedRoutine struct {
	SpecificSchema string
	SpecificName   string // unique identifier of function in case of overloading
	RoutineName    string // name of the function
	RoutineType    string
	DataType       *string
	UDTSchema      *string
	UDTName        *string
	Params         []extractedRoutineParam
}

// toDataType returns SQL datatype. If the data type is USER-DEFINED,
// it returns schema and name concatenated by dot.
func (r *extractedRoutine) toDataType() string {
	dataType := *r.DataType
	if *r.DataType == "USER-DEFINED" {
		dataType = *r.UDTSchema + "." + *r.UDTName
	}

	return dataType
}

// extractedRoutineParam is a parameter of a routine fetched from the database.
type extractedRoutineParam struct {
	ParameterName    *string
	ParameterMode    *string
	ParameterDefault *string
	DataType         *string
	UDTSchema        *string
	UDTName          *string
}

// toDataType returns SQL datatype. If the data type is USER-DEFINED,
// it returns schema and name concatenated by dot.
func (p *extractedRoutineParam) toDataType() string {
	dataType := *p.DataType
	if *p.DataType == "USER-DEFINED" {
		dataType = *p.UDTSchema + "." + *p.UDTName
	}

	return dataType
}

// toDefault returns default SQL value if any.
func (p *extractedRoutineParam) toDefault() string {
	var defaultValue string

	if p.ParameterDefault != nil {
		d, _, _ := strings.Cut(*p.ParameterDefault, "::")
		defaultValue = " DEFAULT " + d
	}

	return defaultValue
}

// Extract takes rows and the schema and returns a list routines for the given schema.
//
// The routines that do not belong to the schema is ignored.
func Extract(rows []map[string]any, schema string) []*extractedRoutine {
	routines := map[string]*extractedRoutine{}
	params := map[string][]extractedRoutineParam{}

	for _, row := range rows {
		specificSchema := row["specific_schema"].(string)
		if schema != specificSchema {
			continue
		}

		specificName := row["specific_name"].(string)

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1148
		if strings.Contains(strings.ToLower(specificName), "shard") {
			log.Printf("Skipping %q", specificName)
			continue
		}

		if _, ok := routines[specificName]; !ok {
			var routineType string
			if typ := row["routine_type"]; typ != nil {
				routineType = typ.(string)
			}

			routines[specificName] = &extractedRoutine{
				SpecificSchema: specificSchema,
				RoutineName:    row["routine_name"].(string),
				SpecificName:   row["specific_name"].(string),
				DataType:       getStringP(row["routine_data_type"]),
				UDTSchema:      getStringP(row["routine_udt_schema"]),
				UDTName:        getStringP(row["routine_udt_name"]),
				RoutineType:    routineType,
			}
		}

		params[specificName] = append(params[specificName], extractedRoutineParam{
			ParameterName:    getStringP(row["parameter_name"]),
			ParameterMode:    getStringP(row["parameter_mode"]),
			ParameterDefault: getStringP(row["parameter_default"]),
			DataType:         getStringP(row["data_type"]),
			UDTSchema:        getStringP(row["udt_schema"]),
			UDTName:          getStringP(row["udt_name"]),
		})
	}

	for _, specificName := range maps.Keys(routines) {
		r := routines[specificName]
		r.Params = params[r.SpecificName]
	}

	values := maps.Values(routines)

	sort.SliceStable(values, func(i, j int) bool {
		return values[i].SpecificName > values[j].SpecificName
	})

	return values
}

// getStringP gets string pointer type from the given value.
func getStringP(v any) *string {
	if v == nil {
		return nil
	}

	s := v.(string)

	return &s
}

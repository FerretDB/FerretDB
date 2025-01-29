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
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// Converter converts SQL data types to Go data types.
type Converter struct {
	l *slog.Logger
}

// Convert takes rows containing parameters of routines.
// It returns a map of schemas and routines belonging to each schema by
// converting rows to Go formatted names and types.
//
// For an anonymous SQL parameter, it assigns a unique name.
// It also produces SQL query placeholders and return parameters in strings.
func Convert(rows []map[string]any, l *slog.Logger) map[string]map[string]templateData {
	c := &Converter{
		l: l,
	}

	routineParams := c.groupBySpecificName(rows)
	schemas := map[string]map[string]templateData{}

	for _, specificName := range slices.Sorted(maps.Keys(routineParams)) {
		params := routineParams[specificName]

		var goParams, goReturns, sqlParams, sqlReturns, comment, queryRowArgs, scanArgs, paramNames []string
		var placeholderCounter int

		for _, row := range params {
			name := "anonymous"

			if row["parameter_name"] != nil {
				name = row["parameter_name"].(string)
			}

			if row["parameter_mode"] == "IN" {
				name = c.uniqueName(paramNames, name)
			}

			paramNames = append(paramNames, name)

			if row["parameter_name"] == "p_transaction_id" {
				// skip p_transaction_id, transaction is not supported yet
				// TODO https://github.com/FerretDB/FerretDB/issues/8
				continue
			}

			if row["parameter_name"] == nil {
				// skip a row if the row does not contain a parameter such as BinaryExtendedVersion()
				continue
			}

			comment = append(comment, c.toParamComment(name, row))
			dataType := c.dataType(row)

			if row["parameter_mode"] == "IN" || row["parameter_mode"] == "INOUT" {
				placeholder := fmt.Sprintf("$%d", placeholderCounter+1)
				placeholderCounter++

				goName := c.convertName(name)
				sqlParams = append(sqlParams, c.convertEncodedType(placeholder, dataType))
				goParams = append(goParams, fmt.Sprintf("%s %s", goName, c.convertType(dataType)))
				queryRowArgs = append(queryRowArgs, goName)
			}

			if row["parameter_mode"] == "OUT" || row["parameter_mode"] == "INOUT" {
				goName := "out" + c.pascalCase(c.convertName(name))
				sqlReturns = append(sqlReturns, c.convertEncodedType(name, dataType))
				goReturns = append(goReturns, fmt.Sprintf("%s %s", goName, c.convertType(dataType)))
				scanArgs = append(scanArgs, fmt.Sprintf("&%s", goName))
			}
		}

		routineName := params[0]["routine_name"].(string)

		if len(goReturns) == 0 && params[0]["routine_type"] == "FUNCTION" {
			// function such as binary_extended_version() does not have
			// parameter data type, but it has routine data type for the return variable.
			goName := "out" + c.pascalCase(c.convertName(routineName))
			dataType := c.routineDataType(params[0])
			sqlReturns = append(sqlReturns, c.convertEncodedType(routineName, dataType))
			goReturns = append(goReturns, fmt.Sprintf("%s %s", goName, c.convertType(dataType)))
			scanArgs = append(scanArgs, fmt.Sprintf("&%s", goName))
			comment = append(comment, fmt.Sprintf("OUT %s %s", routineName, dataType))
		}

		schema := params[0]["specific_schema"].(string)

		if _, ok := schemas[schema]; !ok {
			schemas[schema] = map[string]templateData{}
		}

		// unique name is used to handle function overloading
		uniqueFunctionName := c.uniqueName(slices.Collect(maps.Keys(schemas[schema])), routineName)

		r := templateData{
			FuncName:     c.pascalCase(uniqueFunctionName),
			SQLFuncName:  fmt.Sprintf("%s.%s", schema, routineName),
			Comment:      fmt.Sprintf("%s.%s(%s)", schema, routineName, strings.Join(comment, ", ")),
			IsProcedure:  params[0]["routine_type"] == "PROCEDURE",
			SQLArgs:      strings.Join(sqlParams, ", "),
			SQLReturns:   strings.Join(sqlReturns, ", "),
			Params:       strings.Join(goParams, ", "),
			Returns:      strings.Join(goReturns, ", "),
			QueryRowArgs: strings.Join(queryRowArgs, ", "),
			ScanArgs:     strings.Join(scanArgs, ", "),
		}

		c.handleFunctionOverloading(&r)

		schemas[schema][uniqueFunctionName] = r
	}

	return schemas
}

// convertEncodedType appends binary data encoding to documentdb_core.bson data type.
func (c *Converter) convertEncodedType(parameter string, dataType string) string {
	res := parameter

	switch dataType {
	case "documentdb_core.bson", "documentdb_core.bsonsequence":
		res += "::bytea"
	}

	return res
}

// camelCase converts a string to camelCase.
func (c *Converter) camelCase(s string) string {
	var nextCapital bool
	var out []byte

	for _, ch := range s {
		if ch == '_' {
			nextCapital = true
			continue
		}

		if nextCapital {
			ch = unicode.ToUpper(ch)
			nextCapital = false
		}

		out = append(out, byte(ch))
	}

	return string(out)
}

// convertType converts DocumentDB and PostgreSQL types to Go types.
func (c *Converter) convertType(typ string) string {
	switch typ {
	case "ARRAY":
		return "[]any"
	case "text", "cstring":
		return "string"
	case "documentdb_core.bson", "documentdb_core.bsonquery":
		return "wirebson.RawDocument"
	case "documentdb_core.bsonsequence", "timestamp with time zone", "uuid", "bytea":
		return "[]byte"
	case "boolean":
		return "bool"
	case `"any"`, "anyelement", "documentdb_api_catalog.index_spec_type", "internal", "index_am_handler", "oid",
		"public.vector", "record", "regclass", "trigger", "tsquery", "void":
		c.l.Debug("use string for PostgreSQL/DocumentDB type without known conversion", slog.String("data_type", typ))
		return "string"
	case "smallint":
		return "int16"
	case "integer":
		return "int32"
	case "bigint":
		return "int64"
	case "double precision":
		return "float64"
	default:
		return typ
	}
}

// convertName converts to golang friendly formatted name.
func (c *Converter) convertName(name string) string {
	var found bool
	if name, found = strings.CutPrefix(name, "p_"); found {
		return c.camelCase(name)
	}

	switch name {
	case "ok", "document", "requests", "shard_key_value", "creation_time", "complete",
		"binary_version", "binary_extended_version", "create_collection":
		return c.camelCase(name)

	case "dbname":
		return "databaseName"
	case "database":
		return "database"
	case "commandspec":
		return "commandSpec"
	case "countspec":
		return "countSpec"
	case "cursorid":
		return "cursorID"
	case "object_id":
		return "objectID"
	case "cursorpage":
		return "cursorPage"

	case "continuation":
		return "continuation"

	case "retval":
		return "retValue"

	case "persistconnection":
		return "persistConnection"
	case "createspec":
		return "createSpec"
	case "distinctspec":
		return "distinctSpec"
	case "getmorespec":
		return "getMoreSpec"
	case "continuationspec":
		return "continuationSpec"

	default:
		return c.camelCase(name)
	}
}

// pascalCase converts a string to PascalCase.
func (c *Converter) pascalCase(s string) string {
	strArr := []rune(c.camelCase(s))

	strArr[0] = unicode.ToUpper(strArr[0])

	return string(strArr)
}

// handleFunctionOverloading applies different wrapper function name for overloaded functions.
func (c *Converter) handleFunctionOverloading(f *templateData) {
	var funcName string

	var skip bool

	// handle Bson duplicates
	switch f.SQLFuncName {
	case "documentdb_core.bsonquery_compare":
		funcName = "BsonQueryCompareBson"
	default:
		skip = true
	}

	if !skip && strings.Contains(f.Comment, "documentdb_core.bson,") {
		f.FuncName = funcName
	}
}

// groupBySpecificName groups rows by specific_name.
func (c *Converter) groupBySpecificName(rows []map[string]any) map[string][]map[string]any {
	var specificName any

	routines := map[string][]map[string]any{}
	var groupedParams []map[string]any

	for _, row := range rows {
		if specificName != row["specific_name"] && specificName != nil {
			routines[specificName.(string)] = groupedParams
			groupedParams = []map[string]any{}
		}

		groupedParams = append(groupedParams, row)
		specificName = row["specific_name"]
	}

	routines[specificName.(string)] = groupedParams

	return routines
}

// dataType returns SQL datatype of a parameter. If the data type is USER-DEFINED,
// it returns schema and name concatenated by dot.
func (c *Converter) dataType(row map[string]any) string {
	if row["data_type"] == "USER-DEFINED" {
		return row["udt_schema"].(string) + "." + row["udt_name"].(string)
	}

	return row["data_type"].(string)
}

// routineDataType returns SQL datatype of a routine. If the data type is USER-DEFINED,
// it returns schema and name concatenated by dot.
func (c *Converter) routineDataType(row map[string]any) string {
	if row["routine_data_type"] == "USER-DEFINED" {
		return row["routine_udt_schema"].(string) + "." + row["routine_udt_name"].(string)
	}

	return row["routine_data_type"].(string)
}

// toParamComment returns concatenated string of parameter name, data type
// and default value to use for the parameter description of a function.
// If the parameter is not an input, prefix OUT/INOUT is added to the comment.
func (c *Converter) toParamComment(paramName string, row map[string]any) string {
	comment := paramName + " " + c.dataType(row)
	if row["parameter_mode"] != "IN" {
		comment = row["parameter_mode"].(string) + " " + comment
	}

	if row["parameter_default"] != nil {
		d, _, _ := strings.Cut(row["parameter_default"].(string), "::")
		comment += " DEFAULT " + d
	}

	return comment
}

// uniqueName generates a new name if it exists in names slice.
func (c *Converter) uniqueName(names []string, name string) string {
	i := 1
	for slices.Contains(names, name) {
		name = fmt.Sprintf("%s%d", name, i)
		i++
	}

	return name
}

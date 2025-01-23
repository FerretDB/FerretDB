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
	"log"
	"strings"
	"unicode"
)

// convertedRoutineParam represents name and type of Go parameter.
type convertedRoutineParam struct {
	Name string
	Type string
}

// convertedRoutine contains function/procedure and parameter information converted
// from SQL to Go formatted names and types.
type convertedRoutine struct {
	Name         string
	SQLFuncName  string
	QueryArgs    string
	QueryReturns string
	Comment      string
	GoParams     []convertedRoutineParam
	GoReturns    []convertedRoutineParam
	GoInOut      []convertedRoutineParam
	IsProcedure  bool
}

// Convert takes information schema routine and converts SQL data type to Go data type.
// For an anonymous SQL parameter, it assigns a unique name.
// It also produces SQL query placeholders and return parameters in strings.
func Convert(r *extractedRoutine) *convertedRoutine {
	f := convertedRoutine{
		Name:        pascalCase(r.RoutineName),
		SQLFuncName: fmt.Sprintf("%s.%s", r.SpecificSchema, r.RoutineName),
		IsProcedure: r.RoutineType == "PROCEDURE",
	}

	var comment []string
	var sqlParams []string
	var sqlReturns []string

	uniqueNameChecker := map[string]struct{}{}
	var placeholderCounter int

	for _, p := range r.Params {
		if p.DataType == nil {
			continue
		}

		name := "anonymous"
		if p.ParameterName != nil {
			name = *p.ParameterName
		}

		if name == "p_transaction_id" {
			continue
		}

		if *p.ParameterMode == "IN" {
			// assign unique name for parameters such as
			// cursor_state(anonymous documentdb_core.bson, anonymous1 documentdb_core.bson)
			if _, found := uniqueNameChecker[name]; found {
				i := 1

				for {
					newName := fmt.Sprintf("%s%d", name, i)
					if _, found := uniqueNameChecker[newName]; !found {
						name = newName
						break
					}

					i++
				}
			}
			uniqueNameChecker[name] = struct{}{}
		}

		commentParam := name + " " + p.toDataType()
		if *p.ParameterMode != "IN" {
			commentParam = *p.ParameterMode + " " + commentParam
		}

		if p.ParameterDefault != nil {
			commentParam += p.toDefault()
		}
		comment = append(comment, commentParam)

		gp := convertedRoutineParam{
			Name: convertName(name),
			Type: convertType(p.toDataType()),
		}

		switch *p.ParameterMode {
		case "IN":
			placeholder := fmt.Sprintf("$%d", placeholderCounter+1)
			sqlParams = append(sqlParams, convertEncodedType(placeholder, p.toDataType()))
			placeholderCounter++

			f.GoParams = append(f.GoParams, gp)
		case "OUT":
			sqlReturns = append(sqlReturns, convertEncodedType(name, p.toDataType()))
			f.GoReturns = append(f.GoReturns, gp)
		case "INOUT":
			// INOUT parameter appears for SQL procedure
			placeholder := fmt.Sprintf("$%d", placeholderCounter+1)
			sqlParams = append(sqlParams, convertEncodedType(placeholder, p.toDataType()))
			placeholderCounter++

			f.GoInOut = append(f.GoInOut, gp)
		default:
			log.Printf("unrecognized parameter mode: %s", *p.ParameterMode)
		}
	}

	if len(f.GoReturns) == 0 && !f.IsProcedure {
		// function such as binary_extended_version() does not have
		// parameter data type, but it has routine data type for the return variable.
		sqlReturns = append(sqlReturns, convertEncodedType(r.RoutineName, r.toDataType()))

		f.GoReturns = []convertedRoutineParam{{
			Name: convertName(r.RoutineName),
			Type: convertType(r.toDataType()),
		}}

		comment = append(comment, "OUT "+r.RoutineName+" "+r.toDataType())
	}

	f.QueryArgs = strings.Join(sqlParams, ", ")
	f.QueryReturns = strings.Join(sqlReturns, ", ")
	f.Comment = fmt.Sprintf("%s.%s(%s)", r.SpecificSchema, r.RoutineName, strings.Join(comment, ", "))

	handleFunctionOverloading(&f)

	return &f
}

// convertEncodedType appends binary data encoding to documentdb_core.bson data type.
func convertEncodedType(parameter string, dataType string) string {
	res := parameter

	switch dataType {
	case "documentdb_core.bson", "documentdb_core.bsonsequence":
		res += "::bytea"
	}

	return res
}

// camelCase converts a string to camelCase.
func camelCase(s string) string {
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
func convertType(typ string) string {
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
	case `"any"`, "anyelement", "internal", "index_am_handler", "oid", "record", "void":
		// use string for PostgreSQL/DocumentDB type we do not know how to convert
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
func convertName(name string) string {
	var found bool
	if name, found = strings.CutPrefix(name, "p_"); found {
		return camelCase(name)
	}

	switch name {
	case "ok", "document", "requests", "shard_key_value", "creation_time", "complete",
		"binary_version", "binary_extended_version", "create_collection":
		return camelCase(name)

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
		return camelCase(name)
	}
}

// pascalCase converts a string to PascalCase.
func pascalCase(s string) string {
	strArr := []rune(camelCase(s))

	strArr[0] = unicode.ToUpper(strArr[0])

	return string(strArr)
}

// handleFunctionOverloading applies different wrapper function name for overloaded functions.
func handleFunctionOverloading(f *convertedRoutine) {
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
		f.Name = funcName
	}
}

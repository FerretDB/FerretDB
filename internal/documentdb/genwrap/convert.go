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
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// Convert takes rows containing parameters of routines.
// It returns a map of schemas and routines belonging to each schema by
// converting rows to Go formatted names and types.
//
// For an anonymous SQL parameter, it assigns a unique name.
// It also produces SQL query placeholders and return parameters in strings.
func Convert(routineParams map[string][]map[string]any, l *slog.Logger) map[string]map[string]templateData {
	c := &converter{
		l: l,
	}

	schemas := map[string]map[string]templateData{}

	for _, specificName := range slices.Sorted(maps.Keys(routineParams)) {
		params := routineParams[specificName]

		var goParams, goReturns, sqlParams, sqlReturns, comment, queryRowArgs, scanArgs, paramNames []string
		var placeholderCounter int

		for _, row := range params {
			pn, gp, gr, sp, sr, cm, qra, sa, pc := c.convertParam(
				row,
				paramNames,
				placeholderCounter,
			)
			if pn == "" && gp == "" && gr == "" && sp == "" && sr == "" && cm == "" && qra == "" && sa == "" {
				// skip row
				continue
			}
			paramNames = append(paramNames, pn)
			if gp != "" {
				goParams = append(goParams, gp)
			}
			if gr != "" {
				goReturns = append(goReturns, gr)
			}
			if sp != "" {
				sqlParams = append(sqlParams, sp)
			}
			if sr != "" {
				sqlReturns = append(sqlReturns, sr)
			}
			if cm != "" {
				comment = append(comment, cm)
			}
			if qra != "" {
				queryRowArgs = append(queryRowArgs, qra)
			}
			if sa != "" {
				scanArgs = append(scanArgs, sa)
			}
			placeholderCounter = pc
		}

		routineName := params[0]["routine_name"].(string)

		if len(goReturns) == 0 && params[0]["routine_type"] == "FUNCTION" && params[0]["routine_data_type"] != "void" {
			// function such as binary_extended_version() does not have
			// parameter data type, but it has routine data type for the return variable,
			// except void data type which does not return anything.
			goName := "out" + c.pascalCase(c.parameterName(routineName))
			dataType := params[0]["routine_data_type"].(string)
			sqlReturns = append(sqlReturns, c.parameterCast(routineName, dataType))
			goReturns = append(goReturns, fmt.Sprintf("%s %s", goName, c.parameterType(dataType)))
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
			SQLName:      fmt.Sprintf("%s.%s", schema, routineName),
			Comment:      fmt.Sprintf("%s.%s(%s)", schema, routineName, strings.Join(comment, ", ")),
			IsProcedure:  params[0]["routine_type"] == "PROCEDURE",
			SQLArgs:      strings.Join(sqlParams, ", "),
			SQLReturns:   strings.Join(sqlReturns, ", "),
			FuncParams:   strings.Join(goParams, ", "),
			FuncReturns:  strings.Join(goReturns, ", "),
			QueryRowArgs: strings.Join(queryRowArgs, ", "),
			ScanArgs:     strings.Join(scanArgs, ", "),
		}

		// TODO https://github.com/documentdb/documentdb/issues/49
		if r.IsProcedure {
			l.Warn("Procedure skipped due to unable to decode output", slog.String("function", r.SQLName))

			continue
		}

		schemas[schema][uniqueFunctionName] = r
	}

	return schemas
}

// converter is used to group methods used by [Convert].
type converter struct {
	l *slog.Logger
}

// convertParam processes a single parameter row and updates the provided slices accordingly.
func (c *converter) convertParam(
	row map[string]any,
	paramNames []string,
	placeholderCounter int,
) (
	string, // paramName
	string, // goParam
	string, // goReturn
	string, // sqlParam
	string, // sqlReturn
	string, // comment
	string, // queryRowArg
	string, // scanArg
	int, // placeholderCounter
) {
	name := "anonymous"

	if row["parameter_name"] != nil {
		name = row["parameter_name"].(string)
	}

	if row["parameter_mode"] == "IN" {
		name = c.uniqueName(paramNames, name)
	}

	if row["parameter_name"] == "p_transaction_id" {
		return "", "", "", "", "", "", "", "", placeholderCounter
	}

	if row["parameter_mode"] == nil {
		return "", "", "", "", "", "", "", "", placeholderCounter
	}

	comment := c.toParamComment(name, row)
	dataType := row["data_type"].(string)

	var goParam, goReturn, sqlParam, sqlReturn, queryRowArg, scanArg string

	if row["parameter_mode"] == "IN" || row["parameter_mode"] == "INOUT" {
		placeholder := fmt.Sprintf("$%d", placeholderCounter+1)
		placeholderCounter++

		goName := c.parameterName(name)
		sqlParam = c.parameterCast(placeholder, dataType)
		goParam = fmt.Sprintf("%s %s", goName, c.parameterType(dataType))
		queryRowArg = goName
	}

	if row["parameter_mode"] == "OUT" || row["parameter_mode"] == "INOUT" {
		goName := "out" + c.pascalCase(c.parameterName(name))
		sqlReturn = c.parameterCast(name, dataType)
		goReturn = fmt.Sprintf("%s %s", goName, c.parameterType(dataType))
		scanArg = fmt.Sprintf("&%s", goName)
	}

	return name, goParam, goReturn, sqlParam, sqlReturn, comment, queryRowArg, scanArg, placeholderCounter
}

// camelCase converts snake_case to to camelCase.
func (c *converter) camelCase(s string) string {
	var res []rune

	var u bool
	for _, ch := range s {
		if ch == '_' {
			u = true
			continue
		}

		if u {
			ch = unicode.ToUpper(ch)
			u = false
		}

		res = append(res, ch)
	}

	return string(res)
}

// pascalCase converts snake_case to PascalCase.
func (c *converter) pascalCase(s string) string {
	res := []rune(c.camelCase(s))
	res[0] = unicode.ToUpper(res[0])

	return string(res)
}

// parameterName converts PostgreSQL/DocumentDB routine parameter name
// to Go function/method parameter name.
func (c *converter) parameterName(name string) string {
	name = strings.TrimPrefix(name, "p_")

	switch name {
	case "dbname":
		return "database"

	case "getmorespec":
		return "getMoreSpec"
	case "letvariablespec":
		return "letVariableSpec"

	case "cursorid":
		return "cursorID"
	case "cursorpage":
		return "cursorPage"
	case "object_id":
		return "objectID"
	case "persistconnection":
		return "persistConnection"
	case "retval":
		return "retVal"
	}

	if name != "spec" && strings.HasSuffix(name, "spec") {
		name = strings.TrimSuffix(name, "spec") + "_spec"
	}

	return c.camelCase(name)
}

// parameterType converts PostgreSQL/DocumentDB routine parameter type
// to Go function/method parameter type.
func (c *converter) parameterType(typ string) string {
	switch typ {
	case "text":
		return "string"
	case "boolean":
		return "bool"
	case "integer":
		return "int32"
	case "bigint":
		return "int64"
	case "double precision":
		return "float64"
	case "uuid":
		return "[]byte"

	case "documentdb_core.bson":
		return "wirebson.RawDocument"
	case "documentdb_core.bsonsequence":
		return "[]byte"
	}

	c.l.Debug("Unhandled type", slog.String("type", typ))

	return "struct{}"
}

// parameterCast adds a type cast (::type) to a parameter if needed.
func (c *converter) parameterCast(sqlName string, sqlType string) string {
	switch sqlType {
	case "documentdb_core.bson", "documentdb_core.bsonsequence":
		return sqlName + "::bytea"
	default:
		return sqlName
	}
}

// toParamComment returns concatenated string of parameter name, data type
// and default value to use for the parameter description of a function.
// If the parameter is not an input, prefix OUT/INOUT is added to the comment.
func (c *converter) toParamComment(paramName string, row map[string]any) string {
	comment := paramName + " " + row["data_type"].(string)
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
func (c *converter) uniqueName(names []string, name string) string {
	i := 1
	for slices.Contains(names, name) {
		name = fmt.Sprintf("%s%d", name, i)
		i++
	}

	return name
}

// generateSQL builds SQL query and arguments for the given function definition.
func generateSQL(f *templateData) string {
	if f.IsProcedure {
		return fmt.Sprintf("CALL %s(%s)", f.SQLName, f.SQLArgs)
	}

	if f.SQLReturns == "" {
		return fmt.Sprintf("SELECT FROM %s(%s)", f.SQLName, f.SQLArgs)
	}

	return fmt.Sprintf(
		"SELECT %s FROM %s(%s)",
		f.SQLReturns,
		f.SQLName,
		f.SQLArgs,
	)
}

// generateGoFunction uses template data to produce go function for querying DocumentDB API.
//
// The function is generated by using template and written to the writer.
func generateGoFunction(writer io.Writer, data *templateData) error {
	q := generateSQL(data)

	queryRowArgs := fmt.Sprintf(`ctx, "%s"`, q)
	if data.QueryRowArgs != "" {
		queryRowArgs = fmt.Sprintf("%s, %s", queryRowArgs, data.QueryRowArgs)
	}
	data.QueryRowArgs = queryRowArgs

	params := "ctx context.Context, conn *pgx.Conn, l *slog.Logger"
	if data.FuncParams != "" {
		params = fmt.Sprintf("%s, %s", params, data.FuncParams)
	}
	data.FuncParams = params

	returns := "err error"
	if data.FuncReturns != "" {
		returns = fmt.Sprintf("%s, %s", data.FuncReturns, returns)
	}
	data.FuncReturns = returns

	return funcTemplate.Execute(writer, &data)
}

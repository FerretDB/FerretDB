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

// Package main contains linter for switches.
package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// orderTypes is the preferred order of types in the switch.
var orderTypes = map[string]int{
	"Document":     1,
	"documentType": 1,

	"Array":     2,
	"arrayType": 2,

	"float64":    3,
	"doubleType": 3,

	"string":     4,
	"stringType": 4,

	"Binary":     5,
	"binaryType": 5,

	"ObjectID":     6,
	"objectIDType": 6,

	"bool":     7,
	"boolType": 7,

	"Time":         8,
	"dateTimeType": 8,

	"NullType": 9,
	"nullType": 9,

	"Regex":     10,
	"regexType": 10,

	"int32":     11,
	"int32Type": 11,

	"Timestamp":     12,
	"timestampType": 12,

	"int64":     13,
	"int64Type": 13,

	"CString": 14,
}

var analyzer = &analysis.Analyzer{
	Name: "checkswitch",
	Doc:  "check the preferred order of types in the switch",
	Run:  run,
}

func main() {
	singlechecker.Main(analyzer)
}

// run is the function to be called by the driver to execute analysis on a single package.
//
// It analyzes the presence of types in 'case' in ascending order of indexes 'orderTypes'.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			var idx int
			var lastName string
			switch n := n.(type) {
			case *ast.TypeSwitchStmt:
				var name string
				for _, el := range n.Body.List {
					if len(el.(*ast.CaseClause).List) < 1 {
						continue
					}

					firstTypeCase := el.(*ast.CaseClause).List[0]
					switch firstTypeCase := firstTypeCase.(type) {
					case *ast.StarExpr:
						if sexp, ok := firstTypeCase.X.(*ast.SelectorExpr); ok {
							name = sexp.Sel.Name
						}
						if sexp, ok := firstTypeCase.X.(*ast.Ident); ok {
							name = sexp.Name
						}

					case *ast.SelectorExpr:
						name = firstTypeCase.Sel.Name

					case *ast.Ident:
						name = firstTypeCase.Name
					}

					idxSl, ok := orderTypes[name]
					if ok && (idxSl < idx) {
						pass.Reportf(n.Pos(), "%s should go before %s in the switch", name, lastName)
					}
					idx, lastName = idxSl, name

					// handling with multiple types,
					// e.g. 'case int32, int64'
					if len(el.(*ast.CaseClause).List) > 1 {
						subidx, sublastName := idx, lastName
						for i := 0; i < len(el.(*ast.CaseClause).List); i++ {
							cs := el.(*ast.CaseClause).List[i]
							switch cs := cs.(type) {
							case *ast.StarExpr:
								if sexp, ok := cs.X.(*ast.SelectorExpr); ok {
									name = sexp.Sel.Name
								}
								if sexp, ok := cs.X.(*ast.Ident); ok {
									name = sexp.Name
								}

							case *ast.SelectorExpr:
								name = fmt.Sprintf("%s.%s", cs.X, cs.Sel.Name)

							case *ast.Ident:
								name = cs.Name
							}

							subidxSl, ok := orderTypes[name]
							if ok && (subidxSl < subidx) {
								pass.Reportf(n.Pos(), "%s should go before %s in the switch", name, sublastName)
							}
							subidx, sublastName = subidxSl, name
						}
					}
				}
			// Handle *ast.SwitchStmt, which represents a switch statement in Go.
			case *ast.SwitchStmt:
				// Declare a variable to hold the name of the case.
				var name string

				// Iterate over each case in the switch statement.
				for _, el := range n.Body.List {
					// If a case has no expressions (i.e., it's a default case), skip to the next case.
					if len(el.(*ast.CaseClause).List) < 1 {
						continue
					}

					// Get the first expression from the case clause.
					firstTypeCase := el.(*ast.CaseClause).List[0]

					// Check if the first expression is a basic literal of type string.
					switch firstTypeCase := firstTypeCase.(type) {
					case *ast.BasicLit:
						if firstTypeCase.Kind == token.STRING {
							// If it is, assign the value of the literal to name.
							name = firstTypeCase.Value
						}
					}

					// Check if the name is in the orderTypes map.
					idxSl, ok := orderTypes[name]
					if ok && (idxSl < idx) {
						// If it is, and the current index is less than the previous index,
						// report that the current case should go before the last case.
						pass.Reportf(n.Pos(), "%s should go before %s in the switch", name, lastName)
					}

					// Update the index and last name to the current index and name.
					idx, lastName = idxSl, name

					// If a case has more than one expression (e.g., 'case "tag1", "tag2":'),
					// check the order of each expression.
					if len(el.(*ast.CaseClause).List) > 1 {
						// Initialize subindex and sublastname with current index and name.
						subidx, sublastName := idx, lastName
						for i := 0; i < len(el.(*ast.CaseClause).List); i++ {
							cs := el.(*ast.CaseClause).List[i]
							switch cs := cs.(type) {
							case *ast.BasicLit:
								if cs.Kind == token.STRING {
									// If it is, assign the value of the literal to name.
									name = cs.Value
								}
							}

							// Check if the name is in the orderTypes map.
							subidxSl, ok := orderTypes[name]
							if ok && (subidxSl < subidx) {
								// If it is, and the current subindex is less than the previous subindex,
								// report that the current case should go before the last case.
								pass.Reportf(n.Pos(), "%s should go before %s in the switch", name, sublastName)
							}

							// Update the subindex and sublastname to the current subindex and name.
							subidx, sublastName = subidxSl, name
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

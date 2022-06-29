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
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// orderTypes is preferred order of the types in the switch.
var orderTypes = map[string]int{
	"Document":       1,
	"documentType":   1,
	"typeCodeObject": 1,

	"Array":         2,
	"arrayType":     2,
	"typeCodeArray": 2,

	"float64":        3,
	"doubleType":     3,
	"typeCodeDouble": 3,

	"string":         4,
	"stringType":     4,
	"typeCodeString": 4,

	"Binary":          5,
	"binaryType":      5,
	"typeCodeBinData": 5,

	"ObjectID":         6,
	"objectIDType":     6,
	"typeCodeObjectID": 6,

	"bool":         7,
	"boolType":     7,
	"typeCodeBool": 7,

	"time.Time":    8,
	"dateTimeType": 8,
	"typeCodeDate": 8,

	"NullType":     9,
	"nullType":     9,
	"typeCodeNull": 9,

	"Regex":         10,
	"regexType":     10,
	"typeCodeRegex": 10,

	"int32":       11,
	"int32Type":   11,
	"typeCodeInt": 11,

	"Timestamp":         12,
	"timestampType":     12,
	"typeCodeTimestamp": 12,

	"int64":        13,
	"int64Type":    13,
	"typeCodeLong": 13,

	"typeCodeNumber": 14,

	"CString": 15,
}

var Analyzer = &analysis.Analyzer{
	Name: "checkswitch",
	Doc:  "checking the preferred order of types in the switch",
	Run:  run,
}

func main() {
	singlechecker.Main(Analyzer)
}

// run is function to be called by driver to execute analysis on a single package.
//
// function analyzes presence of types in 'case' in ascending order of indexes 'orderTypes'.
func run(pass *analysis.Pass) (interface{}, error) {
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
						msg := fmt.Sprintf("non-observance of the preferred order of types: %s <-> %s", lastName, name)
						pass.Reportf(n.Pos(), msg)
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
								msg := fmt.Sprintf("non-observance of the preferred order of types: %s <-> %s", sublastName, name)
								pass.Reportf(n.Pos(), msg)
							}
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

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
	"Document":      1,
	"documentType":  2,
	"Array":         3,
	"arrayType":     4,
	"float64":       5,
	"doubleType":    6,
	"string":        7,
	"stringType":    8,
	"Binary":        9,
	"binaryType":    10,
	"ObjectID":      11,
	"objectIDType":  12,
	"bool":          13,
	"boolType":      14,
	"time.Time":     15,
	"dateTimeType":  16,
	"NullType":      17,
	"nullType":      18,
	"Regex":         19,
	"regexType":     20,
	"int32":         21,
	"int32Type":     22,
	"Timestamp":     23,
	"timestampType": 24,
	"int64":         25,
	"int64Type":     26,
	"CString":       27,
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

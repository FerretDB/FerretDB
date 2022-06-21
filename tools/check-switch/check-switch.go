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
	"Document":    0,
	"primitive.D": 1,
	"Array":       2,
	"primitive.A": 3,
	"float64":     4,
	"string":      5,
	"Binary":      6,
	"ObjectID":    7,
	"bool":        8,
	"time.Time":   9,
	"NullType":    10,
	"Regex":       11,
	"int32":       12,
	"Timestamp":   13,
	"int64":       14,
}

var Analyzer = &analysis.Analyzer{
	Name: "checkerswitch",
	Doc:  "checking the preferred order of types in the switch",
	Run:  run,
}

func main() {
	singlechecker.Main(Analyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			var idx int
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
							// name = fmt.Sprintf("%s.%s", sexp.X.(*ast.Ident).Name, sexp.X.(*ast.Ident).Name)
						}

					case *ast.SelectorExpr:
						name = fmt.Sprintf("%s.%s", firstTypeCase.X, firstTypeCase.Sel.Name)

					case *ast.Ident:
						name = firstTypeCase.Name
					}

					idxSl, ok := orderTypes[name]
					if ok && (idxSl < idx) {
						//						var ls string
						//						for a, b := range orderTypes {
						//							if b == idx {
						//								ls = a
						//								break
						//							}
						//						}
						//						msg := fmt.Sprintf("non-observance of the preferred order of types: %s <-> %s", ls, name)
						//						fmt.Println(msg)
						pass.Reportf(n.Pos(), "T S T")
					}
					idx = idxSl

					if len(el.(*ast.CaseClause).List) > 1 {
						subidx := idx
						for i := 0; i < len(el.(*ast.CaseClause).List); i++ {
							cs := el.(*ast.CaseClause).List[i]
							switch cs := cs.(type) {
							case *ast.StarExpr:
								if sexp, ok := cs.X.(*ast.SelectorExpr); ok {
									name = sexp.Sel.Name
									// name = fmt.Sprintf("%s.%s", sexp.X.(*ast.Ident).Name, sexp.X.(*ast.Ident).Name)
								}

							case *ast.SelectorExpr:
								name = fmt.Sprintf("%s.%s", cs.X, cs.Sel.Name)

							case *ast.Ident:
								name = cs.Name
							}

							subidxSl, ok := orderTypes[name]
							if ok && (subidxSl < subidx) {
								pass.Reportf(n.Pos(), "non-observance of the preferred order of types")
							}
							subidx = subidxSl
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

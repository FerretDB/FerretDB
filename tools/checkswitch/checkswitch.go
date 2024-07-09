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

// orderTags is the preferred order of Tags in the switch.
// //nolint: mnd // the numbers represent the order.
var orderTags = map[string]int{
	"tagFloat64":         1,
	"tagString":          2,
	"tagDocument":        3,
	"tagArray":           4,
	"tagBinary":          5,
	"tagUndefined":       6,
	"tagObjectID":        7,
	"tagBool":            8,
	"tagTime":            9,
	"tagNull":            10,
	"tagRegex":           11,
	"tagDBPointer":       12,
	"tagJavaScript":      13,
	"tagSymbol":          14,
	"tagJavaScriptScope": 15,
	"tagInt32":           16,
	"tagTimestamp":       17,
	"tagInt64":           18,
	"tagDecimal128":      19,
	"tagMinKey":          20,
	"tagMaxKey":          21,
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
// It analyzes the presence of types in 'case' in ascending order of indexes 'orderTypes' and 'orderTags'.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.TypeSwitchStmt:
				checkOrder(orderTypes, n.Body.List, pass, n.Pos())
			case *ast.SwitchStmt:
				checkOrder(orderTags, n.Body.List, pass, n.Pos())
			}

			return true
		})
	}

	return nil, nil
}

// checkOrder checks the order of the case elements in switch statements according to the given `orders`.
func checkOrder(orders map[string]int, list []ast.Stmt, pass *analysis.Pass, pos token.Pos) {
	var order int
	var name string

	// outer loop checks the order of case clauses
	// case "int32":
	// case "int64":
	for _, stmt := range list {
		caseOrder, caseName := order, name

		// inner loop checks the order within a case statement
		// case "int32", "int64":
		for i, caseElem := range stmt.(*ast.CaseClause).List {
			var elemName string

			switch expr := caseElem.(type) {
			case *ast.StarExpr:
				switch starExpr := expr.X.(type) {
				case *ast.SelectorExpr:
					elemName = starExpr.Sel.Name
				case *ast.Ident:
					elemName = starExpr.Name
				default:
					// not `types` or `tags`
				}

			case *ast.SelectorExpr:
				elemName = expr.Sel.Name

			case *ast.Ident:
				elemName = expr.Name

			default:
				// not `types` or `tags`
			}

			elemOrder, ok := orders[elemName]
			if ok && (elemOrder < caseOrder) {
				pass.Reportf(pos, "%s should go before %s in the switch", elemName, caseName)
			}

			caseOrder, caseName = elemOrder, elemName

			if i == 0 {
				// i.e. "int64" is larger than "int32" but below is allowed
				// case "float64", "int64":
				// case "int32":
				order, name = elemOrder, elemName
			}
		}
	}
}

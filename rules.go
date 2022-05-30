package main

import (
	"flag"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var Analyzer = &analysis.Analyzer{
	Name:             "checkswitch",
	Doc:              "reports checkswitch",
	Flags:            flag.FlagSet{},
	Run:              run,
	RunDespiteErrors: false,
	Requires:         []*analysis.Analyzer{},
	ResultType:       nil,
	FactTypes:        []analysis.Fact{},
}

func main() {
	singlechecker.Main(Analyzer)
}

// перевести на мап[тип]индекс.
var correctOrder = []string{
	"Document",
	"Array",
	"float64",
	"string",
	"Binary",
	"ObjectID",
	"bool",
	"time.Time",
	"NullType",
	"Regex",
	"int32",
	"Timestamp",
	"int64",
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			var idx int
			switch n := n.(type) {
			case *ast.CommentGroup:
			//	fmt.Printf("%+v\n", n)
			case *ast.TypeSwitchStmt:
				var name string

				for _, el := range n.Body.List {
					//ast.Print(pass.Fset, n.Body.List)
					for _, cs := range el.(*ast.CaseClause).List {
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
						//	fmt.Printf("%+v\n", cs)

						iSl := indexSlice(name)
						if iSl == -1 {
							continue
						}
						if iSl < idx {
							fmt.Println("неправильный порядок case:", pass.Fset.Position(n.Switch), ":", name)
						}
						idx = iSl
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

func indexSlice(typeCase string) int {
	for idx, tp := range correctOrder {
		if tp == typeCase {
			return idx
		}
	}

	return -1
}

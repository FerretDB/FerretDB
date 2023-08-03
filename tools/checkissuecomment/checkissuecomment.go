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

// Package main contains linter for todo issue comments.
package main

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var todoURLPrefix = "// TODO https://github.com/FerretDB/FerretDB/issues/"

var analyzer = &analysis.Analyzer{
	Name: "checkissuecomment",
	Doc:  "check for TODO comments with issue links",
	Run:  run,
}

func main() {
	singlechecker.Main(analyzer)
}

// It analyses the presence of todo issue in code.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if comment, ok := n.(*ast.CommentGroup); ok {
				for _, c := range comment.List {
					if isMatchingTODOComment(c.Text) {
						fmt.Printf("TODO comment on issue found: %s", c.Text)
						pass.Reportf(c.Pos(), "TODO comment on issue found: %s", c.Text)
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

// Check the comment and todo issue link present.
func isMatchingTODOComment(comment string) bool {
	lines := strings.Split(comment, "\n")
	if len(lines) != 2 {
		return false
	}

	if strings.HasPrefix(lines[1], todoURLPrefix) {
		return true
	}

	return false
}

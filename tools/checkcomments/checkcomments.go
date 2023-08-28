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

// Package main contains linter for comments.
package main

import (
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// todoRE represents correct // TODO comment format.
var todoRE = regexp.MustCompile(`^// TODO \Qhttps://github.com/FerretDB/FerretDB/issues/\E(\d+)$`)

var analyzer = &analysis.Analyzer{
	Name: "checkcomments",
	Doc:  "check TODO comments",
	Run:  run,
}

func main() {
	singlechecker.Main(analyzer)
}

// run analyses TODO comments.
func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				// the space between `//` and `TODO` is always added by `task fmt`
				if strings.HasPrefix(c.Text, "// TODO") {
					// skip comments without URLs for now
					// TODO https://github.com/FerretDB/FerretDB/issues/2733
					if !strings.Contains(c.Text, "https://") {
						continue
					}

					if !todoRE.MatchString(c.Text) {
						pass.Reportf(c.Pos(), "invalid TODO comment")
					}
				}
			}
		}
	}

	return nil, nil
}

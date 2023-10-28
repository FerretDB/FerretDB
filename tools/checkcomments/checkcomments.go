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

	_ "github.com/FerretDB/gh" // TODO https://github.com/FerretDB/FerretDB/issues/2733
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
	// TODO https://github.com/FerretDB/FerretDB/issues/2733
	/*
		token := os.Getenv("GITHUB_TOKEN")

		client, err := gh.NewRESTClient(token, nil)
		if err != nil {
			log.Fatal(err)
		}

		issues := make(map[int]bool)
	*/

	for _, f := range pass.Files {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				line := c.Text

				// the space between `//` and `TODO` is always added by `task fmt`
				if !strings.HasPrefix(line, "// TODO") {
					continue
				}

				if f.Name.Name == "testdata" {
					line, _, _ = strings.Cut(line, ` // want "`)
				}

				match := todoRE.FindStringSubmatch(line)

				if match == nil {
					pass.Reportf(c.Pos(), "invalid TODO: incorrect format")
					continue
				}

				// `go vet -vettool` runs checkcomments once per package.
				// That causes same issues to be checked multiple times,
				// and that easily pushes us over the rate limit.
				// We should cache check results using lockedfile.
				// TODO https://github.com/FerretDB/FerretDB/issues/2733

				/*
					n, err := strconv.Atoi(match[1])
					if err != nil {
						log.Fatal(err)
					}

					open, ok := issues[n]
					if !ok {
						issue, _, err := client.Issues.Get(context.TODO(), "FerretDB", "FerretDB", n)
						if err != nil {
							if errors.As(err, new(*github.RateLimitError)) && token == "" {
								log.Println(
									"Rate limit reached. Please set a GITHUB_TOKEN as described at",
									"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token",
								)

								return nil, nil
							}

							log.Fatalf("%[1]T %[1]s", err)
						}

						open = issue.GetState() == "open"
						issues[n] = open
					}

					if !open {
						pass.Reportf(c.Pos(), "invalid TODO: linked issue is closed")
					}
				*/
			}
		}
	}

	return nil, nil
}

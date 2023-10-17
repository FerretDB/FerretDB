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
	"context"
	"errors"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v56/github"
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
	token := os.Getenv("GITHUB_TOKEN")

	client, err := gh.NewRESTClient(token, nil)
	if err != nil {
		log.Fatal(err)
	}

	issues := make(map[int]bool)

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

				// skip comments without URLs for now
				// TODO https://github.com/FerretDB/FerretDB/issues/2733
				if !strings.Contains(line, "https://") {
					continue
				}

				match := todoRE.FindStringSubmatch(line)

				if match == nil {
					pass.Reportf(c.Pos(), "invalid TODO: incorrect format")
					continue
				}

				n, err := strconv.Atoi(match[1])
				if err != nil {
					log.Fatal(err)
				}

				open, ok := issues[n]
				if !ok {
					issue, _, err := client.Issues.Get(context.TODO(), "FerretDB", "FerretDB", n)
					if err != nil {
						if errors.As(err, new(*github.RateLimitError)) && token == "" {
							log.Printf(
								"%[1]T %[1]s\n%[2]s %[3]s",
								err,
								"Please set a GITHUB_TOKEN as described at",
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
			}
		}
	}

	return nil, nil
}

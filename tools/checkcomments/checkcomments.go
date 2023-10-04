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
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
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

					if !todoRE.MatchString(c.Text) && isIssueOpen(c.Text) {
						pass.Reportf(c.Pos(), "invalid TODO comment and issue is still open")
					}
				}
			}
		}
	}

	return nil, nil
}

// isIssueopen check the issue open or closed.
func isIssueOpen(todoText string) bool {
	issueURL := getURL(todoText)
	if issueURL == "" {
		return false
	}

	token := ""
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	client := github.NewClient(httpClient)

	owner := "owner"
	repo := "https://github.com/FerretDB/FerretDB"
	issueNumber, err := getIssueNumber(issueURL)
	if err != nil {
		log.Fatalf("error in getting issue number: %s", err.Error())
		return false
	}
	issue, _, err := client.Issues.Get(ctx, owner, repo, issueNumber)
	if err != nil {
		log.Fatalf("error in getting status of issue: %s for issue: %s", err.Error(), issueURL)
		return false
	}

	return issue.GetState() == "open"
}

// extracting url from TODO comment if present.
func getURL(todoText string) string {
	arrText := strings.Split(todoText, " ")
	for _, text := range arrText {
		if strings.Contains(text, "https://") {
			parts := strings.Split(text, "/")
			if len(parts) >= 5 {
				issueNumber := parts[6]
				apiURL := fmt.Sprintf("https://api.github.com/repos/FerretDB/FerretDB/issues/%s", issueNumber)

				return apiURL
			}
		}
	}

	return ""
}

// get the issue number from  issue url.
func getIssueNumber(todoText string) (int, error) {
	pattern := `\/issues\/(\d+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(todoText)

	if len(match) >= 2 {
		issueNumberStr := match[1]
		issueNumber, err := strconv.Atoi(issueNumberStr)
		if err != nil {
			return 0, err
		}

		return issueNumber, nil
	}

	return 0, fmt.Errorf("invalid issue url: %s", todoText)
}

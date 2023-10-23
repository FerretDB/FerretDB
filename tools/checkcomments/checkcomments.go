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
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v56/github"
	"github.com/rogpeppe/go-internal/lockedfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

type IssueCache struct {
	Issues map[string]bool `json:"Issues"`
}

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
	current_path, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	cache_path := getCacheFilePath(current_path)

	cf, err := lockedfile.OpenFile(cache_path, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		log.Fatal(err.Error())
	}

	cfs, err := cf.Stat()
	if err != nil {
		log.Fatal(err.Error())
	}

	buffer := make([]byte, cfs.Size())
	_, err = cf.Read(buffer)
	if err != nil {
		log.Fatal(err.Error())
	}

	iCache := &IssueCache{}
	json.Unmarshal(buffer, iCache)

	if len(iCache.Issues) == 0 {
		iCache.Issues = make(map[string]bool)
	}

	token := os.Getenv("GITHUB_TOKEN")

	client, err := gh.NewRESTClient(token, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

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

				iNum := match[1]

				_, ok := iCache.Issues[iNum]

				if !ok {
					n, err := strconv.Atoi(iNum)
					if err != nil {
						log.Fatal(err)
					}

					isOpen, err := isIssueOpen(client, n)
					if err != nil {
						log.Fatal(err)
					}

					if !isOpen {
					}

					iCache.Issues[iNum] = isOpen
				}

				if !iCache.Issues[iNum] {
					pass.Reportf(c.Pos(), "invalid TODO: linked issue is closed")
				}
			}
		}
	}

	jsonb, err := json.Marshal(iCache)
	if err != nil {
		log.Fatal(err.Error())
	}
	cf.WriteAt(jsonb, 0)

	return nil, nil
}

func isIssueOpen(client *github.Client, n int) (bool, error) {
	issue, _, err := client.Issues.Get(context.TODO(), "FerretDB", "FerretDB", n)
	if err != nil {
		if errors.As(err, new(*github.RateLimitError)) && os.Getenv("GITHUB_TOKEN") == "" {
			log.Println(
				"Rate limit reached. Please set a GITHUB_TOKEN as described at",
				"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token",
			)
			return false, err
		}
	}
	return issue.GetState() == "open", nil
}

func getCacheFilePath(p string) string {
	path := filepath.Dir(p)

	readmePath := filepath.Join(path, "README.md")

	_, err := os.Stat(readmePath)

	if os.IsNotExist(err) {
		return getCacheFilePath(path)
	}

	return filepath.Join(path, "tmp", "issue_cache.json")
}

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
	"fmt"
	"go/ast"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-github/v56/github"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// checkIssueComments is used to enable/disable the linter quickly.
const checkIssueComments = true

// struct used to hold open status of issues, true if open otherwise false.
type issueCache struct {
	ReachedRateLimit bool                  `json:"reachedRateLimit"`
	Issues           map[string]issueState `json:"issues"`
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
	if !checkIssueComments {
		return nil, nil
	}

	comments := make(chan ast.Comment, 50)
	go collectTodoComments(pass, comments)

	checkTodoComments(context.Background(), pass, comments, nil, nil)

	//	lockedfile.Transform()

	/*	var iCache issueCache

		currentPath, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		cachePath := getCacheFilePath(currentPath)

		// lockedfile is used to coordinate multiple invocations of the same cache file.
		cf, err := lockedfile.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0o666)
		if err != nil {
			return nil, err
		}

		defer func() {
			if err = cf.Close(); err != nil {
				log.Println(err)
			}
		}()

		stat, err := cf.Stat()
		if err != nil {
			return nil, err
		}

		if stat.Size() > 0 {
			buffer := make([]byte, stat.Size())

			_, err = cf.Read(buffer)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(buffer, &iCache)
			if err != nil {
				return nil, err
			}
		} else {
			iCache.Issues = make(map[string]bool)
		}

		token := os.Getenv("GITHUB_TOKEN")

		client, err := gh.NewRESTClient(token, nil)
		if err != nil {
			return nil, err
		}

		if err := checkTodoComments(pass, &iCache, client); err != nil {
			return nil, err
		}

		if len(iCache.Issues) > 0 {
			jsonb, err := json.Marshal(iCache)
			if err != nil {
				return nil, err
			}

			_, err = cf.WriteAt(jsonb, 0)
			if err != nil {
				return nil, err
			}
		}
	*/

	return nil, nil
}

// collectTodoComments collects comments that contain TODO messages and sends them to a channel.
func collectTodoComments(pass *analysis.Pass, comments chan<- ast.Comment) {
	wgf := sync.WaitGroup{}
	wgc := sync.WaitGroup{}

	wgf.Add(len(pass.Files))

	for _, f := range pass.Files {
		go func(f *ast.File) {
			for _, cg := range f.Comments {
				wgc.Add(len(cg.List))

				go func(cg *ast.CommentGroup) {
					for _, c := range cg.List {
						line := c.Text

						// the space between `//` and `TODO` is always added by `task fmt`
						if !strings.HasPrefix(line, "// TODO") {
							continue
						}

						if f.Name.Name == "testdata" {
							line, _, _ = strings.Cut(line, ` // want "`)
						}

						comments <- *c

						wgc.Done()
					}
				}(cg)
			}

			wgf.Done()
		}(f)
	}

	wgf.Wait()
	wgc.Wait()
	close(comments)
}

// checkTodoComments goes through the given comments and reports if TODOs are valid and linked issues are open.
func checkTodoComments(ctx context.Context, pass *analysis.Pass, comments <-chan ast.Comment, cache *issueCache, client *github.Client) error {
	for c := range comments {
		match := todoRE.FindStringSubmatch(c.Text)

		if match == nil {
			pass.Reportf(c.Pos(), "invalid TODO: incorrect format")
			continue
		}

		issueLink := match[0]

		if state, ok := cache.Issues[issueLink]; ok {
			if state != issueOpen {
				message := fmt.Sprintf("invalid TODO: linked issue %s is %s", issueLink, state)
				pass.Reportf(c.Pos(), message)
			}

			continue
		}

		issueNum, err := strconv.Atoi(match[1])
		if err != nil {
			log.Panicf("invalid issue number: %s", match[1])
		}

		issue, _, err := client.Issues.Get(ctx, "FerretDB", "FerretDB", issueNum)
		var re *github.ErrorResponse

		switch {
		case errors.As(err, new(*github.RateLimitError)):
			cache.ReachedRateLimit = true
			msg := "Rate limit reached."

			if os.Getenv("GITHUB_TOKEN") == "" {
				msg += " Please set a GITHUB_TOKEN as described at " +
					"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token"
			}

			log.Println(msg)

			return nil

		case errors.As(err, &re) && re.Response.StatusCode == 404:
			cache.Issues[issueLink] = issueNotFound
			message := fmt.Sprintf("invalid TODO: linked issue %s does not exist", issueLink)
			pass.Reportf(c.Pos(), message)

		case err != nil:
			return err

		default:
			if issue.GetState() == "closed" {
				cache.Issues[issueLink] = issueClosed
				message := fmt.Sprintf("invalid TODO: linked issue %s is closed", issueLink)
				pass.Reportf(c.Pos(), message)
			} else {
				cache.Issues[issueLink] = issueOpen
			}
		}
	}

	return nil
}

/*

// checkTodoComments goes through the given files' comments and checks if TODOs are valid and linked issues are open.
func checkTodoComments(pass *analysis.Pass, cache *issueCache, client *github.Client) error {
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

				num := match[1]

				if isOpen, ok := cache.Issues[num]; ok {
					if !isOpen {
						message := fmt.Sprintf("invalid TODO: linked issue %s is closed", num)
						pass.Reportf(c.Pos(), message)
					}

					continue
				}

				number, err := strconv.Atoi(num)
				if err != nil {
					return err
				}

				isOpen, err := isIssueOpen(client, number)
				var re *github.ErrorResponse

				switch {
				case errors.As(err, new(*github.RateLimitError)):
					msg := "Rate limit reached."

					if os.Getenv("GITHUB_TOKEN") == "" {
						msg += " Please set a GITHUB_TOKEN as described at " +
							"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token"
					}

					log.Println(msg)

					return nil

				case errors.As(err, &re) && re.Response.StatusCode == 404:
					message := fmt.Sprintf("invalid TODO: linked issue %s does not exist", num)
					pass.Reportf(c.Pos(), message)

				case err != nil:
					return err

				default:
					cache.Issues[num] = isOpen

					if !isOpen {
						message := fmt.Sprintf("invalid TODO: linked issue %s is closed", num)
						pass.Reportf(c.Pos(), message)
					}
				}
			}
		}
	}

	return nil
}

// getCacheFilePath returns the path to the cache file.
//
// Due to analysis tools changing cwd for each package they check,
// we need to find the root of the project in order to get the common cache file.
// This is done by recursively traversing the current path up until README.md is found.
func getCacheFilePath(p string) string {
	path := filepath.Dir(p)

	readmePath := filepath.Join(path, "README.md")

	_, err := os.Stat(readmePath)

	if os.IsNotExist(err) {
		return getCacheFilePath(path)
	}

	return filepath.Join(path, "tmp", "checkcomments", "cache.json")
}

*/

type issueState int

const (
	issueOpen issueState = iota
	issueClosed
	issueNotFound
)

func (i issueState) String() string {
	switch i {
	case issueOpen:
		return "open"
	case issueClosed:
		return "closed"
	case issueNotFound:
		return "not found"
	default:
		panic("invalid issue state")
	}
}

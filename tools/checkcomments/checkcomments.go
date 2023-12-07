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
	"fmt"
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v56/github"
	"github.com/rogpeppe/go-internal/lockedfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// checkIssueComments is used to enable/disable the linter quickly.
const checkIssueComments = true

// issueCache stores info regarding rate limiting and the status of issues.
type issueCache struct {
	ReachedRateLimit bool                  `json:"reachedRateLimit"`
	Issues           map[string]issueState `json:"issues"`
}

// todoRE represents correct // TODO comment format.
var todoRE = regexp.MustCompile(`^// TODO (\Qhttps://github.com/FerretDB/FerretDB/issues/\E(\d+))$`)

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

	comments := collectTodoComments(pass)

	if len(comments) == 0 {
		return nil, nil
	}

	var err error
	var currentPath, cachePath string

	if currentPath, err = os.Getwd(); err != nil {
		log.Panicf("could not get current path: %s", err)
	}

	if cachePath, err = cacheFilePath(currentPath); err != nil {
		log.Panicf("could not get cache file path: %s", err)
	}

	token := os.Getenv("GITHUB_TOKEN")

	client, err := gh.NewRESTClient(token, nil)
	if err != nil {
		log.Panicf("could not create GitHub client: %s", err)
	}

	if err = lockedfile.Transform(cachePath, func(data []byte) ([]byte, error) {
		var cache issueCache

		if len(data) == 0 {
			cache.Issues = make(map[string]issueState)
		} else {
			if err = json.Unmarshal(data, &cache); err != nil {
				log.Panicf("could not unmarshal cache file data: %s", err)
			}
		}

		if cache.ReachedRateLimit {
			return nil, nil
		}

		c := todoChecker{
			client:   client,
			pass:     pass,
			comments: comments,
			cache:    &cache,
			mx:       sync.Mutex{},
		}
		c.run(context.Background())

		data, err = json.Marshal(cache)
		if err != nil {
			log.Panicf("could not marshal cache: %s", err)
		}

		return data, nil
	}); err != nil {
		log.Panicf("could not save cache file: %s", err)
	}

	return nil, nil
}

// collectTodoComments returns comments that contain TODO messages.
func collectTodoComments(pass *analysis.Pass) []*ast.Comment {
	var (
		comments []*ast.Comment
		mx       sync.Mutex
		wgf      sync.WaitGroup
		wgc      sync.WaitGroup
	)
	wgf.Add(len(pass.Files))

	for _, f := range pass.Files {
		go func(f *ast.File) {
			defer wgf.Done()

			for _, cg := range f.Comments {
				wgc.Add(len(cg.List))

				for _, c := range cg.List {
					go func(c *ast.Comment) {
						defer wgc.Done()
						line := c.Text

						// the space between `//` and `TODO` is always added by `task fmt`
						if !strings.HasPrefix(line, "// TODO") {
							return
						}

						if f.Name.Name == "testdata" {
							c.Text, _, _ = strings.Cut(line, ` // want "`)
						}

						mx.Lock()
						comments = append(comments, c)
						mx.Unlock()
					}(c)
				}
			}
		}(f)
	}

	wgf.Wait()
	wgc.Wait()

	return comments
}

// todoChecker is used to go through the given comments and reports if TODOs are valid and linked issues are open.
type todoChecker struct {
	client   *github.Client
	pass     *analysis.Pass
	cache    *issueCache
	comments []*ast.Comment
	mx       sync.Mutex
}

// run executes TODO checker.
func (c *todoChecker) run(ctx context.Context) {
	done := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(len(c.comments))

	for _, comment := range c.comments {
		go func(comment *ast.Comment) {
			defer wg.Done()

			if err := c.processComment(ctx, comment); err != nil {
				done <- struct{}{}
			}
		}(comment)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
}

// processComment checks if the given TODO comment is valid and linked issue is open.
//
// If a critical error occurs, it returns an error. In this case the caller should stop the processing and exit.
func (c *todoChecker) processComment(ctx context.Context, comment *ast.Comment) error {
	match := todoRE.FindStringSubmatch(comment.Text)

	if len(match) != 3 {
		c.mx.Lock()
		c.pass.Reportf(comment.Pos(), "invalid TODO: incorrect format")
		c.mx.Unlock()

		return nil
	}

	issueLink := match[1]

	if state, ok := c.cache.Issues[issueLink]; ok {
		if state != issueOpen {
			message := fmt.Sprintf("invalid TODO: linked issue %s is %s", issueLink, state)

			c.mx.Lock()
			c.pass.Reportf(comment.Pos(), message)
			c.mx.Unlock()
		}

		return nil
	}

	issueNum, err := strconv.Atoi(match[2])
	if err != nil {
		log.Panicf("invalid issue number: %s", match[1])
	}

	issue, _, err := c.client.Issues.Get(ctx, "FerretDB", "FerretDB", issueNum)
	var re *github.ErrorResponse

	switch {
	case errors.As(err, new(*github.RateLimitError)):
		if c.cache.ReachedRateLimit {
			// we already printed the error, so we can just return
			return nil
		}

		c.mx.Lock()
		c.cache.ReachedRateLimit = true
		c.mx.Unlock()

		msg := "Rate limit reached."

		if os.Getenv("GITHUB_TOKEN") == "" {
			msg += " Please set a GITHUB_TOKEN as described at " +
				"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token"
		}

		log.Println(msg)

		return errors.New(msg)

	case errors.As(err, &re) && re.Response.StatusCode == 404:
		c.mx.Lock()
		c.cache.Issues[issueLink] = issueNotFound
		message := fmt.Sprintf("invalid TODO: linked issue %s does not exist", issueLink)
		c.pass.Reportf(comment.Pos(), message)
		c.mx.Unlock()

	case err != nil:
		msg := fmt.Sprintf("could not get issue %d: %s", issueNum, err)
		log.Println(msg)

		return errors.New(msg)

	default:
		c.mx.Lock()
		if issue.GetState() == "closed" {
			c.cache.Issues[issueLink] = issueClosed

			message := fmt.Sprintf("invalid TODO: linked issue %s is closed", issueLink)
			c.pass.Reportf(comment.Pos(), message)
		} else {
			c.cache.Issues[issueLink] = issueOpen
		}
		c.mx.Unlock()
	}

	return nil
}

// cacheFilePath returns the path to the cache file.
//
// It locates the root project directory by moving up from the current location
// until it finds the .git directory. Using .git ensures we find the project's top level,
// unlike other file or directory names that could be found in subdirectories.
//
// If the project root is not found, an error is returned.
func cacheFilePath(p string) (string, error) {
	path := filepath.Dir(p)
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)

	if os.IsNotExist(err) {
		if p == path {
			return "", errors.New("could not find project root")
		}

		return cacheFilePath(path)
	}

	return filepath.Join(path, "tmp", "checkcomments", "cache.json"), nil
}

// issueState represents the state of an issue (open, closed, not found).
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

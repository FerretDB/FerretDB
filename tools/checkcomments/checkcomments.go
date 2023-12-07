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

	"github.com/google/go-github/v56/github"
	"github.com/rogpeppe/go-internal/lockedfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/FerretDB/gh"
)

// checkIssueComments is used to enable/disable the linter quickly.
const checkIssueComments = true

// issueCache stores info regarding rate limiting and the status of issues.
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

	var data []byte
	var cache issueCache

	data, err = lockedfile.Read(cachePath)
	switch {
	case err == nil:
		if err = json.Unmarshal(data, &cache); err != nil {
			log.Panicf("could not unmarshal cache file: %s", err)
		}

	case errors.Is(err, os.ErrNotExist):
		// can't open file, consider the cache empty
		cache.Issues = make(map[string]issueState)

	default:
		log.Panicf("could not read cache file: %s", err)
	}

	if cache.ReachedRateLimit {
		return nil, nil
	}

	token := os.Getenv("GITHUB_TOKEN")

	client, err := gh.NewRESTClient(token, nil)
	if err != nil {
		log.Panicf("could not create GitHub client: %s", err)
	}

	c := todoChecker{
		pass:     pass,
		comments: comments,
		cache:    &cache,
		client:   client,
	}
	c.run(context.Background())

	/*data, err := json.Marshal(cache)
	if err != nil {
		log.Panicf("could not marshal cache file: %s", err)
	}

	lockedfile.Transform()*/

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

// collectTodoComments returns comments that contain TODO messages and sends them to a channel.
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

// todoChecker is used to gothrough the given comments and reports if TODOs are valid and linked issues are open.
type todoChecker struct {
	pass     *analysis.Pass
	comments []*ast.Comment
	cache    *issueCache
	client   *github.Client
}

func (c *todoChecker) processComment(ctx context.Context, comment *ast.Comment) bool {
	match := todoRE.FindStringSubmatch(comment.Text)

	if match == nil {
		c.pass.Reportf(comment.Pos(), "invalid TODO: incorrect format")
		return false
	}

	issueLink := match[0]

	if state, ok := c.cache.Issues[issueLink]; ok {
		if state != issueOpen {
			message := fmt.Sprintf("invalid TODO: linked issue %s is %s", issueLink, state)
			c.pass.Reportf(comment.Pos(), message)
		}

		return false
	}

	issueNum, err := strconv.Atoi(match[1])
	if err != nil {
		log.Panicf("invalid issue number: %s", match[1])
	}

	issue, _, err := c.client.Issues.Get(ctx, "FerretDB", "FerretDB", issueNum)
	var re *github.ErrorResponse

	switch {
	case errors.As(err, new(*github.RateLimitError)):
		c.cache.ReachedRateLimit = true
		msg := "Rate limit reached."

		if os.Getenv("GITHUB_TOKEN") == "" {
			msg += " Please set a GITHUB_TOKEN as described at " +
				"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token"
		}

		log.Println(msg)
		return true

	case errors.As(err, &re) && re.Response.StatusCode == 404:
		c.cache.Issues[issueLink] = issueNotFound
		message := fmt.Sprintf("invalid TODO: linked issue %s does not exist", issueLink)
		c.pass.Reportf(comment.Pos(), message)

	case err != nil:
		log.Printf("could not get issue %d: %s", issueNum, err)
		return true

	default:
		if issue.GetState() == "closed" {
			c.cache.Issues[issueLink] = issueClosed
			message := fmt.Sprintf("invalid TODO: linked issue %s is closed", issueLink)
			c.pass.Reportf(comment.Pos(), message)
		} else {
			c.cache.Issues[issueLink] = issueOpen
		}
	}

	return false
}

// run TODO checker.
func (c *todoChecker) run(ctx context.Context) {
	done := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(len(c.comments))

	for _, comment := range c.comments {
		go func(comment *ast.Comment) {
			defer wg.Done()

			if c.processComment(ctx, comment) {
				close(done)
			}
		}(comment)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
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
*/

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

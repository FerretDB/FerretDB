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
	"flag"
	"log"
	"regexp"
	"strings"

	"github.com/FerretDB/gh"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/FerretDB/FerretDB/tools/github"
)

// todoRE represents correct "// TODO" comment format.
var todoRE = regexp.MustCompile(`^// TODO (\Qhttps://github.com/FerretDB/\E[-\w]+/issues/\d+)$`)

// analyzer represents the checkcomments analyzer.
var analyzer = &analysis.Analyzer{
	Name:  "checkcomments",
	Doc:   "check TODO comments",
	Run:   run,
	Flags: *flag.NewFlagSet("", flag.ExitOnError),
}

// init initializes the analyzer flags.
func init() {
	analyzer.Flags.Bool("offline", false, "do not check issues open/closed status")
	analyzer.Flags.Bool("cache-debug", false, "log cache hits/misses")
	analyzer.Flags.Bool("client-debug", false, "log GitHub API requests/responses")
}

// main runs the analyzer.
func main() {
	singlechecker.Main(analyzer)
}

// run analyses TODO comments.
func run(pass *analysis.Pass) (any, error) {
	var client *github.Client

	if !pass.Analyzer.Flags.Lookup("offline").Value.(flag.Getter).Get().(bool) {
		p, err := github.CacheFilePath()
		if err != nil {
			log.Panic(err)
		}

		cacheDebugF := gh.NoopPrintf
		if pass.Analyzer.Flags.Lookup("cache-debug").Value.(flag.Getter).Get().(bool) {
			cacheDebugF = log.New(log.Writer(), "", log.Flags()).Printf
		}

		clientDebugF := gh.NoopPrintf
		if pass.Analyzer.Flags.Lookup("client-debug").Value.(flag.Getter).Get().(bool) {
			clientDebugF = log.New(log.Writer(), "client-debug: ", log.Flags()).Printf
		}

		if client, err = github.NewClient(p, log.Printf, cacheDebugF, clientDebugF); err != nil {
			log.Panic(err)
		}
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

				if len(match) != 2 {
					pass.Reportf(c.Pos(), "invalid TODO: incorrect format")
					continue
				}

				url := match[1]

				status, err := client.IssueStatus(context.TODO(), url)

				switch {
				case err == nil:
					// nothing
				case errors.Is(err, github.ErrIncorrectURL),
					errors.Is(err, github.ErrIncorrectIssueNumber):
					log.Print(err.Error())
				default:
					log.Panic(err)
				}

				switch status {
				case github.IssueOpen:
					// nothing
				case github.IssueClosed:
					pass.Reportf(c.Pos(), "invalid TODO: linked issue %s is closed", url)
				case github.IssueNotFound:
					pass.Reportf(c.Pos(), "invalid TODO: linked issue %s is not found", url)
				default:
					log.Panicf("unknown issue status: %s", status)
				}
			}
		}
	}

	return nil, nil
}

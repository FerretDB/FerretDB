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

// Package main contains linter for blog posts.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/FerretDB/FerretDB/tools/github"
)

func main() {
	blogFiles, err := filepath.Glob(filepath.Join("website", "blog", "*.md"))
	if err != nil {
		log.Fatal(err)
	}

	err = checkBlogFiles(blogFiles)
	if err != nil {
		log.Fatal(err)
	}
}

// issueRE represents FerretDB issue url with label in Markdown format,
// such as: `[Label](https://github.com/FerretDB/FerretDB/issues/1)`.
// It returns url as a submatch.
var issueRE = regexp.MustCompile(`\[\w+]\((\Qhttps://github.com/FerretDB/\E[-\w]+/issues/\d+)\)`)

// checkBlogFiles verifies that blog posts are correctly formatted.
func checkBlogFiles(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("No blog posts found")
	}

	var failed bool

	for _, file := range files {
		fileInBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("Couldn't read file %s: %s", file, err)
		}

		b, err := extractFrontMatter(fileInBytes)
		if err != nil {
			return fmt.Errorf("Couldn't extract front matter from %s: %s", file, err)
		}

		if err = verifySlug(filepath.Base(file), b); err != nil {
			log.Printf("%q: %s", file, err)
			failed = true
		}

		if err = verifyDateNotPresent(b); err != nil {
			log.Printf("%q: %s", file, err)
			failed = true
		}

		if err = verifyTags(b); err != nil {
			log.Printf("%q: %s", file, err)
			failed = true
		}

		if err = verifyTruncateString(fileInBytes); err != nil {
			log.Printf("%q: %s", file, err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("One or more blog posts are not correctly formatted")
	}

	return nil
}

// verifyTruncateString checks that the truncate string is present.
func verifyTruncateString(b []byte) error {
	if !bytes.Contains(b, []byte("<!--truncate-->")) {
		return fmt.Errorf("<!--truncate--> must be included to have \"Read more\" link on the homepage")
	}

	return nil
}

// extractFrontMatter returns the front matter of a blog post.
func extractFrontMatter(fm []byte) ([]byte, error) {
	var in bool
	var res []byte

	s := bufio.NewScanner(bytes.NewReader(fm))
	for s.Scan() {
		if !in {
			if s.Text() != "---" {
				return nil, fmt.Errorf("expected front matter start on the first line, got %q", s.Text())
			}

			in = true
			continue
		}

		if s.Text() == "---" {
			return res, nil
		}

		res = append(res, s.Bytes()...)
		res = append(res, '\n')
	}

	return nil, fmt.Errorf("front matter end not found")
}

// verifySlug checks that file name and slug match each other.
func verifySlug(fn string, fm []byte) error {
	fnRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-(.*)\.md$`)

	sm := fnRe.FindStringSubmatch(fn)
	if len(sm) != 2 {
		return fmt.Errorf("file name doesn't match the expected pattern")
	}

	fnSlug := sm[1]

	re := regexp.MustCompile("^slug: (.*)")

	var slug string

	s := bufio.NewScanner(bytes.NewReader(fm))
	for s.Scan() {
		if sm = re.FindStringSubmatch(s.Text()); len(sm) > 1 {
			slug = sm[1]
			break
		}
	}

	if err := s.Err(); err != nil {
		return err
	}

	if slug == "" {
		return fmt.Errorf("slug field should be present in the front matter")
	}

	if slug != fnSlug {
		return fmt.Errorf("slug %q doesn't match the file name", slug)
	}

	return nil
}

// verifyDateNotPresent checks date field is not present.
func verifyDateNotPresent(fm []byte) error {
	re := regexp.MustCompile("^date:")

	var found bool

	s := bufio.NewScanner(bytes.NewReader(fm))
	for s.Scan() {
		if re.MatchString(s.Text()) {
			found = true
			break
		}
	}

	if err := s.Err(); err != nil {
		return err
	}

	if found {
		return fmt.Errorf("date field should not be present in the front matter")
	}

	return nil
}

// verifyTags checks that tags are in the allowed list.
func verifyTags(fm []byte) error {
	// tags may span multiple lines after formatting
	re := regexp.MustCompile(`(?ms)^tags:\s+\[(.+)\]`)

	sm := re.FindStringSubmatch(string(fm))
	if len(sm) != 2 {
		return fmt.Errorf("tags field should be present in the front matter")
	}

	// keep in sync with writing-guide.md
	expectedTags := map[string]struct{}{
		"cloud":                   {},
		"community":               {},
		"compatible applications": {},
		"demo":                    {},
		"document databases":      {},
		"events":                  {},
		"hacktoberfest":           {},
		"javascript frameworks":   {},
		"mongodb compatible":      {},
		"mongodb gui":             {},
		"observability":           {},
		"open source":             {},
		"postgresql tools":        {},
		"product":                 {},
		"release":                 {},
		"sspl":                    {},
		"tutorial":                {},
	}

	for _, tag := range strings.Split(sm[1], ",") {
		tag = strings.TrimSpace(tag)

		if _, ok := expectedTags[tag]; !ok {
			return fmt.Errorf("tag %q is not in the allowed list", tag)
		}
	}

	return nil
}

// checkIssueURLs validates FerretDB issues URLs if they occur in the r [io.Reader].
// If URL formatting is invalid, the represented issue is closed or not found - the appropriate
// message is sent to l [*log.Logger].
//
// At the end of scan the true value is returned if any of above was detected.
// An error is returned only if something fatal happened.
func checkIssueURLs(client *github.Client, r io.Reader, l *log.Logger) (bool, error) {
	s := bufio.NewScanner(r)

	var failed bool

	for s.Scan() {
		line := s.Text()

		match := issueRE.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}

		if len(match) != 2 {
			l.Printf("invalid Markdown URL format:\n %s", line)
			failed = true

			continue
		}

		url := match[1]

		var status github.IssueStatus
		status, err := client.IssueStatus(context.TODO(), url)

		switch {
		case err == nil:
			// nothing
		case errors.Is(err, github.ErrIncorrectURL),
			errors.Is(err, github.ErrIncorrectIssueNumber):
			l.Print(err.Error())
		default:
			log.Panic(err)
		}

		switch status {
		case github.IssueOpen:
			// nothing

		case github.IssueClosed:
			failed = true

			l.Printf("linked issue %s is closed", url)

		case github.IssueNotFound:
			failed = true

			l.Printf("linked issue %s is not found", url)

		default:
			return false, fmt.Errorf("unknown issue status: %s", status)
		}
	}

	if err := s.Err(); err != nil {
		return false, fmt.Errorf("error reading input: %s", err)
	}

	return failed, nil
}

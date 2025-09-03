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

// Package main contains linter for documentation and blog posts.
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/FerretDB/gh"

	"github.com/FerretDB/FerretDB/v2/tools/github"
)

func main() {
	blogFiles, err := filepath.Glob(filepath.Join("website", "blog", "*.md"))
	if err != nil {
		log.Fatal(err)
	}

	docsFiles, err := getMarkdownFiles(filepath.Join("website", "docs"))
	if err != nil {
		log.Fatal(err)
	}

	p, err := github.CacheFilePath()
	if err != nil {
		log.Fatal(err)
	}

	client, err := github.NewClient(p, log.Printf, gh.NoopPrintf, gh.NoopPrintf)
	if err != nil {
		log.Fatal(err)
	}

	err = checkBlogFiles(blogFiles)
	if err != nil {
		log.Fatal(err)
	}

	err = checkDocFiles(client, docsFiles)
	if err != nil {
		log.Fatal(err)
	}
}

// checkDocFiles checks files are valid.
func checkDocFiles(client *github.Client, files []string) error {
	var failed bool

	for _, file := range files {
		docFailed, err := checkDocFile(client, file)
		if err != nil {
			return err
		}

		if docFailed {
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("one or more docs contain invalid issue URLs")
	}

	return nil
}

// checkDocFile verifies the file contain valid issue URLs.
func checkDocFile(client *github.Client, file string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return true, fmt.Errorf("could not read file %s: %s", file, err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	issueURLFailed, err := checkIssueURLs(client, f, file, log.Default())
	if err != nil {
		log.Printf("%q: %s", file, err)
	}

	return issueURLFailed, nil
}

// getMarkdownFiles returns markdown files in the given directory.
func getMarkdownFiles(path string) ([]string, error) {
	var markdownFiles []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".md" || filepath.Ext(path) == ".mdx" {
			markdownFiles = append(markdownFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return markdownFiles, nil
}

// issueRE represents FerretDB and documentdb issue url.
// It returns owner, repo and issue number as submatches.
var issueRE = regexp.MustCompile(`\Qhttps://github.com/\E(FerretDB|documentdb)/([-\w]+)/issues/(\d+)`)

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

	// keep in sync with content-process.md
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

// checkIssueURLs validates FerretDB and documentdb issues URLs if they occur in the r [io.Reader].
// If URL formatting is invalid, the represented issue is closed or not found - the appropriate
// message is sent to l [*log.Logger].
//
// At the end of scan the true value is returned if any of above was detected.
// An error is returned only if something fatal happened.
func checkIssueURLs(client *github.Client, r io.Reader, filename string, l *log.Logger) (bool, error) {
	s := bufio.NewScanner(r)

	var failed bool

	for s.Scan() {
		line := s.Text()

		match := issueRE.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}

		if len(match) != 4 {
			l.Printf("invalid URL format:\n %s", line)
			failed = true

			continue
		}

		url, owner, repo := match[0], match[1], match[2]

		num, err := strconv.Atoi(match[3])
		if err != nil {
			panic(err)
		}

		if num <= 0 {
			l.Printf("incorrect issue number: %s in %s\n", line, filename)
			failed = true

			continue
		}

		status, err := client.IssueStatus(context.TODO(), owner, repo, num)
		if err != nil {
			log.Panic(err)
		}

		switch status {
		case github.IssueOpen:
			// nothing

		case github.IssueClosed:
			failed = true

			l.Printf("linked issue %s is closed in %s", url, filename)

		case github.IssueNotFound:
			failed = true

			l.Printf("linked issue %s is not found in %s", url, filename)

		default:
			return false, fmt.Errorf("unknown issue status: %s in %s", status, filename)
		}
	}

	if err := s.Err(); err != nil {
		return false, fmt.Errorf("error reading input: %s", err)
	}

	return failed, nil
}

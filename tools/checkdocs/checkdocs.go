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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	files, err := filepath.Glob(filepath.Join("website", "blog", "*.md"))
	if err != nil {
		log.Fatal(err)
	}

	checkFiles(files, log.Printf, log.Fatalf)
}

// checkFiles verifies that blog posts are correctly formatted,
// using logf for progress reporting and fatalf for errors.
func checkFiles(files []string, logf, fatalf func(string, ...any)) {
	if len(files) == 0 {
		fatalf("No blog posts found")
	}

	var failed bool

	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			fatalf("Couldn't read file %s: %s", file, err)
		}

		if b, err = extractFrontMatter(b); err != nil {
			fatalf("Couldn't extract front matter from %s: %s", file, err)
		}

		if err = verifySlug(filepath.Base(file), b); err != nil {
			logf("%q: %s", file, err)
			failed = true
		}

		if err = verifyDateNotPresent(b); err != nil {
			logf("%q: %s", file, err)
			failed = true
		}

		if err = verifyTags(b); err != nil {
			logf("%q: %s", file, err)
			failed = true
		}
	}

	if failed {
		fatalf("One or more blog posts are not correctly formatted")
	}
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

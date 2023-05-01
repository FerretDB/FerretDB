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

// Package contains linter for blog posts.
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// fileSlug describes the file name and expected slug of blog posts.
type fileSlug struct {
	slug     string
	fileName string
}

func main() {
	dir := filepath.Join("website", "blog")

	fs, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	slugs := getBlogSlugs(fs)

	pass := true

	for _, slug := range slugs {
		// wrap in func to avoid possible resource leak of calling defer in the loop.
		func() {
			fo, err := os.Open(filepath.Join(dir, slug.fileName))
			if err != nil {
				log.Fatalf("Couldn't open file: %s", slug.fileName)
			}

			defer fo.Close()

			if err = verifySlug(slug, fo); err != nil {
				log.Print(err)
				pass = false
			}
		}()
	}

	if !pass {
		log.Fatal("One or more blog posts are not correctly formatted")
	}
}

// getBlogSlugs returns slice containing fileSlug for each DirEntry.
func getBlogSlugs(fs []fs.DirEntry) []fileSlug {
	// start with 4 digits then a 0[1-9] or 1 [0 1 or 2]
	// then - and either 0 [1-9] or [1 or 2][0-9] or 3[0 or 1] - slug(any) and end with .md.
	fnRegex := regexp.MustCompile(`^\d{4}\-(?:0[1-9]|1[012])\-(?:0[1-9]|[12][0-9]|3[01])-(.*).md$`)
	mdRegex := regexp.MustCompile(`.md$`)

	var fileSlugs []fileSlug

	for _, f := range fs {
		fn := f.Name()

		if !mdRegex.MatchString(fn) {
			continue
		}

		sm := fnRegex.FindStringSubmatch(fn)

		if len(sm) > 2 {
			log.Fatalf("File %s is not correctly formated (yyyy-mm-dd-'slug'.md)", fn)
			continue
		}

		fileSlugs = append(fileSlugs, fileSlug{sm[len(sm)-1], fn})
	}

	return fileSlugs
}

// verifySlug returns error if file doesn't contain expected slug.
func verifySlug(fS fileSlug, f io.Reader) error {
	r := regexp.MustCompile("^slug: (.*)")

	pass := false

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		sm := r.FindStringSubmatch(scanner.Text())
		if len(sm) > 1 && sm[len(sm)-1] == fS.slug {
			pass = true
		}
	}

	if !pass {
		return fmt.Errorf("slug is not correctly formated in file %s", fS.fileName)
	}

	return nil
}

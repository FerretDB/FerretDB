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

package main

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//func TestReal(t *testing.T) {
//	files, err := filepath.Glob(filepath.Join("..", "..", "website", "blog", "*.md"))
//	require.NoError(t, err)
//
//	checkBlogFiles(files)
//
//	tableFile, err := filepath.Abs(filepath.Join("website", "docs", "reference", "supported-commands.md"))
//	require.NoError(t, err)
//
//	f, err := os.OpenFile(tableFile, os.O_RDONLY, 0o666)
//	if err != nil {
//		log.Fatalf("couldn't open the file %s: %s", f, err)
//	}
//
//	defer f.Close()
//
//	checkSupportedCommands(f)
//}

var fm = bytes.TrimSpace([]byte(`
slug: using-ferretdb-with-studio-3t
date: 2023-04-18
title: Using FerretDB 1.0 with Studio 3T
authors: [alex]
description: >
	Discover how to use FerretDB 1.0 with Studio 3T, and explore ways to leverage FerretDB for MongoDB GUI applications.
image: /img/blog/ferretdb-studio3t.png
tags:
	[
		tutorial,
		mongodb compatible,
		mongodb gui,
		compatible applications,
		documents databases
	]
	`))

func TestVerifySlug(t *testing.T) {
	err := verifySlug("2023-04-18-using-ferretdb-with-studio.md", fm)
	assert.EqualError(t, err, `slug "using-ferretdb-with-studio-3t" doesn't match the file name`)
}

func TestVerifyDateNotPresent(t *testing.T) {
	err := verifyDateNotPresent(fm)
	assert.EqualError(t, err, `date field should not be present in the front matter`)
}

func TestVerifyTags(t *testing.T) {
	err := verifyTags(fm)
	assert.EqualError(t, err, `tag "documents databases" is not in the allowed list`)
}

func TestVerifyTruncateString(t *testing.T) {
	err := verifyTruncateString(fm)
	assert.EqualError(t, err, "<!--truncate--> must be included to have \"Read more\" link on the homepage")
}

func TestCheckSupportedCommands(t *testing.T) {
	//pr, pw := io.Pipe()
	//t.Cleanup(func() {
	//	require.NoError(t, pw.Close())
	//})

	//lr := bufio.NewReader(pr)

	var buf bytes.Buffer

	l := log.New(&buf, "", 0)

	a, err := NewSupportedCommandsAnalyzer(l)
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		Payload        string
		ExpectedOutput string
	}{
		"OpenIssueLink": {
			Payload:        "|                 | `openIssueLink`          | ❌     | [Issue](https://github.com/FerretDB/FerretDB/issues/3413) |",
			ExpectedOutput: "",
		},

		"ClosedIssueLink": {
			Payload:        "|                 | `closedIssueLink`          | ❌     | [Issue](https://github.com/FerretDB/FerretDB/issues/1) |",
			ExpectedOutput: "linked issue https://github.com/FerretDB/FerretDB/issues/1 is closed\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := strings.NewReader(tc.Payload)
			require.NoError(t, a.Scan(r))

			actualOutput := buf.String()
			assert.Equal(t, tc.ExpectedOutput, actualOutput)
		})
	}
}

//func TestVerifyIssues(t *testing.T) {
//	t.Parallel()
//
//	path, err := github.CacheFilePath()
//	require.NoError(t, err)
//
//	err = os.MkdirAll(filepath.Dir(path), 0o777)
//	require.NoError(t, err)
//
//	t.Run("Empty fm", func(t *testing.T) {
//		verifyIssues(nil, gh.NoopPrintf, gh.NoopPrintf)
//	})
//
//	t.Run("Sample row", func(t *testing.T) {
//		fm = []byte("| ❌     | [Issue](https://github.com/FerretDB/FerretDB/issues/4035) |")
//
//		verifyIssues(fm, gh.NoopPrintf, gh.NoopPrintf)
//	})
//}

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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReal(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "website", "blog", "*.md"))
	require.NoError(t, err)

	checkFiles(files, t.Logf, t.Fatalf)
}

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

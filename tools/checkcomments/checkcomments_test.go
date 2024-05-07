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
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go-simpler.org/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/FerretDB/FerretDB/tools/github"
)

func TestCheckCommentIssue(t *testing.T) {
	t.Parallel()

	path, err := github.CacheFilePath("checkcomments")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Dir(path), 0o777)
	require.NoError(t, err)

	analysistest.Run(t, analysistest.TestData(), analyzer)
}

func TestCacheFilePath(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	expected := filepath.Join(wd, "..", "..", "tmp", "githubcache", "cache.json")

	actual, err := cacheFilePath()
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestClient(t *testing.T) {
	t.Parallel()

	cacheFilePath := filepath.Join(t.TempDir(), "cache.json")
	ctx := context.Background()

	t.Run("CheckIssueStatus", func(t *testing.T) {
		t.Parallel()

		c, err := newClient(cacheFilePath, t.Logf, t.Logf, t.Logf)
		require.NoError(t, err)

		actual, err := c.checkIssueStatus(ctx, "FerretDB", 10)
		require.NoError(t, err)
		assert.Equal(t, issueOpen, actual)

		actual, err = c.checkIssueStatus(ctx, "FerretDB", 1)
		require.NoError(t, err)
		assert.Equal(t, issueClosed, actual)

		actual, err = c.checkIssueStatus(ctx, "FerretDB", 999999)
		require.NoError(t, err)
		assert.Equal(t, issueNotFound, actual)
	})

	t.Run("IssueStatus", func(t *testing.T) {
		t.Parallel()

		c, err := newClient(cacheFilePath, t.Logf, t.Logf, t.Logf)
		require.NoError(t, err)

		actual, err := c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/10", "FerretDB", 10)
		require.NoError(t, err)
		assert.Equal(t, issueOpen, actual)

		actual, err = c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/1", "FerretDB", 1)
		require.NoError(t, err)
		assert.Equal(t, issueClosed, actual)

		actual, err = c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/999999", "FerretDB", 999999)
		require.NoError(t, err)
		assert.Equal(t, issueNotFound, actual)

		// The following tests should use cache and not the client,
		// but it may be empty if tests above failed for some reason.

		if t.Failed() {
			return
		}

		c.c = nil

		actual, err = c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/10", "FerretDB", 10)
		require.NoError(t, err)
		assert.Equal(t, issueOpen, actual)

		actual, err = c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/1", "FerretDB", 1)
		require.NoError(t, err)
		assert.Equal(t, issueClosed, actual)

		actual, err = c.IssueStatus(ctx, "https://github.com/FerretDB/FerretDB/issues/999999", "FerretDB", 999999)
		require.NoError(t, err)
		assert.Equal(t, issueNotFound, actual)
	})
}

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

// To prevent fork bomb.
//go:build !ferretdb_testcover

package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx, cancel := context.WithTimeout(testutil.Ctx(t), 5*time.Second)
	t.Cleanup(cancel)

	bin := filepath.Join("..", "..", "bin", "ferretdb")

	cmd := exec.CommandContext(ctx, bin, "--version")
	b, err := cmd.Output()
	require.NoError(t, err)
	assert.Regexp(t, `version: v([0-9]+)\.([0-9]+)\.([0-9]+)`, string(b))
	assert.Regexp(t, `branch: \w+`, string(b))
	commit := regexp.MustCompile(`commit: ([0-9a-f]{40})`).FindStringSubmatch(string(b))
	require.Len(t, commit, 2)

	t.Skip("https://github.com/FerretDB/FerretDB/issues/2102")

	cmd = exec.CommandContext(ctx, "go", "version", "-m", bin)
	b, err = cmd.Output()
	require.NoError(t, err)
	revision := regexp.MustCompile(`vcs.revision=([0-9a-f]{40})`).FindStringSubmatch(string(b))
	require.NotEmpty(t, revision)
	require.Len(t, revision, 2)

	assert.Equal(t, commit[1], revision[1])
}

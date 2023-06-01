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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShardIntegrationTests(t *testing.T) {
	t.Parallel()

	var w *bytes.Buffer
	output, err := shardIntegrationTests(w, 0, 2)
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Regexp(t, "^\\^.*\\$$", output)
}

func TestShardTests(t *testing.T) {
	t.Parallel()

	testNames, err := shardTests(0, 2, "integration")
	assert.NoError(t, err)
	assert.Len(t, testNames, 2)
}

func TestGetAllTestNames(t *testing.T) {
	t.Parallel()

	testNames, err := getAllTestNames("integration")
	assert.NoError(t, err)
	assert.NotEmpty(t, testNames)
}

func TestChdirToRoot(t *testing.T) {
	t.Parallel()

	// while running the tests the current working location is where the test is
	oldWorkingDir, err := os.Getwd()
	assert.NoError(t, err)
	err = chdirToRoot()
	assert.NoError(t, err)
	NewWorkingDir, err := os.Getwd()
	assert.NoError(t, err)
	assert.NotEqual(t, oldWorkingDir, NewWorkingDir)
}

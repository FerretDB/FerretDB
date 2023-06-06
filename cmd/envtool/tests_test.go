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
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShardIntegrationTests(t *testing.T) {
	t.Parallel()

	var w bytes.Buffer
	err := shardIntegrationTests(&w, 99, 150)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	s := w.String()
	assert.Regexp(t, "^\\^.*\\$$", s)
}

func TestShardTests(t *testing.T) {
	t.Parallel()

	t.Run("Shard odd amount of tests into 2 parts", func(t *testing.T) {
		tests := []string{}
		for i := 0; i < 10; i++ {
			tests = append(tests, strconv.Itoa(i))
		}

		testNames, err := shardTests(0, 2, tests...)
		assert.NoError(t, err)
		assert.Len(t, testNames, 5)
		assert.Equal(t, tests[0:5], testNames)

		testNames, err = shardTests(1, 2, tests...)
		assert.NoError(t, err)
		assert.Len(t, testNames, 5)
		assert.Equal(t, tests[5:], testNames)
	})

	t.Run("Shard even amount of tests into 2 parts", func(t *testing.T) {
		tests := []string{}
		for i := 0; i < 11; i++ {
			tests = append(tests, strconv.Itoa(i))
		}

		testNames, err := shardTests(0, 2, tests...)
		assert.NoError(t, err)
		assert.Len(t, testNames, 5)
		assert.Equal(t, tests[0:5], testNames)

		testNames, err = shardTests(1, 2, tests...)
		assert.NoError(t, err)
		assert.Len(t, testNames, 6)
		assert.Equal(t, tests[5:], testNames)
	})

	t.Run("Shard one test when total is 2", func(t *testing.T) {
		tests := []string{"1"}

		testNames, err := shardTests(0, 2, tests...)
		assert.EqualError(t, err, "Cannot shard when Total is greater than amount of tests (2 > 1)")
		assert.Nil(t, testNames)
	})

	t.Run("Shard empty tests", func(t *testing.T) {
		tests := []string{}

		testNames, err := shardTests(0, 2, tests...)
		assert.EqualError(t, err, "Cannot shard when Total is greater than amount of tests (2 > 0)")
		assert.Nil(t, testNames)
	})

	t.Run("Index is greater than Total", func(t *testing.T) {
		tests := []string{}

		testNames, err := shardTests(2, 0, tests...)
		assert.EqualError(t, err, "Cannot shard when Index is greater or equal to Total (2 >= 0)")
		assert.Nil(t, testNames)
	})

	t.Run("Index is equal to Total", func(t *testing.T) {
		tests := []string{}

		testNames, err := shardTests(0, 0, tests...)
		assert.EqualError(t, err, "Cannot shard when Index is greater or equal to Total (0 >= 0)")
		assert.Nil(t, testNames)
	})
}

func TestGetAllTestNames(t *testing.T) {
	t.Parallel()

	testNames, err := getAllTestNames("integration")
	assert.NoError(t, err)
	assert.NotEmpty(t, testNames)
}

func TestGetNewWorkingDir(t *testing.T) {
	t.Parallel()

	// while running the tests the current working location is where the test is
	oldWorkingDir, err := os.Getwd()
	assert.NoError(t, err)
	dir, err := getNewWorkingDir("integration")
	assert.NoError(t, err)
	assert.NotEqual(t, oldWorkingDir, dir)
	assert.Contains(t, dir, "integration")
	assert.NotContains(t, dir, "envtool")
	assert.True(t, strings.HasSuffix(dir, "/FerretDB/integration"))
}

func TestGetRootDir(t *testing.T) {
	t.Parallel()

	oldWorkingDir, err := os.Getwd()
	assert.NoError(t, err)
	rootDir, err := getRootDir()
	assert.NoError(t, err)
	assert.NotEqual(t, oldWorkingDir, rootDir)
	assert.NotContains(t, rootDir, "integration")
	assert.NotContains(t, rootDir, "envtool")
	assert.True(t, strings.HasSuffix(rootDir, "/FerretDB"))
}

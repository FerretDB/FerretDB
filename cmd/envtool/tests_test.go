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

	t.Run("Successfully Shard Integration tests", func(t *testing.T) {
		var w bytes.Buffer
		err := shardIntegrationTests(&w, 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, w.Bytes())
		s := w.String()
		assert.Regexp(t, "^\\^.*\\$$", s)
	})

	t.Run("Fail Shard Integration tests - too big index", func(t *testing.T) {
		var w bytes.Buffer
		err := shardIntegrationTests(&w, 12, 10)
		assert.EqualError(t, err, "Cannot shard when Index is greater or equal to Total (12 >= 10)")
		assert.Nil(t, w.Bytes())
	})

	t.Run("Failp Shard Integration tests - too big index", func(t *testing.T) {
		var w bytes.Buffer
		err := shardIntegrationTests(&w, 12, 100000000000000)
		assert.ErrorContains(t, err, "Cannot shard when Total is greater than amount of tests")
		assert.Nil(t, w.Bytes())
	})
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

func TestPrepareTestNamesOutput(t *testing.T) {
	t.Parallel()

	t.Run("Fails when the status is not OK", func(t *testing.T) {
		output := `TestAggregateCommandCompat
fail      github.com/FerretDB/FerretDB/integration        0.023s`

		tests, err := prepareTestNamesOutput(output)
		assert.ErrorContains(t, err, "Could not read test names:")
		assert.Nil(t, tests)
	})

	t.Run("Passes when the status is OK", func(t *testing.T) {
		output := `TestAggregateCommandCompat
TestAggregateCompatStages
BenchmarkInsertMany
ok      github.com/FerretDB/FerretDB/integration        0.023s`
		expectedOutput := []string{"BenchmarkInsertMany", "TestAggregateCommandCompat", "TestAggregateCompatStages"}

		tests, err := prepareTestNamesOutput(output)
		assert.NoError(t, err)
		assert.Equal(t, tests, expectedOutput)
	})
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

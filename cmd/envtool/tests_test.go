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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// makeTestLogger returns a logger that adds all messages to the given slice.
func makeTestLogger(messages *[]string) (*zap.Logger, error) {
	logger, err := makeLogger(zap.InfoLevel, nil)
	if err != nil {
		return nil, err
	}

	logger = logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		*messages = append(*messages, entry.Message)
		return nil
	}))

	return logger, nil
}

func TestRunGoTest(t *testing.T) {
	t.Parallel()

	// sometimes it is "(0.01s)" on CI
	timingRe := regexp.MustCompile(` \(\d+\.\d+s\)$`)

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(context.TODO(), []string{"./testdata", "-run=TestNormal"}, 2, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"PASS TestNormal1 1/2",
			"PASS TestNormal2 2/2",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}
		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(context.TODO(), []string{"./testdata", "-run=TestError"}, 2, false, logger.Sugar())

		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		assert.Equal(t, 1, exitErr.ExitCode())

		expected := []string{
			"FAIL TestError1 1/2:",
			"  === RUN   TestError1",
			"  error_test.go:20: not hidden 1",
			"  error_test.go:22: Error 1",
			"  error_test.go:24: not hidden 2",
			"  --- FAIL: TestError1",
			"",
			"FAIL TestError2/Parallel:",
			"    === RUN   TestError2/Parallel",
			"    error_test.go:35: not hidden 5",
			"    === PAUSE TestError2/Parallel",
			"    === CONT  TestError2/Parallel",
			"    error_test.go:39: not hidden 6",
			"    error_test.go:41: Error 2",
			"    error_test.go:43: not hidden 7",
			"    --- FAIL: TestError2/Parallel",
			"",
			"FAIL TestError2 2/2:",
			"  === RUN   TestError2",
			"  error_test.go:28: not hidden 3",
			"  === PAUSE TestError2",
			"  === CONT  TestError2",
			"  error_test.go:32: not hidden 4",
			"  === RUN   TestError2/Parallel",
			"  error_test.go:35: not hidden 5",
			"  === PAUSE TestError2/Parallel",
			"  === RUN   TestError2/NotParallel",
			"  error_test.go:47: not hidden for parent",
			"  --- PASS: TestError2/NotParallel",
			"  === CONT  TestError2/Parallel",
			"  error_test.go:39: not hidden 6",
			"  error_test.go:41: Error 2",
			"  error_test.go:43: not hidden 7",
			"  --- FAIL: TestError2/Parallel",
			"  --- FAIL: TestError2",
			"",
			"FAIL github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		for i, line := range expected {
			actual[i] = timingRe.ReplaceAllString(line, "")
		}

		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("Skip", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(context.TODO(), []string{"./testdata", "-run=TestSkip"}, 1, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"SKIP TestSkip1 1/1:",
			"  === RUN   TestSkip1",
			"  skip_test.go:20: not hidden 1",
			"  skip_test.go:22: Skip 1",
			"  --- SKIP: TestSkip1",
			"",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		for i, line := range expected {
			actual[i] = timingRe.ReplaceAllString(line, "")
		}

		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})
}

func TestListTestFuncs(t *testing.T) {
	t.Parallel()

	actual, err := listTestFuncs("./testdata")
	require.NoError(t, err)
	expected := []string{
		"TestError1",
		"TestError2",
		"TestNormal1",
		"TestNormal2",
		"TestSkip1",
	}
	assert.Equal(t, expected, actual)
}

func TestShardTestFuncs(t *testing.T) {
	t.Parallel()

	testFuncs, err := listTestFuncs(filepath.Join("..", "..", "integration"))
	require.NoError(t, err)
	assert.Contains(t, testFuncs, "TestQueryCompatLimit")

	t.Run("InvalidIndex", func(t *testing.T) {
		t.Parallel()

		_, err := shardTestFuncs(0, 3, testFuncs)
		assert.EqualError(t, err, "index must be greater than 0")

		_, err = shardTestFuncs(3, 3, testFuncs)
		assert.NoError(t, err)

		_, err = shardTestFuncs(4, 3, testFuncs)
		assert.EqualError(t, err, "cannot shard when index is greater than total (4 > 3)")
	})

	t.Run("InvalidTotal", func(t *testing.T) {
		t.Parallel()

		_, err := shardTestFuncs(3, 1000, testFuncs[:42])
		assert.EqualError(t, err, "cannot shard when total is greater than a number of test functions (1000 > 42)")
	})

	t.Run("Valid", func(t *testing.T) {
		t.Parallel()

		res, err := shardTestFuncs(1, 3, testFuncs)
		require.NoError(t, err)
		assert.Equal(t, testFuncs[0], res[0])
		assert.NotEqual(t, testFuncs[1], res[1])
		assert.NotEqual(t, testFuncs[2], res[1])
		assert.Equal(t, testFuncs[3], res[1])

		res, err = shardTestFuncs(3, 3, testFuncs)
		require.NoError(t, err)
		assert.NotEqual(t, testFuncs[0], res[0])
		assert.NotEqual(t, testFuncs[1], res[0])
		assert.Equal(t, testFuncs[2], res[0])
	})
}

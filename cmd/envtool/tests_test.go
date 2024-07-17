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
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
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

var (
	cleanupTimingRe    = regexp.MustCompile(`(\d+\.\d+)s`)                   // duration are different
	cleanupGoroutineRe = regexp.MustCompile(`goroutine (\d+)`)               // goroutine IDs are different
	cleanupPathRe      = regexp.MustCompile(`^\t(.+)/cmd/envtool/testdata/`) // absolute file paths are different
	cleanupAddrRe      = regexp.MustCompile(` (\+0x[0-9a-f]+)$`)             // addresses are different
)

// cleanup removes variable parts of the output.
func cleanup(lines []string) {
	for i, line := range lines {
		if loc := cleanupTimingRe.FindStringSubmatchIndex(line); loc != nil {
			line = line[:loc[2]] + "<SEC>" + line[loc[3]:]
		}

		if loc := cleanupGoroutineRe.FindStringSubmatchIndex(line); loc != nil {
			line = line[:loc[2]] + "<ID>" + line[loc[3]:]
		}

		if loc := cleanupPathRe.FindStringSubmatchIndex(line); loc != nil {
			line = line[:loc[2]] + "<DIR>" + line[loc[3]:]
		}

		if loc := cleanupAddrRe.FindStringSubmatchIndex(line); loc != nil {
			line = line[:loc[2]] + "<ADDR>" + line[loc[3]:]
		}

		lines[i] = line
	}
}

func TestRunGoTest(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestNormal"}, 2, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"PASS TestNormal1 1/2",
			"PASS TestNormal2 2/2",
			"PASS",
			"ok  	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		cleanup(actual)
		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("SubtestsPartial", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestWithSubtest/Third"}, 1, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"PASS TestWithSubtest 1/1",
			"PASS",
			"ok  	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		cleanup(actual)

		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("SubtestsNotFound", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestWithSubtest/None"}, 1, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"PASS TestWithSubtest 1/1",
			"testing: warning: no tests to run",
			"PASS",
			"ok  	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s [no tests to run]",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		cleanup(actual)

		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestError"}, 2, false, logger.Sugar())

		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		assert.Equal(t, 1, exitErr.ExitCode())

		expected := []string{
			"FAIL TestError1 1/2:",
			"=== RUN   TestError1",
			"    error_test.go:20: not hidden 1",
			"    error_test.go:22: Error 1",
			"    error_test.go:24: not hidden 2",
			"--- FAIL: TestError1 (<SEC>s)",
			"",
			"FAIL TestError2/Parallel:",
			"=== RUN   TestError2/Parallel",
			"    error_test.go:35: not hidden 5",
			"=== PAUSE TestError2/Parallel",
			"=== CONT  TestError2/Parallel",
			"    error_test.go:39: not hidden 6",
			"    error_test.go:41: Error 2",
			"    error_test.go:43: not hidden 7",
			"--- FAIL: TestError2/Parallel (<SEC>s)",
			"",
			"FAIL TestError2 2/2:",
			"=== RUN   TestError2",
			"    error_test.go:28: not hidden 3",
			"=== PAUSE TestError2",
			"=== CONT  TestError2",
			"    error_test.go:32: not hidden 4",
			"--- FAIL: TestError2 (<SEC>s)",
			"",
			"FAIL",
			"FAIL	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s",
			"FAIL github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		cleanup(actual)
		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("Skip", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestSkip"}, 1, false, logger.Sugar())
		require.NoError(t, err)

		expected := []string{
			"SKIP TestSkip1 1/1:",
			"=== RUN   TestSkip1",
			"    skip_test.go:20: not hidden 1",
			"    skip_test.go:22: Skip 1",
			"--- SKIP: TestSkip1 (<SEC>s)",
			"",
			"PASS",
			"ok  	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s",
			"PASS github.com/FerretDB/FerretDB/cmd/envtool/testdata",
		}

		cleanup(actual)
		assert.Equal(t, expected, actual, "actual:\n%s", strings.Join(actual, "\n"))
	})

	t.Run("Panic", func(t *testing.T) {
		t.Parallel()

		var actual []string
		logger, err := makeTestLogger(&actual)
		require.NoError(t, err)

		err = runGoTest(ctx, []string{"./testdata", "-count=1", "-run=TestPanic"}, 1, false, logger.Sugar())
		require.Error(t, err)

		expected := []string{
			"FAIL	github.com/FerretDB/FerretDB/cmd/envtool/testdata	<SEC>s",
			"FAIL github.com/FerretDB/FerretDB/cmd/envtool/testdata",
			"",
			"Some tests did not finish:",
			"  github.com/FerretDB/FerretDB/cmd/envtool/testdata.TestPanic1",
			"",
			"github.com/FerretDB/FerretDB/cmd/envtool/testdata.TestPanic1:",
			"=== RUN   TestPanic1",
			"panic: Panic 1",
			"",
			"goroutine <ID> [running]:",
			"github.com/FerretDB/FerretDB/cmd/envtool/testdata.TestPanic1.func1()",
			"	<DIR>/cmd/envtool/testdata/panic_test.go:25 <ADDR>",
			"created by github.com/FerretDB/FerretDB/cmd/envtool/testdata.TestPanic1 in goroutine <ID>",
			"	<DIR>/cmd/envtool/testdata/panic_test.go:23 <ADDR>",
			"",
		}

		cleanup(actual)
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
		"TestPanic1",
		"TestSkip1",
		"TestWithSubtest",
	}
	assert.Equal(t, expected, actual)
}

func TestListTestFuncsWithRegex(t *testing.T) {
	tests := []struct {
		wantErr  assert.ErrorAssertionFunc
		name     string
		run      string
		skip     string
		expected []string
	}{
		{
			name: "NoRunNoSkip",
			run:  "",
			skip: "",
			expected: []string{
				"TestError1",
				"TestError2",
				"TestNormal1",
				"TestNormal2",
				"TestPanic1",
				"TestSkip1",
				"TestWithSubtest",
			},
			wantErr: assert.NoError,
		},
		{
			name: "Run",
			run:  "TestError",
			skip: "",
			expected: []string{
				"TestError1",
				"TestError2",
			},
			wantErr: assert.NoError,
		},
		{
			name: "Skip",
			run:  "",
			skip: "TestError",
			expected: []string{
				"TestNormal1",
				"TestNormal2",
				"TestPanic1",
				"TestSkip1",
				"TestWithSubtest",
			},
			wantErr: assert.NoError,
		},
		{
			name: "RunSkip",
			run:  "TestError",
			skip: "TestError2",
			expected: []string{
				"TestError1",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "RunSkipAll",
			run:      "TestError",
			skip:     "TestError",
			expected: []string{},
			wantErr:  assert.NoError,
		},
		{
			name: "InvalidRun",
			run:  "[",
			skip: "",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Contains(t, err.Error(), "error parsing regexp")
			},
		},
		{
			name: "InvalidSkip",
			run:  "",
			skip: "[",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Contains(t, err.Error(), "error parsing regexp")
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := listTestFuncsWithRegex("./testdata", tt.run, tt.skip)
			tt.wantErr(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFilterStringsByRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tests    []string
		include  *regexp.Regexp
		exclude  *regexp.Regexp
		expected []string
	}{
		{
			name:     "Empty",
			tests:    []string{},
			include:  nil,
			exclude:  nil,
			expected: []string{},
		},
		{
			name:     "Include",
			tests:    []string{"Test1", "Test2"},
			include:  regexp.MustCompile("Test1"),
			exclude:  nil,
			expected: []string{"Test1"},
		},
		{
			name:     "Exclude",
			tests:    []string{"Test1", "Test2"},
			include:  nil,
			exclude:  regexp.MustCompile("Test1"),
			expected: []string{"Test2"},
		},
		{
			name:     "IncludeExclude",
			tests:    []string{"Test1", "Test2"},
			include:  regexp.MustCompile("Test1"),
			exclude:  regexp.MustCompile("Test1"),
			expected: []string{},
		},
		{
			name:     "IncludeExclude2",
			tests:    []string{"Test1", "Test2"},
			include:  regexp.MustCompile("Test1"),
			exclude:  regexp.MustCompile("Test2"),
			expected: []string{"Test1"},
		},
		{
			name:     "NotMatch",
			tests:    []string{"Test1", "Test2"},
			include:  regexp.MustCompile("Test3"),
			exclude:  regexp.MustCompile("Test3"),
			expected: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := filterStringsByRegex(tt.tests, tt.include, tt.exclude)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestShardTestFuncs(t *testing.T) {
	t.Parallel()

	testFuncs, err := listTestFuncs(testutil.IntegrationDir)
	require.NoError(t, err)
	assert.Contains(t, testFuncs, "TestQueryCompatLimit")
	assert.Contains(t, testFuncs, "TestCursorsGetMoreCommand")

	t.Run("InvalidIndex", func(t *testing.T) {
		t.Parallel()

		_, _, err := shardTestFuncs(0, 3, testFuncs)
		assert.EqualError(t, err, "index must be greater than 0")

		_, _, err = shardTestFuncs(3, 3, testFuncs)
		assert.NoError(t, err)

		_, _, err = shardTestFuncs(4, 3, testFuncs)
		assert.EqualError(t, err, "cannot shard when index is greater than total (4 > 3)")
	})

	t.Run("InvalidTotal", func(t *testing.T) {
		t.Parallel()

		_, _, err := shardTestFuncs(3, 1000, testFuncs[:42])
		assert.EqualError(t, err, "cannot shard when total is greater than a number of test functions (1000 > 42)")
	})

	t.Run("Valid", func(t *testing.T) {
		t.Parallel()

		res, skip, err := shardTestFuncs(1, 3, testFuncs)
		require.NoError(t, err)
		assert.Equal(t, testFuncs[0], res[0])
		assert.NotEqual(t, testFuncs[1], res[1])
		assert.NotEqual(t, testFuncs[2], res[1])
		assert.Equal(t, testFuncs[3], res[1])
		assert.NotEmpty(t, skip)

		lastRes, lastSkip, err := shardTestFuncs(3, 3, testFuncs)
		require.NoError(t, err)
		assert.NotEqual(t, testFuncs[0], lastRes[0])
		assert.NotEqual(t, testFuncs[1], lastRes[0])
		assert.Equal(t, testFuncs[2], lastRes[0])
		assert.NotEmpty(t, lastSkip)

		assert.NotEqual(t, res, lastRes)
		assert.NotEqual(t, skip, lastSkip)
	})
}

func TestListTestFuncsWithSkip(t *testing.T) {
	t.Parallel()

	testFuncs, err := listTestFuncsWithRegex("testdata", "", "Skip")
	require.NoError(t, err)

	sort.Strings(testFuncs)

	res, skip, err := shardTestFuncs(1, 2, testFuncs)

	assert.Equal(t, []string{"TestError2", "TestNormal2", "TestWithSubtest"}, skip)
	assert.Equal(t, []string{"TestError1", "TestNormal1", "TestPanic1"}, res)
	assert.Nil(t, err)

	lastRes, lastSkip, err := shardTestFuncs(3, 3, testFuncs)
	assert.Equal(t, []string{"TestNormal1", "TestWithSubtest"}, lastRes)
	assert.Equal(t, []string{"TestError1", "TestError2", "TestNormal2", "TestPanic1"}, lastSkip)
	require.NoError(t, err)
}

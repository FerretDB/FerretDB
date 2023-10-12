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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type action string

//nolint:godot // false positive for unexported identifiers
const (
	// actionRun means the test has started running.
	actionRun action = "run"
	// actionPause means the test has been paused.
	actionPause action = "pause"
	// actionCont means the test has continued running.
	actionCont action = "cont"
	// actionPass means the test passed.
	actionPass action = "pass"
	// actionBench means the benchmark printed log output but did not fail.
	actionBench action = "bench"
	// actionFail means the test or benchmark failed.
	actionFail action = "fail"
	// actionOutput means the test printed output.
	actionOutput action = "output"
	// actionSkip means the test was skipped or the package contained no tests.
	actionSkip action = "skip"
)

// Status represents the status of a single test.
type Status string

// Constants representing different test statuses.
const (
	Fail    Status = "fail"
	Skip    Status = "skip"
	Pass    Status = "pass"
	Ignore  Status = "ignore"  // for fluky tests
	Unknown Status = "unknown" // result can't be parsed
)

// TestResults represents the collection of results from multiple tests.
type TestResults struct {
	// Test results by full test name.
	TestResults map[string]TestResult
}

// TestResult represents the outcome of a single test.
type TestResult struct {
	Status Status
	Output string
}

type testEvent struct {
	Time           time.Time `json:"Time"`
	Action         action    `json:"Action"`
	Package        string    `json:"Package"`
	Test           string    `json:"Test"`
	Output         string    `json:"Output"`
	ElapsedSeconds float64   `json:"Elapsed"`
}

func (te testEvent) Elapsed() time.Duration {
	return time.Duration(te.ElapsedSeconds * float64(time.Second))
}

func extractRootTestName(fullTestName string) string {
	parts := strings.Split(fullTestName, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return fullTestName
}

// testsRun runs tests specified by the shard index and total or by the run regex
// using `go test` with given extra args.
func testsRun(w io.Writer, index, total uint, run string, args []string) error {
	zap.S().Debugf("testsRun: index=%d, total=%d, run=%q, args=%q", index, total, run, args)

	var totalTest int
	if run == "" {
		if index == 0 || total == 0 {
			return fmt.Errorf("--shard-index and --shard-total must be specified when --run is not")
		}

		all, err := listTests("")
		if err != nil {
			return lazyerrors.Error(err)
		}

		shard, err := shardTests(index, total, all)
		if err != nil {
			return lazyerrors.Error(err)
		}

		run = "^("

		for i, t := range shard {
			run += t
			if i != len(shard)-1 {
				run += "|"
			}
		}

		totalTest = len(shard)
		run += ")$"
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3055

	args = append([]string{"test", "-json", "-run=" + run}, args...)

	return runTest("go", args, totalTest, zap.S())
}

// listTests returns a sorted slice of all tests in the specified directory and subdirectories.
func listTests(dir string) ([]string, error) {
	var buf bytes.Buffer

	cmd := exec.Command("go", "test", "-list=.", "./...")
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	tests := make(map[string]struct{}, 200)

	s := bufio.NewScanner(&buf)
	for s.Scan() {
		l := s.Text()

		switch {
		case strings.HasPrefix(l, "Test"):
		case strings.HasPrefix(l, "Benchmark"):
		case strings.HasPrefix(l, "Example"):
		case strings.HasPrefix(l, "Fuzz"):
		case strings.HasPrefix(l, "? "):
			continue
		case strings.HasPrefix(l, "ok "):
			continue
		default:
			return nil, fmt.Errorf("can't parse line %q", l)
		}

		if _, dup := tests[l]; dup {
			return nil, fmt.Errorf("duplicate test name %q", l)
		}

		tests[l] = struct{}{}
	}

	if err := s.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := maps.Keys(tests)
	sort.Strings(res)

	return res, nil
}

// shardTests shards given test names.
func shardTests(index, total uint, tests []string) ([]string, error) {
	if index == 0 {
		return nil, fmt.Errorf("index must be greater than 0")
	}

	if total == 0 {
		return nil, fmt.Errorf("total must be greater than 0")
	}

	if index > total {
		return nil, fmt.Errorf("cannot shard when index is greater to total (%d > %d)", index, total)
	}

	testsLen := uint(len(tests))
	if total > testsLen {
		return nil, fmt.Errorf("cannot shard when total is greater than amount of tests (%d > %d)", total, testsLen)
	}

	res := make([]string, 0, testsLen/total)
	var test uint
	shard := uint(1)

	// use different shards for tests with similar names for better load balancing
	for {
		if test == testsLen {
			return res, nil
		}

		if index == shard {
			res = append(res, tests[test])
		}

		test++
		shard = shard%total + 1
	}
}

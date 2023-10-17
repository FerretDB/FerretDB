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
	"context"
	"encoding/json"
	"errors"
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

// testEvent represents a single even emitted by `go test -json`.
//
// See https://pkg.go.dev/cmd/test2json#hdr-Output_Format.
type testEvent struct {
	Time           time.Time `json:"Time"`
	Action         string    `json:"Action"`
	Package        string    `json:"Package"`
	Test           string    `json:"Test"`
	Output         string    `json:"Output"`
	ElapsedSeconds float64   `json:"Elapsed"`
}

// Elapsed returns an elapsed time.
func (te testEvent) Elapsed() time.Duration {
	return time.Duration(te.ElapsedSeconds * float64(time.Second))
}

// testResult represents the outcome of a single test.
type testResult struct {
	output string
}

// topLevelName returns a top-level test function name.
func topLevelName(fullTestName string) string {
	res, _, _ := strings.Cut(fullTestName, "/")
	return res
}

// runGoTest runs `go test` with given extra args.
func runGoTest(ctx context.Context, args []string, total int, logger *zap.SugaredLogger) error {
	cmd := exec.CommandContext(ctx, "go", append([]string{"test", "-json"}, args...)...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stderr = os.Stderr

	p, err := cmd.StdoutPipe()
	if err != nil {
		return lazyerrors.Error(err)
	}

	if err = cmd.Start(); err != nil {
		return lazyerrors.Error(err)
	}

	defer cmd.Cancel() //nolint:errcheck // safe to ignore

	var done int
	results := make(map[string]*testResult, 250)

	d := json.NewDecoder(p)
	d.DisallowUnknownFields()

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				return cmd.Wait()
			}

			return lazyerrors.Error(err)
		}

		// logger.Desugar().Debug("decoded event", zap.Any("event", event))

		res := results[event.Test]
		if res == nil {
			res = new(testResult)
			results[event.Test] = res
		}

		res.output += event.Output

		topLevel := topLevelName(event.Test) == event.Test

		switch event.Action {
		case "pass": // the test passed
			fallthrough

		case "fail": // the test or benchmark failed
			fallthrough

		case "skip": // the test was skipped or the package contained no tests
			msg := strings.ToTitle(event.Action)

			if event.Test == "" {
				msg += " " + event.Package
			} else {
				msg += " " + event.Test
			}

			if topLevel {
				if event.Test != "" {
					done++
				}
				msg += fmt.Sprintf(" (%d/%d)", done, total)
			}

			logger.Info(msg)

		case "start": // the test binary is about to be executed
		case "run": // the test has started running
		case "pause": // the test has been paused
		case "cont": // the test has continued running
		case "output": // the test printed output
		case "bench": // the benchmark printed log output but did not fail

		default:
			return lazyerrors.Errorf("unknown action %q", event.Action)
		}
	}
}

// testsRun runs tests specified by the shard index and total or by the run regex
// using `go test` with given extra args.
func testsRun(index, total uint, run string, args []string, logger *zap.SugaredLogger) error {
	logger.Debugf("testsRun: index=%d, total=%d, run=%q, args=%q", index, total, run, args)

	var totalTest int
	if run == "" {
		if index == 0 || total == 0 {
			return fmt.Errorf("--shard-index and --shard-total must be specified when --run is not")
		}

		all, err := listTestFuncs("")
		if err != nil {
			return lazyerrors.Error(err)
		}

		shard, err := shardTestFuncs(index, total, all)
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

	return runGoTest(context.TODO(), append([]string{"-run=" + run}, args...), totalTest, logger)
}

// listTestFuncs returns a sorted slice of all top-level test functions (tests, benchmarks, examples, fuzz functions)
// in the specified directory and subdirectories.
func listTestFuncs(dir string) ([]string, error) {
	var buf bytes.Buffer

	cmd := exec.Command("go", "test", "-list=.", "./...")
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	testFuncs := make(map[string]struct{}, 300)

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

		if _, dup := testFuncs[l]; dup {
			return nil, fmt.Errorf("duplicate test function name %q", l)
		}

		testFuncs[l] = struct{}{}
	}

	if err := s.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := maps.Keys(testFuncs)
	sort.Strings(res)

	return res, nil
}

// shardTestFuncs shards given top-level test functions.
func shardTestFuncs(index, total uint, testFuncs []string) ([]string, error) {
	if index == 0 {
		return nil, fmt.Errorf("index must be greater than 0")
	}

	if total == 0 {
		return nil, fmt.Errorf("total must be greater than 0")
	}

	if index > total {
		return nil, fmt.Errorf("cannot shard when index is greater than total (%d > %d)", index, total)
	}

	l := uint(len(testFuncs))
	if total > l {
		return nil, fmt.Errorf("cannot shard when total is greater than a number of test functions (%d > %d)", total, l)
	}

	res := make([]string, 0, l/total+1)
	shard := uint(1)

	// use different shards for tests with similar names for better load balancing
	for _, test := range testFuncs {
		if index == shard {
			res = append(res, test)
		}

		shard = shard%total + 1
	}

	return res, nil
}

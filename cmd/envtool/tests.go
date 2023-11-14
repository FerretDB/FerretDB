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
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	run     time.Time
	cont    time.Time
	outputs []string
}

// levelTest returns test level (starting from 0) for the given (sub)test name.
func levelTest(testName string) int {
	if testName == "" {
		panic("empty test name")
	}

	return strings.Count(testName, "/")
}

// parentTest returns parent test name for the given subtest, or empty string.
func parentTest(testName string) string {
	if i := strings.LastIndex(testName, "/"); i >= 0 {
		return testName[:i]
	}

	return ""
}

// runGoTest runs `go test` with given extra args.
func runGoTest(ctx context.Context, args []string, total int, times bool, logger *zap.SugaredLogger) error {
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
	results := make(map[string]*testResult, 300)
	contexts := make(map[string]context.Context)
	spans := make(map[string]trace.Span)

	d := json.NewDecoder(p)
	d.DisallowUnknownFields()

	totalTests := "?"
	if total > 0 {
		totalTests = strconv.Itoa(total)
	}

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				return cmd.Wait()
			}

			return lazyerrors.Error(err)
		}

		// logger.Desugar().Debug("decoded event", zap.Any("event", event))

		for t := event.Test; t != ""; t = parentTest(t) {
			res := results[t]
			if res == nil {
				res = &testResult{
					outputs: make([]string, 0, 2),
				}
				results[t] = res
			}

			if out := strings.TrimSpace(event.Output); out != "" {
				res.outputs = append(res.outputs, strings.Repeat("  ", levelTest(t)+1)+out)
			}
		}

		switch event.Action {
		case "start": // the test binary is about to be executed
			// nothing

		case "run": // the test has started running
			// Start a new span for the test or subtest
			var newCtx context.Context
			var span trace.Span

			if levelTest(event.Test) == 0 {
				newCtx, span = otel.Tracer("").Start(ctx, event.Test)
			} else {
				parentCtx := contexts[parentTest(event.Test)]
				newCtx, span = otel.Tracer("").Start(parentCtx, event.Test)
			}
			contexts[event.Test] = newCtx
			results[event.Test].run = event.Time
			results[event.Test].cont = event.Time

			// Add an event to the span at the start of the test.
			span.AddEvent("Test Started", trace.WithTimestamp(event.Time), trace.WithAttributes(
				attribute.String("test.name", event.Test),
				attribute.String("test.action", event.Action),
			))
			spans[event.Test] = span

		case "pause": // the test has been paused
			// nothing

		case "cont": // the test has continued running
			results[event.Test].cont = event.Time
			span := spans[event.Test]
			span.AddEvent("Test Continued", trace.WithTimestamp(event.Time), trace.WithAttributes(
				attribute.String("test.name", event.Test),
				attribute.String("test.action", event.Action),
			))

		case "output": // the test printed output
			// nothing

		case "bench": // the benchmark printed log output but did not fail
			// nothing

		case "pass": // the test passed
			// End the span for the test or subtest.
			if span, ok := spans[event.Test]; ok {
				span.End(trace.WithTimestamp(event.Time))
				delete(spans, event.Test)
			}
			fallthrough

		case "fail": // the test or benchmark failed
			// End the span for the test or subtest.
			if span, ok := spans[event.Test]; ok {
				span.AddEvent("Test Failed", trace.WithTimestamp(event.Time), trace.WithAttributes(
					attribute.String("test.name", event.Test),
					attribute.String("test.action", event.Action),
				))
				span.End(trace.WithTimestamp(event.Time))
				delete(spans, event.Test)
			}
			fallthrough

		case "skip": // the test was skipped or the package contained no tests
			if event.Test == "" {
				logger.Info(strings.ToTitle(event.Action) + " " + event.Package)
				continue
			}

			top := levelTest(event.Test) == 0
			if !top && event.Action == "pass" {
				continue
			}

			res := results[event.Test]

			msg := strings.ToTitle(event.Action) + " " + event.Test

			if times {
				msg += fmt.Sprintf(" (%.2fs", event.Time.Sub(res.cont).Seconds())

				if res.run != res.cont {
					msg += fmt.Sprintf("/%.2fs", event.Time.Sub(res.run).Seconds())
				}

				if event.ElapsedSeconds > 0 {
					msg += fmt.Sprintf("/%.2fs", event.ElapsedSeconds)
				}

				msg += ")"
			}

			if top {
				done++
				msg += fmt.Sprintf(" %d/%s", done, totalTests)
			}

			if event.Action == "pass" {
				logger.Info(msg)
				continue
			}

			msg += ":"
			logger.Info(msg)

			for _, l := range results[event.Test].outputs {
				logger.Info(l)
			}

			logger.Info("")

			// End the span for the test or subtest.
			if span, ok := spans[event.Test]; ok {
				span.End(trace.WithTimestamp(event.Time))
				delete(spans, event.Test)
			}

		default:
			return lazyerrors.Errorf("unknown action %q", event.Action)
		}
	}
}

// testsRun runs tests specified by the shard index and total or by the run regex
// using `go test` with given extra args.
func testsRun(ctx context.Context, index, total uint, run string, args []string, logger *zap.SugaredLogger) error {
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

	return runGoTest(ctx, append([]string{"-run=" + run}, args...), totalTest, true, logger)
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

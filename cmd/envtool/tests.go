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
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
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
	ctx        context.Context
	run        time.Time
	cont       time.Time
	lastAction string
	outputs    []string
}

// parentTest returns parent test name for the given subtest, or empty string.
func parentTest(testName string) string {
	if i := strings.LastIndex(testName, "/"); i >= 0 {
		return testName[:i]
	}

	return ""
}

// resultKey returns a key for the given package and test name.
func resultKey(packageName, testName string) string {
	must.NotBeZero(packageName)

	if testName == "" {
		return packageName
	}

	return packageName + "." + testName
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

	// Keys are:
	// - "package/name"
	// - "package/name.TestName"
	// - "package/name.TestName/subtest"
	//
	// See [resultKey].
	results := make(map[string]*testResult, 300)

	var done int

	d := json.NewDecoder(p)
	d.DisallowUnknownFields()

	totalTests := "?"
	if total > 0 {
		totalTests = strconv.Itoa(total)
	}

	var root oteltrace.Span
	ctx, root = otel.Tracer("").Start(ctx, "run")

	defer root.End()

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if !errors.Is(err, io.EOF) {
				return lazyerrors.Error(err)
			}

			break
		}

		// logger.Desugar().Info("decoded event", zap.Any("event", event))

		must.NotBeZero(event.Package)

		res := results[resultKey(event.Package, event.Test)]
		if res == nil {
			res = new(testResult)
			results[resultKey(event.Package, event.Test)] = res

			attributes := []otelattribute.KeyValue{
				otelattribute.String("package", event.Package),
				otelattribute.String("test", event.Test),
			}

			var parentCtx context.Context
			var spanName string

			if event.Test == "" {
				parentCtx = ctx
				spanName = event.Package
			} else {
				parentCtx = results[resultKey(event.Package, parentTest(event.Test))].ctx
				spanName = event.Test
			}

			must.NotBeZero(parentCtx)
			must.NotBeZero(spanName)

			res.ctx, _ = otel.Tracer("").Start(parentCtx, spanName, oteltrace.WithAttributes(attributes...))
			res.outputs = make([]string, 0, 2)
		}

		res.lastAction = event.Action

		switch event.Action {
		case "start": // the test binary is about to be executed
			oteltrace.SpanFromContext(res.ctx).AddEvent(event.Action)

		case "run": // the test has started running
			must.NotBeZero(event.Test)

			oteltrace.SpanFromContext(res.ctx).AddEvent(event.Action)

			res.run = event.Time
			res.cont = event.Time

		case "pause": // the test has been paused
			oteltrace.SpanFromContext(res.ctx).AddEvent(event.Action)

		case "cont": // the test has continued running
			must.NotBeZero(event.Test)

			oteltrace.SpanFromContext(res.ctx).AddEvent(event.Action)

			res.cont = event.Time

		case "output": // the test printed output
			// do not add span event

			out := strings.TrimSuffix(event.Output, "\n")

			// initial setup output or early panic
			if event.Test == "" {
				logger.Info(out)
				continue
			}

			res.outputs = append(res.outputs, out)

		case "bench": // the benchmark printed log output but did not fail
			// do not add span event

		case "pass": // the test passed
			fallthrough

		case "fail": // the test or benchmark failed
			fallthrough

		case "skip": // the test was skipped or the package contained no tests
			code := otelcodes.Ok
			if event.Action == "fail" {
				code = otelcodes.Error
			}

			testSpan := oteltrace.SpanFromContext(res.ctx)
			testSpan.AddEvent(event.Action)
			testSpan.SetStatus(code, event.Action)
			testSpan.End()

			if event.Test == "" {
				logger.Info(strings.ToTitle(event.Action) + " " + event.Package)
				continue
			}

			top := parentTest(event.Test) == ""
			if !top && event.Action == "pass" {
				continue
			}

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
			logger.Warn(msg)

			for _, l := range res.outputs {
				logger.Warn(l)
			}

			logger.Warn("")

		default:
			return lazyerrors.Errorf("unknown action %q", event.Action)
		}
	}

	var unfinished []string

	for t, res := range results {
		switch res.lastAction {
		case "pass", "fail", "skip":
			continue
		}

		unfinished = append(unfinished, t)
	}

	if unfinished == nil {
		return cmd.Wait()
	}

	slices.Sort(unfinished)

	logger.Error("")

	logger.Error("Some tests did not finish:")

	for _, t := range unfinished {
		logger.Errorf("  %s", t)
	}

	logger.Error("")

	// On panic, the last event will not be "fail"; see https://github.com/golang/go/issues/38382.
	// Try to provide the best possible output in that case.

	var panicked string

	for _, t := range unfinished {
		if !slices.ContainsFunc(results[t].outputs, func(s string) bool {
			return strings.Contains(s, "panic: ")
		}) {
			continue
		}

		if panicked != "" {
			break
		}

		panicked = t
	}

	for _, t := range unfinished {
		if panicked != "" && t != panicked {
			continue
		}

		logger.Errorf("%s:", t)

		for _, l := range results[t].outputs {
			logger.Error(l)
		}

		logger.Error("")
	}

	return cmd.Wait()
}

// testsRun runs tests specified by the shard index and total or by the run regex
// using `go test` with given extra args.
func testsRun(ctx context.Context, index, total uint, run, skip string, args []string, logger *zap.SugaredLogger) error {
	logger.Debugf("testsRun: index=%d, total=%d, run=%q, args=%q", index, total, run, args)

	if run == "" && (index == 0 || total == 0) {
		return fmt.Errorf("--shard-index and --shard-total must be specified when --run is not")
	}

	tests, err := listTestFuncsWithRegex("", run, skip)
	if err != nil {
		return lazyerrors.Error(err)
	}

	// Then, shard all the tests but only run the ones that match the regex and that should
	// be run on the specific shard.
	shard, skipShard, err := shardTestFuncs(index, total, tests)
	if err != nil {
		return lazyerrors.Error(err)
	}

	args = append(args, "-run="+run)

	if len(skipShard) > 0 {
		if skip != "" {
			skip += "|"
		}
		skip += "^(" + strings.Join(skipShard, "|") + ")$"
	}

	if skip != "" {
		args = append(args, "-skip="+skip)
	}

	ot, err := observability.NewOtelTracer(&observability.OtelTracerOpts{
		Logger:   logger.Desugar(),
		Service:  "envtool-tests",
		Endpoint: "127.0.0.1:4318",
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	go ot.Run(ctx)

	return runGoTest(ctx, args, len(shard), true, logger)
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
			// testutil.DatabaseName and other helpers depend on test names being unique across packages
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

// listTestFuncsWithRegex returns regex-filtered names of all top-level test
// functions (tests, benchmarks, examples, fuzz functions) in the specified
// directory and subdirectories.
func listTestFuncsWithRegex(dir, run, skip string) ([]string, error) {
	tests, err := listTestFuncs(dir)
	if err != nil {
		return nil, err
	}

	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests to run")
	}

	var (
		rxRun  *regexp.Regexp
		rxSkip *regexp.Regexp
	)

	if run != "" {
		rxRun, err = regexp.Compile(run)
		if err != nil {
			return nil, err
		}
	}

	if skip != "" {
		rxSkip, err = regexp.Compile(skip)
		if err != nil {
			return nil, err
		}
	}

	return filterStringsByRegex(tests, rxRun, rxSkip), nil
}

// filterStringsByRegex filters a slice of strings based on inclusion and exclusion
// criteria defined by regular expressions.
func filterStringsByRegex(tests []string, include, exclude *regexp.Regexp) []string {
	res := []string{}

	for _, test := range tests {
		if exclude != nil && exclude.MatchString(test) {
			continue
		}

		if include != nil && !include.MatchString(test) {
			continue
		}

		res = append(res, test)
	}

	return res
}

// shardTestFuncs shards given top-level test functions.
// It returns a slice of test functions to run and what test functions to skip for the given shard.
func shardTestFuncs(index, total uint, testFuncs []string) (run, skip []string, err error) {
	if index == 0 {
		return nil, nil, fmt.Errorf("index must be greater than 0")
	}

	if total == 0 {
		return nil, nil, fmt.Errorf("total must be greater than 0")
	}

	if index > total {
		return nil, nil, fmt.Errorf("cannot shard when index is greater than total (%d > %d)", index, total)
	}

	l := uint(len(testFuncs))
	if total > l {
		return nil, nil, fmt.Errorf("cannot shard when total is greater than a number of test functions (%d > %d)", total, l)
	}

	run = make([]string, 0, l/total+1)
	skip = make([]string, 0, len(testFuncs)-len(run))
	shard := uint(1)

	// use different shards for tests with similar names for better load balancing
	for _, test := range testFuncs {
		if index == shard {
			run = append(run, test)
		} else {
			skip = append(skip, test)
		}

		shard = shard%total + 1
	}

	return run, skip, nil
}

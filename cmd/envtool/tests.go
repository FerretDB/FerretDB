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
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
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

// listTestFuncs returns a sorted slice of all top-level test functions (tests, benchmarks, examples, fuzz functions)
// matching given regular expression in the specified directory and subdirectories.
func listTestFuncs(ctx context.Context, dir, re string, logger *slog.Logger) ([]string, error) {
	ctx, span := otel.Tracer("").Start(ctx, "listTestFuncs")
	defer span.End()

	var buf bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "test", "-list="+re, "./...")
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	logger.InfoContext(ctx, fmt.Sprintf("Running %s", strings.Join(cmd.Args, " ")))

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

	return slices.Sorted(maps.Keys(testFuncs)), nil
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

// testArgs handles `envtool tests run` arguments and returns a slice of `go test` arguments.
func testArgs(ctx context.Context, dir string, index, total uint, run, skip string, logger *slog.Logger) ([]string, uint, error) {
	ctx, span := otel.Tracer("").Start(ctx, "testArgs")
	defer span.End()

	listRE := "."

	if run != "" || skip != "" {
		if index != 0 || total != 0 {
			return nil, 0, fmt.Errorf("--run or --skip can't be used together with --shard-index or --shard-total")
		}

		// don't try to handle subtests or -skip ourselves
		if strings.Contains(run, "/") || skip != "" {
			res := make([]string, 0, 2)

			if run != "" {
				res = append(res, "-run="+run)
			}

			if skip != "" {
				res = append(res, "-skip="+skip)
			}

			return res, 0, nil
		}

		listRE = run
	}

	testFuncs, err := listTestFuncs(ctx, dir, listRE, logger)
	if err != nil {
		return nil, 0, err
	}

	if index != 0 || total != 0 {
		testFuncs, err = shardTestFuncs(index, total, testFuncs)
		if err != nil {
			return nil, 0, err
		}
	}

	return []string{"-run=^(" + strings.Join(testFuncs, "|") + ")$"}, uint(len(testFuncs)), nil
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

// startGoTest starts `go test` with given arguments.
func startGoTest(ctx context.Context, args []string, raw string, l *slog.Logger) (*json.Decoder, func() error, error) {
	var cleanups []func() error

	cleanup := func() error {
		var err error

		for _, f := range slices.Backward(cleanups) {
			if e := f(); e != nil {
				l.DebugContext(ctx, fmt.Sprintf("Cleanup error: %s", e))

				if err == nil {
					err = lazyerrors.Error(e)
				}
			}
		}

		return err
	}

	cmd := exec.CommandContext(ctx, "go", append([]string{"test", "-json"}, args...)...)

	l.InfoContext(ctx, fmt.Sprintf("Running %s", strings.Join(cmd.Args, " ")))

	cmd.Stderr = os.Stderr

	if raw != "" {
		f, e := os.Create(raw + "-stderr.txt.tmp")
		if e != nil {
			_ = cleanup()
			return nil, nil, lazyerrors.Error(e)
		}

		cleanups = append(cleanups, f.Close)

		cmd.Stderr = io.MultiWriter(f, os.Stderr)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = cleanup()
		return nil, nil, lazyerrors.Error(err)
	}

	d := json.NewDecoder(stdout)

	if raw != "" {
		f, e := os.Create(raw + "-stdout.json.tmp")
		if e != nil {
			_ = cleanup()
			return nil, nil, lazyerrors.Error(e)
		}

		cleanups = append(cleanups, f.Close)

		d = json.NewDecoder(io.TeeReader(stdout, f))
	}

	d.DisallowUnknownFields()

	if err = cmd.Start(); err != nil {
		_ = cleanup()
		return nil, nil, lazyerrors.Error(err)
	}

	cleanups = append(cleanups, cmd.Wait)

	return d, cleanup, nil
}

// runGoTestOpts represents [runGoTest] options.
//
//nolint:vet // for readability
type runGoTestOpts struct {
	args      []string
	total     uint
	times     bool
	rawPrefix string
	logger    *slog.Logger
}

// runGoTest runs `go test` with given extra args.
func runGoTest(runCtx context.Context, opts *runGoTestOpts) (resErr error) {
	totalTests := "?"
	if opts.total > 0 {
		totalTests = strconv.Itoa(int(opts.total))
	}

	runCtx, runSpan := otel.Tracer("").Start(runCtx, "run")
	runSpan.SetAttributes(otelattribute.String("db.ferretdb.envtool.total_tests", totalTests))
	defer runSpan.End()

	d, cleanup, err := startGoTest(runCtx, opts.args, opts.rawPrefix, opts.logger)
	if err != nil {
		resErr = lazyerrors.Error(err)
		return
	}

	defer func() {
		if e := cleanup(); resErr == nil && e != nil {
			resErr = lazyerrors.Error(e)
		}
	}()

	runSpan.AddEvent("started")

	// Keys are:
	// - "package/name"
	// - "package/name.TestName"
	// - "package/name.TestName/subtest"
	//
	// See [resultKey].
	results := make(map[string]*testResult, 300)

	var done int
	var firstEvent bool

	for {
		var event testEvent
		if err = d.Decode(&event); err != nil {
			if !errors.Is(err, io.EOF) {
				resErr = lazyerrors.Error(err)
				return
			}

			break
		}

		if !firstEvent {
			runSpan.AddEvent("first event")
			firstEvent = true
		}

		must.NotBeZero(event.Package)

		key := resultKey(event.Package, event.Test)
		res := results[key]
		if res == nil {
			res = new(testResult)
			results[key] = res

			var parentCtx context.Context
			var spanName string

			if event.Test == "" {
				parentCtx = runCtx
				spanName = event.Package
			} else {
				key = resultKey(event.Package, parentTest(event.Test))
				parent := results[key]

				// TODO https://github.com/FerretDB/FerretDB/issues/4465
				if parent == nil {
					panic(fmt.Sprintf(
						"no parent test found: package=%q, test=%q, key=%q", event.Package, event.Test, key,
					))
				}

				parentCtx = parent.ctx
				spanName = event.Test
			}

			must.NotBeZero(parentCtx)
			must.NotBeZero(spanName)

			attributes := []otelattribute.KeyValue{
				otelattribute.String("envtool.package", event.Package),
				otelattribute.String("envtool.test", event.Test),
			}

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
				opts.logger.InfoContext(runCtx, out)
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
			testSpan := oteltrace.SpanFromContext(res.ctx)
			if event.Action == "fail" {
				testSpan.SetStatus(otelcodes.Error, event.Action)
			}

			testSpan.AddEvent(event.Action)
			testSpan.End()

			if event.Test == "" {
				opts.logger.InfoContext(runCtx, strings.ToTitle(event.Action)+" "+event.Package)
				continue
			}

			top := parentTest(event.Test) == ""
			if !top && event.Action == "pass" {
				continue
			}

			msg := strings.ToTitle(event.Action) + " " + event.Test

			if opts.times {
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

			logOutput := opts.logger.WarnContext

			if event.Action == "pass" {
				if !opts.logger.Enabled(runCtx, slog.LevelDebug) {
					opts.logger.InfoContext(runCtx, msg)
					continue
				}

				logOutput = opts.logger.DebugContext
			}

			msg += ":"
			logOutput(runCtx, msg)

			for _, l := range res.outputs {
				logOutput(runCtx, l)
			}

			logOutput(runCtx, "")

		default:
			resErr = lazyerrors.Errorf("unknown action %q", event.Action)
			return
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
		return
	}

	slices.Sort(unfinished)

	opts.logger.ErrorContext(runCtx, "")

	opts.logger.ErrorContext(runCtx, "Some tests did not finish:")

	for _, t := range unfinished {
		opts.logger.ErrorContext(runCtx, fmt.Sprintf("  %s", t))
	}

	opts.logger.ErrorContext(runCtx, "")

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

		opts.logger.ErrorContext(runCtx, fmt.Sprintf("%s:", t))

		for _, l := range results[t].outputs {
			opts.logger.ErrorContext(runCtx, l)
		}

		opts.logger.ErrorContext(runCtx, "")
	}

	return
}

// testsRun runs tests specified by the shard index and total or by the run regex
// using `go test` with given extra args.
func testsRun(ctx context.Context, params *TestsRunParams, logger *slog.Logger) error {
	logger.DebugContext(ctx, fmt.Sprintf("testsRun: %+v", params))

	ot, err := observability.NewOTelTraceExporter(&observability.OTelTraceExporterOpts{
		Logger:  logger,
		Service: "envtool-tests",
		URL:     "http://127.0.0.1:4318/v1/traces",
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	ctx, span := otel.Tracer("").Start(ctx, "testsRun")

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		ot.Run(ctx)
		close(done)
	}()

	defer func() {
		span.End()
		cancel()
		<-done
	}()

	args, total, err := testArgs(ctx, "", params.ShardIndex, params.ShardTotal, params.Run, params.Skip, logger)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if len(params.Args) > 0 && params.Args[0] == "--" {
		params.Args = params.Args[1:]
	}

	args = append(args, params.Args...)

	err = runGoTest(ctx, &runGoTestOpts{
		args:      args,
		total:     total,
		times:     true,
		rawPrefix: params.RawPrefix,
		logger:    logger,
	})

	return err
}

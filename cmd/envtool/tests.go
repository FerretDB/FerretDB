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
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// testsShard shards integration test names.
func testsShard(w io.Writer, index, total uint) error {
	all, err := getAllTestNames("integration")
	if err != nil {
		return err
	}

	sharded, err := shardTests(index, total, all)
	if err != nil {
		return err
	}

	fmt.Fprint(w, "^(")
	for i, t := range sharded {
		fmt.Fprint(w, t)
		if i != len(sharded)-1 {
			fmt.Fprint(w, "|")
		}
	}
	fmt.Fprint(w, ")$")

	return nil
}

// getAllTestNames returns a sorted slice of all tests in the specified directory and subdirectories.
func getAllTestNames(dir string) ([]string, error) {
	cmd := exec.Command("go", "test", "-list=.", "./...")
	cmd.Dir = dir

	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	tests := make(map[string]struct{})

	s := bufio.NewScanner(bytes.NewReader(b))
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
		return nil, err
	}

	res := maps.Keys(tests)
	sort.Strings(res)

	return res, nil
}

// shardTests shards given test names.
func shardTests(index, total uint, tests []string) ([]string, error) {
	if index >= total {
		return nil, fmt.Errorf("cannot shard when index is greater or equal to total (%d >= %d)", index, total)
	}

	testsLen := uint(len(tests))
	if total > testsLen {
		return nil, fmt.Errorf("cannot shard when Total is greater than amount of tests (%d > %d)", total, testsLen)
	}

	// use different shards for tests with similar names for better load balancing
	res := make([]string, 0, testsLen/total)
	var test, shard uint
	for {
		if test == testsLen {
			return res, nil
		}

		if index == shard {
			res = append(res, tests[test])
		}

		test++
		shard = (shard + 1) % total
	}
}

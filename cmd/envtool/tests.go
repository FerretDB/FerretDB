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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

// shardIntegrationTests shards integration test names from the specified directory.
func shardIntegrationTests(w io.Writer, index, total uint) error {
	var output strings.Builder

	tests, err := getAllTestNames("integration")
	if err != nil {
		return err
	}

	shardedTests, err := shardTests(index, total, tests...)
	if err != nil {
		return err
	}

	output.WriteString("^(")
	output.WriteString(strings.Join(shardedTests, "|"))
	output.WriteString(")$")

	w.Write([]byte(output.String()))

	return nil
}

// shardTests shards gotten test names.
func shardTests(index, total uint, tests ...string) ([]string, error) {
	if index >= total {
		return nil, fmt.Errorf("Cannot shard when Index is greater or equal to Total (%d >= %d)", index, total)
	}

	testsLen := uint(len(tests))
	if total > testsLen {
		return nil, fmt.Errorf("Cannot shard when Total is greater than amount of tests (%d > %d)", total, testsLen)
	}

	sharder := testsLen / total
	start := sharder * index
	end := sharder*index + sharder

	if index == total-1 {
		modulo := testsLen % total
		end += modulo
	}

	return tests[start:end], nil
}

// getAllTestNames gets all test names from the specified directory.
func getAllTestNames(dir string) ([]string, error) {
	var tests []string

	newDir, err := getNewWorkingDir(dir)
	if err != nil {
		return tests, err
	}

	cmd := exec.Command("go", "test", "-list=.")
	cmd.Dir = newDir

	outputBytes, err := cmd.Output()
	if err != nil {
		return tests, err
	}
	output := string(outputBytes)

	tests = strings.FieldsFunc(output, func(s rune) bool {
		return s == '\n'
	})
	sort.Strings(tests)

	return tests, nil
}

// getNewWorkingDir gets the new working directory based on the repo root directory and passed destination.
func getNewWorkingDir(dest string) (string, error) {
	var rootDir, newWorkingDir string

	rootDir, err := getRootDir()
	if err != nil {
		return rootDir, err
	}

	newWorkingDir = path.Join(rootDir, dest)
	if _, err := os.Stat(newWorkingDir); err != nil {
		return newWorkingDir, err
	}

	return newWorkingDir, nil
}

// getRootDir gets the repository's root directory.
func getRootDir() (string, error) {
	var rootDir string

	workingDir, err := os.Getwd()
	if err != nil {
		return rootDir, err
	}

	if strings.Contains(workingDir, "envtool") {
		rootDir = path.Join(workingDir, "..", "..")
		return rootDir, nil
	}

	return workingDir, nil
}

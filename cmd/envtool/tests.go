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
	"io"
	"os"
	"os/exec"
	"strings"
)

// shardIntegrationTests shards integration test names from the specified path.
func shardIntegrationTests(w io.Writer, index, total uint) (string, error) {
	var output strings.Builder

	testNames, err := shardTests(index, total, "integration")
	if err != nil {
		return output.String(), err
	}
	shardedTestNames := testNames[index:total]

	output.WriteString("^(")
	output.WriteString(strings.Join(shardedTestNames, "|"))
	output.WriteString(")$")

	return output.String(), nil
}

// shardTests shards gotten test names from the specified path.
func shardTests(index, total uint, path string) ([]string, error) {
	testNames, err := getAllTestNames(path)
	if err != nil {
		return nil, err
	}

	return testNames[index:total], nil
}

// getAllTestNames gets all test names from the specified path.
func getAllTestNames(path string) ([]string, error) {
	var tests []string

	if err := chdirToRoot(); err != nil {
		return tests, err
	}

	if err := os.Chdir(path); err != nil {
		return tests, err
	}

	cmd := exec.Command("go", "test", "-list=.")

	outputBytes, err := cmd.Output()
	if err != nil {
		return tests, err
	}
	output := string(outputBytes)
	tests = strings.Split(output, "\n")

	return tests, nil
}

// chdirToRoot changes the directory to the root of the repository.
func chdirToRoot() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if strings.Contains(workingDir, "envtool") {
		os.Chdir("../..")
	}

	return nil
}

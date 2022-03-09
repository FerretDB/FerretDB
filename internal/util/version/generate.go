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

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const gitBin = "git"

func runGit(args []string, stdin io.Reader, stdout io.Writer, logger *zap.SugaredLogger) {
	if err := tryGit(args, stdin, stdout, logger); err != nil {
		logger.Fatal(err)
	}
}

func tryGit(args []string, stdin io.Reader, stdout io.Writer, logger *zap.SugaredLogger) error {
	cmd := exec.Command(gitBin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}

func main() {
	var wg sync.WaitGroup
	logger := zap.S().Named("git")

	// git describe --tags --dirty > version.txt
	{
		file := "version.txt"
		args := `describe --tags --dirty`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	// git rev-parse HEAD > commit.txt
	{
		file := "commit.txt"
		args := `rev-parse HEAD`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	// git branch --show-current > branch.txt
	{
		file := "branch.txt"
		args := `branch --show-current`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	wg.Wait()
}

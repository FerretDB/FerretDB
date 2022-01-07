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
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

func main() {
	logging.Setup(zap.InfoLevel)
	flag.Parse()

	debugF := flag.Bool("debug", false, "enable debug mode")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s target_directory\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *debugF {
		logging.Setup(zap.DebugLevel)
	}

	logger := zap.S()

	module, err := readModulePath()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Module path: %s.", module)

	cacheRoot, err := readCacheRoot(module)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Cache root: %s.", cacheRoot)

	root, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		logger.Fatal(err)
	}
	existingFiles, err := collectExistingFiles(root, logger)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d existing files.", len(existingFiles))

	cacheFiles, err := collectCacheFiles(cacheRoot, logger)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d cached files.", len(cacheFiles))

	diffFiles := diff(cacheFiles, existingFiles)
	logger.Infof("Copying %d files to corpus.", len(diffFiles))
	for _, p := range diffFiles {
		from := filepath.Join(cacheRoot, p)
		to := filepath.Join(root, p)
		logger.Debugf("%s -> %s", from, to)
		if err := copyFile(from, to); err != nil {
			logger.Fatal(err)
		}
	}
}

func readModulePath() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("failed to read build info")
	}

	parts := strings.Split(info.Main.Path, "/")
	if len(parts) != 4 {
		return "", fmt.Errorf("unexpected path format: %q", info.Main.Path)
	}

	return strings.Join(parts[:3], "/"), nil
}

func readCacheRoot(module string) (string, error) {
	b, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", err
	}
	return filepath.Join(strings.TrimSpace(string(b)), "fuzz", module, "internal"), nil
}

func collectExistingFiles(root string, logger *zap.SugaredLogger) (map[string]struct{}, error) {
	existingFiles := make(map[string]struct{}, 1000)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// skip .git, etc
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// skip .env, .DS_Store, etc
		if len(info.Name()) != 64 {
			return nil
		}

		path, err = filepath.Rel(root, path)
		if err != nil {
			return err
		}
		logger.Debug(path)
		existingFiles[path] = struct{}{}
		return nil
	})

	return existingFiles, err
}

func collectCacheFiles(cacheRoot string, logger *zap.SugaredLogger) (map[string]struct{}, error) {
	cacheFiles := make(map[string]struct{}, 1000)
	err := filepath.Walk(cacheRoot, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return lazyerrors.Error(err)
		}

		if info.IsDir() {
			// skip .git, etc
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// skip .env, .DS_Store, etc
		if len(info.Name()) != 64 {
			return nil
		}

		path, err = filepath.Rel(cacheRoot, path)
		if err != nil {
			return lazyerrors.Error(err)
		}
		logger.Debug(path)
		cacheFiles[path] = struct{}{}
		return nil
	})

	return cacheFiles, err
}

func diff(from, to map[string]struct{}) []string {
	res := make([]string, 0, 50)
	for p := range from {
		if _, ok := to[p]; !ok {
			res = append(res, p)
		}
	}
	return res
}

func copyFile(from, to string) error {
	fromF, err := os.Open(from)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer fromF.Close()

	toF, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = io.Copy(toF, fromF)
	if closeErr := toF.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		os.Remove(to)
		return lazyerrors.Error(err)
	}

	return nil
}

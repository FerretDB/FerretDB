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
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

// generatedCorpus returns $GOCACHE/fuzz/github.com/FerretDB/FerretDB/internal.
func generatedCorpus() (string, error) {
	b, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	gocache := strings.TrimSpace(string(b))
	return filepath.Join(gocache, "fuzz", "github.com", "FerretDB", "FerretDB", "internal"), nil
}

// collectFiles returns a map of all interesting files in the given directory.
func collectFiles(root string, logger *zap.SugaredLogger) (map[string]struct{}, error) {
	existingFiles := make(map[string]struct{}, 1000)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
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

		// skip .env, .DS_Store, README.md, etc
		if len(info.Name()) != 64 {
			return nil
		}

		path, err = filepath.Rel(root, path)
		if err != nil {
			return lazyerrors.Error(err)
		}
		logger.Debug(path)
		existingFiles[path] = struct{}{}
		return nil
	})

	return existingFiles, err
}

// diff returns the set of files in srd that are not in dst.
func diff(src, dst map[string]struct{}) []string {
	res := make([]string, 0, 50)
	for p := range src {
		if _, ok := dst[p]; !ok {
			res = append(res, p)
		}
	}

	sort.Strings(res)
	return res
}

// copyFile copies a file from src to dst, overwriting dst if it exists.
func copyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer srcF.Close()

	dir := filepath.Dir(dst)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o777)
	}
	if err != nil {
		return lazyerrors.Error(err)
	}

	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = io.Copy(dstF, srcF)
	if closeErr := dstF.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		os.Remove(dst)
		return lazyerrors.Error(err)
	}

	return nil
}

func main() {
	debugF := flag.Bool("debug", false, "enable debug mode")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s target_directory\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(), "one arguments expected")
		os.Exit(2)
	}

	logging.Setup(zap.InfoLevel)
	if *debugF {
		logging.Setup(zap.DebugLevel)
	}
	logger := zap.S()

	generatedCorpus, err := generatedCorpus()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Generated corpus: %s.", generatedCorpus)

	collectedCorpus, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Collected corpus: %s.", collectedCorpus)

	collectedCorpusFiles, err := collectFiles(collectedCorpus, logger)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d files in collected corpus.", len(collectedCorpusFiles))

	generatedCorpusFiles, err := collectFiles(generatedCorpus, logger)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Found %d files in generated corpus.", len(generatedCorpusFiles))

	files := diff(generatedCorpusFiles, collectedCorpusFiles)
	logger.Infof("Copying %d files to generated corpus.", len(files))
	for _, p := range files {
		from := filepath.Join(generatedCorpus, p)
		to := filepath.Join(collectedCorpus, p)
		logger.Debugf("%s -> %s", from, to)
		if err := copyFile(from, to); err != nil {
			logger.Fatal(err)
		}
	}
}

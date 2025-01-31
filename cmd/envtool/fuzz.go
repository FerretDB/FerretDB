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
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// copyFile copies a file from src to dst, overwriting dst if it exists.
func copyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer srcF.Close() //nolint:errcheck // we are only reading it

	dir := filepath.Dir(dst)

	_, err = os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) {
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
		_ = os.Remove(dst)
		return lazyerrors.Error(err)
	}

	return nil
}

// fuzzGeneratedCorpus returns $GOCACHE/fuzz/github.com/FerretDB/FerretDB/v2,
// ensuring that this directory exists.
func fuzzGeneratedCorpus() (string, error) {
	b, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	path := filepath.Join(string(bytes.TrimSpace(b)), "fuzz", "github.com", "FerretDB", "FerretDB", "v2")

	if _, err = os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(path, 0o777)
		}

		if err != nil {
			return "", lazyerrors.Error(err)
		}
	}

	return path, err
}

// fuzzCollectFiles returns a map of all fuzz files in the given directory.
func fuzzCollectFiles(root string, logger *slog.Logger) (map[string]struct{}, error) {
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

		// skip other files
		if _, err = hex.DecodeString(info.Name()); err != nil {
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

// fuzzCutTestdata returns s with "/testdata/fuzz" removed.
//
// That converts seed corpus entry like `internal/bson/testdata/fuzz/FuzzArray/HEX`
// to format used by generated and collected corpora `internal/bson/FuzzArray/HEX`.
func fuzzCutTestdata(s string) string {
	old := string(filepath.Separator) + filepath.Join("testdata", "fuzz")
	return strings.Replace(s, old, "", 1)
}

// fuzzDiff returns the set of files in src that are not in dst, with and without applying `fuzzCutTestdata`.
func fuzzDiff(src, dst map[string]struct{}) []string {
	res := make([]string, 0, 50)

	for p := range src {
		if _, ok := dst[p]; ok {
			continue
		}

		if _, ok := dst[fuzzCutTestdata(p)]; ok {
			continue
		}

		res = append(res, p)
	}

	slices.Sort(res)

	return res
}

// fuzzCopyCorpus copies all new corpus files from srcRoot to dstRoot.
func fuzzCopyCorpus(srcRoot, dstRoot string, logger *slog.Logger) error {
	logger.Info(fmt.Sprintf("Copying from %s to %s", srcRoot, dstRoot))

	srcFiles, err := fuzzCollectFiles(srcRoot, logger)
	if err != nil {
		return lazyerrors.Error(err)
	}

	logger.Info(fmt.Sprintf("Found %d files in src", len(srcFiles)))

	dstFiles, err := fuzzCollectFiles(dstRoot, logger)
	if err != nil {
		return lazyerrors.Error(err)
	}

	logger.Info(fmt.Sprintf("Found %d existing files in dst", len(dstFiles)))

	files := fuzzDiff(srcFiles, dstFiles)
	logger.Info(fmt.Sprintf("Copying new %d files to dst", len(files)))

	for _, p := range files {
		src := filepath.Join(srcRoot, p)
		dst := fuzzCutTestdata(filepath.Join(dstRoot, p))

		if err := copyFile(src, dst); err != nil {
			return lazyerrors.Errorf("failed to copy %s to %s: %w", src, dst, err)
		}

		logger.Debug(fmt.Sprintf("%s -> %s", src, dst))
	}

	return nil
}

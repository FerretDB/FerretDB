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
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

// generatedCorpus returns $GOCACHE/fuzz/github.com/FerretDB/FerretDB,
// ensuring that this directory exists.
func generatedCorpus() (string, error) {
	b, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	path := filepath.Join(string(bytes.TrimSpace(b)), "fuzz", "github.com", "FerretDB", "FerretDB")

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

// collectFiles returns a map of all fuzz files in the given directory.
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

// cutTestdata returns s with "/testdata/fuzz" removed.
//
// That converts seed corpus entry like `internal/bson/testdata/fuzz/FuzzArray/HEX`
// to format used by generated and collected corpora `internal/bson/FuzzArray/HEX`.
func cutTestdata(s string) string {
	old := string(filepath.Separator) + filepath.Join("testdata", "fuzz")
	return strings.Replace(s, old, "", 1)
}

// diff returns the set of files in src that are not in dst, with and without applying `cutTestdata`.
func diff(src, dst map[string]struct{}) []string {
	res := make([]string, 0, 50)

	for p := range src {
		if _, ok := dst[p]; ok {
			continue
		}

		if _, ok := dst[cutTestdata(p)]; ok {
			continue
		}

		res = append(res, p)
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
		os.Remove(dst)
		return lazyerrors.Error(err)
	}

	return nil
}

// copyCorpus copies all new corpus files from srcRoot to dstRoot.
func copyCorpus(srcRoot, dstRoot string) {
	logger := zap.S()

	srcFiles, err := collectFiles(srcRoot, logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Found %d files in src.", len(srcFiles))

	dstFiles, err := collectFiles(dstRoot, logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Found %d existing files in dst.", len(dstFiles))

	files := diff(srcFiles, dstFiles)
	logger.Infof("Copying new %d files to dst.", len(files))

	for _, p := range files {
		src := filepath.Join(srcRoot, p)
		dst := cutTestdata(filepath.Join(dstRoot, p))
		logger.Debugf("%s -> %s", src, dst)

		if err := copyFile(src, dst); err != nil {
			logger.Fatal(err)
		}
	}
}

// The cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
var cli struct {
	Corpus struct {
		Src string `arg:"" help:"Source, one of: 'seed', 'generated', or collected corpus' directory."`
		Dst string `arg:"" help:"Destination, one of: 'seed', 'generated', or collected corpus' directory."`
	} `cmd:""`
	Debug bool `help:"Enable debug mode."`
}

func main() {
	ctx := kong.Parse(&cli)

	logging.Setup(zap.InfoLevel, "")

	if cli.Debug {
		logging.Setup(zap.DebugLevel, "")
	}

	logger := zap.S()

	seedCorpus, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	generatedCorpus, err := generatedCorpus()
	if err != nil {
		logger.Fatal(err)
	}

	var src, dst string

	switch ctx.Command() {
	case "corpus <src> <dst>":
		switch cli.Corpus.Src {
		case "seed":
			src = seedCorpus
		case "generated":
			src = generatedCorpus
		default:
			src, err = filepath.Abs(cli.Corpus.Src)
			if err != nil {
				logger.Fatal(err)
			}
		}

		switch cli.Corpus.Dst {
		case "seed":
			// Because we would need to add `/testdata/fuzz` back, and that's not very easy.
			logger.Fatal("Copying to seed corpus is not supported.")
		case "generated":
			dst = generatedCorpus
		default:
			dst, err = filepath.Abs(cli.Corpus.Dst)
			if err != nil {
				logger.Fatal(err)
			}
		}

	default:
		panic(ctx.Command())
	}

	logger.Infof("Copying from %s to %s.", src, dst)
	copyCorpus(src, dst)
}

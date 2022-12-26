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

package version

import (
	"embed"
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Each pattern in a //go:embed line must match at least one file or non-empty directory,
// but most files in gen/ are generated and are not present when FerretDB is used as library package.
// As a workaround, gen/mongodb.txt is always present.

//go:generate go run ./generate.go

//go:embed gen
var gen embed.FS

// Info provides details about the current build.
type Info struct {
	Version          string
	Commit           string
	Branch           string
	Dirty            bool
	Debug            bool // -tags=ferretdb_testcover or -race
	BuildEnvironment *types.Document
}

var (
	// MongoDBVersion is a fake MongoDB version for clients that check major.minor to adjust their behavior.
	MongoDBVersion string

	// MongoDBVersionArray is MongoDBVersion, but as an array.
	MongoDBVersionArray *types.Array

	info *Info
)

// unknown is a placeholder for unknown version, commit, and branch values.
const unknown = "unknown"

// Get returns current build's info.
func Get() *Info {
	return info
}

func init() {
	b := must.NotFail(gen.ReadFile("gen/mongodb.txt"))
	parts := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`).FindStringSubmatch(strings.TrimSpace(string(b)))
	if len(parts) != 4 {
		panic("invalid gen/mongodb.txt")
	}
	major := must.NotFail(strconv.Atoi(parts[1]))
	minor := must.NotFail(strconv.Atoi(parts[2]))
	patch := must.NotFail(strconv.Atoi(parts[3]))
	MongoDBVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	MongoDBVersionArray = must.NotFail(types.NewArray(int32(major), int32(minor), int32(patch), int32(0)))

	// those files are not present when FerretDB is used as library package
	version := unknown
	if b, _ := gen.ReadFile("gen/version.txt"); len(b) > 0 {
		version = strings.TrimSpace(string(b))
	}
	commit := unknown
	if b, _ := gen.ReadFile("gen/commit.txt"); len(b) > 0 {
		commit = strings.TrimSpace(string(b))
	}
	branch := unknown
	if b, _ := gen.ReadFile("gen/branch.txt"); len(b) > 0 {
		branch = strings.TrimSpace(string(b))
	}

	info = &Info{
		Version:          version,
		Commit:           commit,
		Branch:           branch,
		BuildEnvironment: must.NotFail(types.NewDocument()),
	}

	// do not expose extra information when FerretDB is used as library package
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if buildInfo.Main.Path != "github.com/FerretDB/FerretDB" {
		return
	}

	for _, s := range buildInfo.Settings {
		switch s.Key {
		case "vcs.revision":
			if s.Value != info.Commit {
				panic(fmt.Sprintf("commit.txt value %q != vcs.revision value %q\n"+
					"Please run `bin/task gen-version`", info.Commit, s.Value,
				))
			}

		case "vcs.modified":
			info.Dirty = must.NotFail(strconv.ParseBool(s.Value))

		case "-compiler":
			info.BuildEnvironment.Set("compiler", s.Value)

		case "-race":
			info.BuildEnvironment.Set("race", s.Value)

			if must.NotFail(strconv.ParseBool(s.Value)) {
				info.Debug = true
			}

		case "-tags":
			info.BuildEnvironment.Set("buildtags", s.Value)

			if slices.Contains(strings.Split(s.Value, ","), "ferretdb_testcover") {
				info.Debug = true
			}

		case "-trimpath":
			info.BuildEnvironment.Set("trimpath", s.Value)

		default:
			info.BuildEnvironment.Set(s.Key, s.Value)
		}
	}
}

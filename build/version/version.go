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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Each pattern in a //go:embed line must match at least one file or non-empty directory,
// but most files are generated and are not present when FerretDB is used as library package.
// As a workaround, mongodb.txt is always present.

//go:generate go run ./generate.go

//go:embed *.txt
var gen embed.FS

// Info provides details about the current build.
type Info struct {
	Version          string
	Commit           string
	Branch           string
	Dirty            bool
	DebugBuild       bool
	BuildEnvironment *types.Document

	// MongoDBVersion is fake MongoDB version for clients that check major.minor to adjust their behavior.
	MongoDBVersion string

	// MongoDBVersionArray is MongoDBVersion, but as an array.
	MongoDBVersionArray *types.Array
}

// info singleton instance set by init().
var info *Info

// unknown is a placeholder for unknown version, commit, and branch values.
const unknown = "unknown"

// Get returns current build's info.
func Get() *Info {
	return info
}

func init() {
	// this file is always present
	b := must.NotFail(gen.ReadFile("mongodb.txt"))
	parts := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`).FindStringSubmatch(strings.TrimSpace(string(b)))
	if len(parts) != 4 {
		panic("invalid mongodb.txt")
	}
	major := must.NotFail(strconv.Atoi(parts[1]))
	minor := must.NotFail(strconv.Atoi(parts[2]))
	patch := must.NotFail(strconv.Atoi(parts[3]))
	mongoDBVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	mongoDBVersionArray := must.NotFail(types.NewArray(int32(major), int32(minor), int32(patch), int32(0)))

	info = &Info{
		Version:             unknown,
		Commit:              unknown,
		Branch:              unknown,
		Dirty:               false,
		DebugBuild:          false,
		BuildEnvironment:    must.NotFail(types.NewDocument()),
		MongoDBVersion:      mongoDBVersion,
		MongoDBVersionArray: mongoDBVersionArray,
	}

	// do not expose extra information when FerretDB is used as library package
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if buildInfo.Main.Path != "github.com/FerretDB/FerretDB" {
		return
	}

	info.DebugBuild = debugBuild

	// this file must always be present, even in non-official builds
	if b, _ := gen.ReadFile("version.txt"); len(b) > 0 {
		info.Version = strings.TrimSpace(string(b))
	}
	if !strings.HasPrefix(info.Version, "v") {
		msg := "Invalid build/version/version.txt file content. Please run `bin/task gen-version`.\n"
		msg += "Alternatively, create this file manually with a content similar to\n"
		msg += "the output of `git describe --tags --dirty`: `v<major>.<minor>.<patch>`."
		panic(msg)
	}

	// those files may be absent in non-official builds
	if b, _ := gen.ReadFile("commit.txt"); len(b) > 0 {
		info.Commit = strings.TrimSpace(string(b))
	}
	if b, _ := gen.ReadFile("branch.txt"); len(b) > 0 {
		info.Branch = strings.TrimSpace(string(b))
	}

	for _, s := range buildInfo.Settings {
		switch s.Key {
		case "vcs.revision":
			if s.Value != info.Commit {
				// for non-official builds
				if info.Commit == unknown {
					info.Commit = s.Value
					continue
				}

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

		case "-tags":
			info.BuildEnvironment.Set("buildtags", s.Value)

		case "-trimpath":
			info.BuildEnvironment.Set("trimpath", s.Value)

		default:
			info.BuildEnvironment.Set(s.Key, s.Value)
		}
	}
}

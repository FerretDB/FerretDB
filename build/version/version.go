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

// Package version provides information about FerretDB version and build configuration.
//
// # Extra files
//
// The following generated text files may be present in this (`build/version`) directory during building:
//   - version.txt (required) contains information about the FerretDB version in a format
//     similar to `git describe` output: `v<major>.<minor>.<patch>`.
//   - commit.txt (optional) contains information about the source git commit.
//   - branch.txt (optional) contains information about the source git branch.
//   - package.txt (optional) contains package type (e.g. "deb", "rpm", "docker", etc).
//
// # Go build tags
//
// The following Go build tags (also known as build constraints) affect builds of FerretDB:
//
//	ferretdb_dev - enables development build (see below; implied by builds with race detector)
//
// # Development builds
//
// Development builds of FerretDB behave differently in a few aspects:
//   - some values that are normally randomized are fixed or less randomized;
//   - some internal errors cause crashes instead of being handled more gracefully;
//   - stack traces are collected more liberally;
//   - metrics are written to stderr on exit;
//   - the default logging level is set to debug.
//
// They are significantly slower.  to make debugging easier.
package version

import (
	"embed"
	"fmt"
	"regexp"
	"runtime"
	runtimedebug "runtime/debug"
	"strconv"
	"strings"

	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Each pattern in a //go:embed line must match at least one file or non-empty directory,
// but most files are generated and may be absent (for example, when embeddable package is being used).
// As a workaround, mongodb.txt is always present and stored in the repository / Go module.

//go:generate go run ./generate.go

//go:embed *.txt
var gen embed.FS

// Info provides details about the current build.
//
//nolint:vet // for readability
type Info struct {
	Version          string
	Commit           string
	Branch           string
	Dirty            bool
	Package          string
	DevBuild         bool
	BuildEnvironment map[string]string

	// MongoDBVersion is fake MongoDB version for clients that check major.minor to adjust their behavior.
	MongoDBVersion string

	// MongoDBVersionArray is MongoDBVersion, but as an array.
	MongoDBVersionArray [4]int32
}

// info singleton instance set by init().
var info *Info

// unknown is a placeholder for unknown version, commit, and branch values.
const unknown = "unknown"

// FerretDB module path from go.mod.
const ferretdbModule = "github.com/FerretDB/FerretDB/v2"

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
//
//nolint:lll // for readibility
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Get returns current build's info.
//
// It returns a shared instance without any synchronization.
// If caller needs to modify the instance, it should make sure there is no concurrent accesses.
func Get() *Info {
	return info
}

func init() {
	mongodbTxt := strings.TrimSpace(string(must.NotFail(gen.ReadFile("mongodb.txt"))))
	match := semVerTag.FindStringSubmatch(mongodbTxt)
	if match == nil || len(match) != semVerTag.NumSubexp()+1 {
		panic("invalid mongodb.txt")
	}

	major := must.NotFail(strconv.ParseInt(match[semVerTag.SubexpIndex("major")], 10, 32))
	minor := must.NotFail(strconv.ParseInt(match[semVerTag.SubexpIndex("minor")], 10, 32))
	patch := must.NotFail(strconv.ParseInt(match[semVerTag.SubexpIndex("patch")], 10, 32))
	mongoDBVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	mongoDBVersionArray := [...]int32{int32(major), int32(minor), int32(patch), int32(0)}

	info = &Info{
		Version:  unknown,
		Commit:   unknown,
		Branch:   unknown,
		Dirty:    false,
		Package:  unknown,
		DevBuild: devbuild.Enabled,
		BuildEnvironment: map[string]string{
			"go.runtime": runtime.Version(),
		},
		MongoDBVersion:      mongoDBVersion,
		MongoDBVersionArray: mongoDBVersionArray,
	}

	// in theory, someone could use embeddable package in a Go program that is built without modules
	buildInfo, ok := runtimedebug.ReadBuildInfo()
	if !ok {
		return
	}

	info.BuildEnvironment["go.version"] = buildInfo.GoVersion

	if buildInfo.Main.Path != ferretdbModule {
		// FIXME
		for _, dep := range buildInfo.Deps {
			if dep.Path == ferretdbModule {
				m := dep
				if dep.Replace != nil {
					m = dep.Replace
				}

				info.Version = m.Version
			}
		}

		return
	}

	// panic("lala")

	// those files may be present only when we build the FerretDB binary
	for f, sp := range map[string]*string{
		"version.txt": &info.Version,
		"commit.txt":  &info.Commit,
		"branch.txt":  &info.Branch,
		"package.txt": &info.Package,
	} {
		b, _ := gen.ReadFile(f)
		if s := strings.TrimSpace(string(b)); s != "" {
			*sp = s
		}
	}

	// buildInfoVersion = buildInfo.Main.Version

	// if !strings.HasPrefix(info.Version, "v") {
	// 	msg := "Invalid build/version/version.txt file content. Please run `bin/task gen-version`.\n"
	// 	msg += "Alternatively, create this file manually with a content similar to\n"
	// 	msg += "the output of `git describe`: `v<major>.<minor>.<patch>`.\n"
	// 	msg += "See https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version"
	// 	panic(msg)
	// }

	for _, s := range buildInfo.Settings {
		if v := s.Value; v != "" {
			info.BuildEnvironment[s.Key] = v
		}

		switch s.Key {
		case "vcs.revision":
			if s.Value != info.Commit {
				if info.Commit == unknown {
					info.Commit = s.Value
				}
			}

		case "vcs.modified":
			info.Dirty = must.NotFail(strconv.ParseBool(s.Value))
		}
	}
}

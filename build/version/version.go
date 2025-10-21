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
//   - they are significantly slower;
//   - some values that are normally randomized are fixed or less randomized to make debugging easier;
//   - some internal errors cause crashes instead of being handled more gracefully;
//   - stack traces are collected more liberally;
//   - metrics are written to stderr on exit;
//   - the default logging level is set to debug.
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
//nolint:lll // for readability
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Get returns current build's info.
//
// It returns a shared instance without any synchronization.
// If caller needs to modify the instance, it should make sure there is no concurrent accesses.
func Get() *Info {
	return info
}

// initFromFiles initializes info from txt files (that might be absent).
// All info fields are set to non-empty values, but some of them may be unknown.
func initFromFiles() {
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

	// Those files are present in two cases:
	// 1. When we are working in the FerretDB repo (building binaries, running integration tests, etc),
	//    those files are generated by `task gen-version`.
	// 2. When someone builds the FerretDB binary as described in the package description.
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
}

// readBuildInfo returns FerretDB version and commit from the build info.
// It also updates info.BuildEnvironment and info.Dirty when it is sure we are building FerretDB,
// and not something that uses FerretDB.
func readBuildInfo() (version, commit string) {
	// in theory, someone could use embeddable package in a Go program that is built without modules
	buildInfo, ok := runtimedebug.ReadBuildInfo()
	if !ok {
		return
	}

	info.BuildEnvironment["go.version"] = buildInfo.GoVersion

	// What buildInfo contains depends on what is being built and how.
	// There are at least five cases:
	//
	// 1. Builds and unit tests in the FerretDB repo (task build-host, task test-unit, etc).
	//      Version is not set even for tags due to https://github.com/golang/go/issues/72877.
	//    Path: "github.com/FerretDB/FerretDB/v2/cmd/ferretdb"
	//      (or "github.com/FerretDB/FerretDB/v2/build/version.test", etc)
	//    Main.Path: "github.com/FerretDB/FerretDB/v2"
	//    Main.Version: "(devel)"
	//
	// 2. Builds with known module version (go install github.com/FerretDB/FerretDB/v2/cmd/ferretdb@v2.0.0).
	//    Path: "github.com/FerretDB/FerretDB/v2/cmd/ferretdb"
	//    Main.Path: "github.com/FerretDB/FerretDB/v2"
	//    Main.Version: "v2.0.0"
	//
	// 3. Ad-hoc builds (go run main.go readyz.go --dev-version) due to https://github.com/golang/go/issues/51279.
	//    Path: "command-line-arguments"
	//    Main.Path: ""
	//    Main.Version: ""
	//    Deps.Path: "github.com/FerretDB/FerretDB/v2"
	//    Deps.Version: "(devel)"
	//
	// 4. Integration tests (both in the test and in the command handler).
	//    Path: "github.com/FerretDB/FerretDB/v2/integration.test"
	//    Main.Path: "github.com/FerretDB/FerretDB/v2/integration"
	//    Main.Version: "(devel)"
	//    Deps: null
	//
	// 5. Embeddable package.
	//    Path: ???
	//    Main.Path: ???
	//    Deps.Path: "github.com/FerretDB/FerretDB/v2"
	//    Deps.Version: "v2.0.0"
	//    With replace directive in go.mod:
	//      Deps.Replace.Path: "../FerretDB"
	//      Deps.Replace.Version: "(devel)"
	//
	// In short, FerretDB version could be derived from the module version only when module version could be known,
	// which is not often the case. And there are bugs in the Go toolchain that prevent version from being set
	// even when it is known.

	// cases 1 and 2
	if buildInfo.Main.Path == ferretdbModule {
		version = buildInfo.Main.Version

		// in the case 1 even for tags due to https://github.com/golang/go/issues/72877
		if version == "(devel)" {
			version = ""
		}

		for _, s := range buildInfo.Settings {
			if v := s.Value; v != "" {
				info.BuildEnvironment[s.Key] = v
			}

			// both are present only in the case 1
			switch s.Key {
			case "vcs.revision":
				commit = s.Value
			case "vcs.modified":
				info.Dirty = must.NotFail(strconv.ParseBool(s.Value))
			}
		}

		return
	}

	// cases 3, 4, and 5

	for _, dep := range buildInfo.Deps {
		if dep.Path != ferretdbModule {
			continue
		}

		version = dep.Version
		if dep.Replace != nil {
			version = dep.Replace.Version
		}

		// cases 3 and 4
		if version == "(devel)" {
			version = ""
		}

		// vcs.revision and other settings refer to the repository that uses FerretDB, not FerretDB itself.
		// We don't want git commit hashes or build configurations of other repos.

		break
	}

	return
}

func init() {
	initFromFiles()

	version, commit := readBuildInfo()

	if info.Version == unknown && version != "" {
		info.Version = version
	}

	if info.Commit == unknown && commit != "" {
		info.Commit = commit
	}

	if info.Version != unknown {
		if match := semVerTag.FindStringSubmatch(info.Version); match == nil || len(match) != semVerTag.NumSubexp()+1 {
			msg := fmt.Sprintf("info.Version: %q, version: %q\n", info.Version, version)
			msg += "Invalid build/version/version.txt file content. Please run `bin/task gen-version`.\n"
			msg += "Alternatively, create this file manually with a content similar to\n"
			msg += "the output of `git describe`: `v<major>.<minor>.<patch>`.\n"
			msg += "See https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version"
			panic(msg)
		}
	}

	// This does not work for git submodules.
	// Investigate and create an issue.
	//
	// if info.Commit != unknown {
	// 	if info.Commit != commit && commit != "" {
	// 		msg := fmt.Sprintf("info.Commit: %q, commit: %q\n", info.Commit, commit)
	// 		msg += "Invalid build/version/commit.txt file content. Please run `bin/task gen-version`."
	// 		panic(msg)
	// 	}
	// }
}

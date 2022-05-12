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
	_ "embed"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate go run ./generate.go

var (
	//go:embed version.txt
	version string

	//go:embed commit.txt
	commit string

	//go:embed branch.txt
	branch string
)

type Info struct {
	Version          string
	Commit           string
	Branch           string
	Dirty            bool
	Debug            bool // testcover or -race
	BuildEnvironment *types.Document
}

var info *Info

func Get() *Info {
	return info
}

func init() {
	info = &Info{
		Version: strings.TrimSpace(version),
		Commit:  strings.TrimSpace(commit),
		Branch:  strings.TrimSpace(branch),
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	info.BuildEnvironment = must.NotFail(types.NewDocument())
	for _, s := range buildInfo.Settings {
		info.BuildEnvironment.Set(s.Key, s.Value)

		switch s.Key {
		case "vcs.revision":
			if s.Value != info.Commit {
				panic(fmt.Sprintf("commit.txt value %q != vcs.revision value %q\n"+
					"Please run `bin/task gen-version`", info.Commit, s.Value,
				))
			}
		case "vcs.modified":
			info.Dirty, _ = strconv.ParseBool(s.Value)
		case "-race":
			if raceEnabled, _ := strconv.ParseBool(s.Value); raceEnabled {
				info.Debug = true
			}
		case "-tags":
			if slices.Contains(strings.Split(s.Value, ","), "testcover") {
				info.Debug = true
			}
		}
	}
}

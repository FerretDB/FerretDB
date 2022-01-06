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
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
)

//go:generate ./version.sh

//go:embed version.txt
var version string

type Info struct {
	Version          string
	Commit           string
	Dirty            bool
	IsDebugBuild     bool
	BuildEnvironment types.Document
}

var info *Info

func Get() *Info {
	return info
}

func init() {
	info = &Info{
		Version: strings.TrimSpace(version),
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	info.BuildEnvironment = types.MustMakeDocument()
	for _, s := range buildInfo.Settings {
		info.BuildEnvironment.Set(s.Key, s.Value)
		switch s.Key {
		case "vcs.revision":
			info.Commit = s.Value
		case "vcs.modified":
			info.Dirty, _ = strconv.ParseBool(s.Value)
		case "-race":
			if raceEnabled, _ := strconv.ParseBool(s.Value); raceEnabled {
				info.IsDebugBuild = true
			}
		case "-tags":
			// TODO: replace for slices.Contains()
			for _, tag := range strings.Split(s.Value, ",") {
				if tag == "testcover" {
					info.IsDebugBuild = true
				}
			}
		}
	}
}

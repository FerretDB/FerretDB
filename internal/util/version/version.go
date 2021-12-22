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
)

//go:generate ./version.sh

//go:embed version.txt
var version string

type Info struct {
	Version      string
	Commit       string
	Dirty        bool
	Architecture int32 // either equals to 32 or 64
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

	for _, s := range buildInfo.Settings {
		switch s.Key {
		case "vcs.revision":
			info.Commit = s.Value
		case "vcs.modified":
			info.Dirty, _ = strconv.ParseBool(s.Value)
		case "architecture":
			temp, _ := strconv.ParseInt(s.Value, 10, 32)
			if temp == 32 {
				info.Architecture = x86
			} else if temp == 64 {
				info.Architecture = x64
			}

		}
	}
}

const (
	x86 int32 = 32
	x64 int32 = 64
)

// Copyright 2021 Baltoro OÜ.
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
	"runtime/debug"
	"strconv"
)

type Info struct {
	Version string
	Commit  string
	Dirty   bool
}

var info *Info

func Get() *Info {
	return info
}

func init() {
	debugInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	info = &Info{
		Version: "0.0.1",
	}

	for _, s := range debugInfo.Settings {
		switch s.Key {
		case "gitrevision":
			info.Commit = s.Value
		case "gituncommitted":
			info.Dirty, _ = strconv.ParseBool(s.Value)
		}
	}
}

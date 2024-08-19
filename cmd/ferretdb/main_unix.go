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

//go:build unix

package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// stateFileProblem adds details to the state file access error.
func stateFileProblem(f string, err error) string {
	res := fmt.Sprintf("Failed to create state provider: %s.\n", err)

	if u, _ := user.Current(); u != nil {
		var group string
		if g, _ := user.LookupGroupId(u.Gid); g != nil {
			group = g.Name
		}

		res += fmt.Sprintf("FerretDB is running as %s:%s (%s:%s). ", u.Username, group, u.Uid, u.Gid)
	}

	if fi, _ := os.Stat(f); fi != nil {
		var username, group string
		var uid, gid uint64

		if s, _ := fi.Sys().(*syscall.Stat_t); s != nil {
			uid = uint64(s.Uid)
			if u, _ := user.LookupId(strconv.FormatUint(uid, 10)); u != nil {
				username = u.Username
			}

			gid = uint64(s.Gid)
			if g, _ := user.LookupGroupId(strconv.FormatUint(gid, 10)); g != nil {
				group = g.Name
			}
		}

		res += fmt.Sprintf("%s permissions are %s.", f, fi.Mode().String())
		if username != "" {
			res += fmt.Sprintf(" Owned by %s:%s (%d:%d).", username, group, uid, gid)
		}
	}

	return res
}

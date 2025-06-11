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

package state

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// newProviderDirErr adds details to the state file access error.
func newProviderDirErr(f string, err error) error {
	var extra []string

	if u, _ := user.Current(); u != nil {
		var group string
		if g, _ := user.LookupGroupId(u.Gid); g != nil {
			group = g.Name
		}

		extra = append(extra, fmt.Sprintf("running as %s:%s/%s:%s", u.Username, group, u.Uid, u.Gid))
	}

	if fi, _ := os.Stat(f); fi != nil {
		extra = append(extra, fmt.Sprintf("%s permissions are %s", f, fi.Mode().String()))

		var username, group string
		if s, _ := fi.Sys().(*unix.Stat_t); s != nil {
			if u, _ := user.LookupId(strconv.Itoa(int(s.Uid))); u != nil {
				username = u.Username
			}

			if g, _ := user.LookupGroupId(strconv.Itoa(int(s.Gid))); g != nil {
				group = g.Name
			}

			extra = append(extra, fmt.Sprintf("owned by %s:%s/%d:%d", username, group, s.Uid, s.Gid))
		}
	}

	if extra == nil {
		return err
	}

	return fmt.Errorf("%s (%s)", err, strings.Join(extra, ", "))
}

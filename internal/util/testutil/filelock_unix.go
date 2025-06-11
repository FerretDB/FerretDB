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

package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// flock wraps flock syscall with a retry mechanism.
func flock(tb testing.TB, f *os.File, op string) {
	var how int

	switch op {
	case "shared":
		how = unix.LOCK_SH
	case "exclusive":
		how = unix.LOCK_EX
	case "unlock":
		how = unix.LOCK_UN
	default:
		panic(fmt.Errorf("unknown flock operation: %s", op))
	}

	for {
		err := unix.Flock(int(f.Fd()), how)
		if err == nil {
			return
		}

		tb.Logf("%s flock: %s %s %s (%#v)", tb.Name(), op, f.Name(), err, err)

		if err == unix.EINTR {
			continue
		}

		require.NoError(tb, err, "%s %s", op, f.Name())
	}
}

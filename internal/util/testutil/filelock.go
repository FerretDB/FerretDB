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

package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// fileLock represents a global lock.
// It acts like a [sync.Mutex], but works across variable instances and cooperative processes.
// It uses `flock` with the same file to synchronize access.
type fileLock struct {
	tb    testing.TB
	f     *os.File
	token *resource.Token
}

// newFileLock creates a new instance of shared file lock.
// It must be unlocked by the caller.
func newFileLock(tb testing.TB) *fileLock {
	f, err := os.OpenFile(filepath.Join(TmpDir, "testutil-filelock.txt"), os.O_RDWR|os.O_CREATE, 0o666)
	require.NoError(tb, err)

	flock(tb, f, "shared")

	fl := &fileLock{
		tb:    tb,
		f:     f,
		token: resource.NewToken(),
	}

	resource.Track(fl, fl.token)

	return fl
}

// Lock upgrades shared lock to exclusive, blocking until the lock is acquired.
func (fl *fileLock) Lock() {
	require.NotNil(fl.tb, fl.f)

	flock(fl.tb, fl.f, "exclusive")

	// This creates a barrier for the race detector (see https://github.com/golang/go/issues/50139)
	// without adding an explicit sync.Mutex to fileLock (that would make testing harder),
	// and also useful for debugging.
	err := fl.f.Truncate(0)
	require.NoError(fl.tb, err)
	_, err = fl.f.WriteString(fl.tb.Name() + "\n")
	require.NoError(fl.tb, err)
}

// Unlock unlocks file.
// This instance can't be used after calling this method.
func (fl *fileLock) Unlock() {
	require.NotNil(fl.tb, fl.f)

	flock(fl.tb, fl.f, "unlock")

	err := fl.f.Close()
	require.NoError(fl.tb, err)

	fl.f = nil

	resource.Untrack(fl, fl.token)
}

// check interfaces
var (
	// We need to implement that interface mainly for the copylock vet check.
	_ sync.Locker = (*fileLock)(nil)
)

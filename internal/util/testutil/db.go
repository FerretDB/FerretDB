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
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	databaseNamesM sync.Mutex
	databaseNames  = map[string][]byte{}

	collectionNamesM sync.Mutex
	collectionNames  = map[string][]byte{}
)

// stack returns the stack trace starting from the caller of caller.
func stack() []byte {
	pc := make([]uintptr, 100)

	callers := runtime.Callers(2, pc)
	if callers == 0 {
		panic("runtime.Callers failed")
	}

	frames := runtime.CallersFrames(pc[:callers])
	var buf bytes.Buffer

	for {
		frame, more := frames.Next()
		if frame.File != "" {
			fmt.Fprintf(&buf, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		}

		if !more {
			return buf.Bytes()
		}
	}
}

// DatabaseName returns a stable FerretDB database name for that test.
//
// It should be called only once per test.
func DatabaseName(tb testing.TB) string {
	tb.Helper()

	// database names are always lowercase
	name := strings.ToLower(tb.Name())

	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 64)

	databaseNamesM.Lock()
	defer databaseNamesM.Unlock()

	// it maybe exactly the same if `go test -count=X` is used
	current := stack()
	if another, ok := databaseNames[name]; ok && !bytes.Equal(current, another) {
		tb.Logf("Database name %q already used by another test:\n%s", name, another)
		panic("duplicate database name")
	}
	databaseNames[name] = current

	return name
}

// CollectionName returns a stable FerretDB collection name for that test.
//
// It should be called only once per test.
func CollectionName(tb testing.TB) string {
	tb.Helper()

	// do not use strings.ToLower because collection names can contain uppercase letters
	name := tb.Name()

	// name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 255)

	collectionNamesM.Lock()
	defer collectionNamesM.Unlock()

	// it may be exactly the same if `go test -count=X` is used
	current := stack()
	if another, ok := collectionNames[name]; ok && !bytes.Equal(current, another) {
		tb.Logf("Collection name %q already used by another test:\n%s", name, another)
		panic("duplicate collection name")
	}
	collectionNames[name] = current

	return name
}

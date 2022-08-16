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
	"runtime/debug"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	databaseNamesM sync.Mutex
	databaseNames  = make(map[string][]byte)

	collectionNamesM sync.Mutex
	collectionNames  = make(map[string][]byte)
)

func stack() []byte {
	s := bytes.Split(debug.Stack(), []byte("\n"))
	return bytes.Join(s[7:], []byte("\n"))
}

// DatabaseName returns a stable database name for that test.
func DatabaseName(tb testing.TB) string {
	tb.Helper()

	// database names are always lowercase
	name := strings.ToLower(tb.Name())

	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 64)

	databaseNamesM.Lock()
	defer databaseNamesM.Unlock()

	current := stack()
	if another, ok := databaseNames[name]; ok {
		tb.Logf("Database name %q already used by another test:\n%s", name, another)
		panic("duplicate database name")
	}
	databaseNames[name] = current

	return name
}

// CollectionName returns a stable collection name for that test.
func CollectionName(tb testing.TB) string {
	tb.Helper()

	// do not use strings.ToLower because collection names can contain uppercase letters
	name := tb.Name()

	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 255)

	collectionNamesM.Lock()
	defer collectionNamesM.Unlock()

	current := stack()
	if another, ok := collectionNames[name]; ok {
		tb.Logf("Collection name %q already used by another test:\n%s", name, another)
		panic("duplicate collection name")
	}
	collectionNames[name] = current

	return name
}

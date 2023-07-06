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
	"fmt"
	"strings"
	"sync"

	"github.com/stretchr/testify/require"
)

var (
	databaseNamesM sync.Mutex
	databaseNames  = map[string]string{}

	collectionNamesM sync.Mutex
	collectionNames  = map[string]string{}
)

// DatabaseName returns a stable FerretDB database name for that test.
//
// It should be called only once per test.
func DatabaseName(tb TB) string {
	tb.Helper()

	// database names may contain lowercase and uppercase characters
	name := tb.Name()

	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 64)

	databaseNamesM.Lock()
	defer databaseNamesM.Unlock()

	// it may be the same test if `go test -count=X` is used
	if t, ok := databaseNames[name]; ok && t != tb.Name() {
		panic(fmt.Sprintf("Database name %q already used by another test %q.", name, tb.Name()))
	}

	databaseNames[name] = tb.Name()

	return name
}

// CollectionName returns a stable FerretDB collection name for that test.
//
// It should be called only once per test.
func CollectionName(tb TB) string {
	tb.Helper()

	// do not use strings.ToLower because collection names can contain uppercase letters
	name := tb.Name()

	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "$", "_")

	require.Less(tb, len(name), 255)

	collectionNamesM.Lock()
	defer collectionNamesM.Unlock()

	// it may be the same test if `go test -count=X` is used
	if t, ok := collectionNames[name]; ok && t != tb.Name() {
		panic(fmt.Sprintf("Collection name %q already used by another test %q.", name, tb.Name()))
	}

	collectionNames[name] = tb.Name()

	return name
}

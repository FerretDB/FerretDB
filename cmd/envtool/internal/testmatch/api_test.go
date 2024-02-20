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

// Tests based on match_test.go
// This file is a modification of
// https://go.googlesource.com/go/+/d31efbc95e6803742aaca39e3a825936791e6b5a/src/testing/match_test.go
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file at
// https://go.googlesource.com/go/+/d31efbc95e6803742aaca39e3a825936791e6b5a/LICENSE

package testmatch

import (
	"testing"
)

func TestMatcherAPI(t *testing.T) {
	testCases := []struct {
		pattern string
		skip    string
		name    string
		ok      bool
	}{
		// Behavior without subtests.
		{"", "", "TestFoo", true},
		{"TestFoo", "", "TestFoo", true},
		{"TestFoo/", "", "TestFoo", true},
		{"TestFoo/bar/baz", "", "TestFoo", true},
		{"TestFoo", "", "TestBar", false},
		{"TestFoo/", "", "TestBar", false},
		{"TestFoo/bar/baz", "", "TestBar/bar/baz", false},
		{"", "TestBar", "TestFoo", true},
		{"", "TestBar", "TestBar", false},

		// Skipping a non-existent test doesn't change anything.
		{"", "TestFoo/skipped", "TestFoo", true},
		{"TestFoo", "TestFoo/skipped", "TestFoo", true},
		{"TestFoo/", "TestFoo/skipped", "TestFoo", true},
		{"TestFoo/bar/baz", "TestFoo/skipped", "TestFoo", true},
		{"TestFoo", "TestFoo/skipped", "TestBar", false},
		{"TestFoo/", "TestFoo/skipped", "TestBar", false},
		{"TestFoo/bar/baz", "TestFoo/skipped", "TestBar/bar/baz", false},
	}

	for _, tc := range testCases {
		m := New(tc.pattern, tc.skip)

		if ok := m.Match(tc.name); ok != tc.ok {
			t.Errorf("for pattern %q, Match(%q) = %v; want ok %v",
				tc.pattern, tc.name, ok, tc.ok)
		}
	}
}

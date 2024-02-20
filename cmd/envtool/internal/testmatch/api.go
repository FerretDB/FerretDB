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

package testmatch

import "regexp"

type Matcher struct {
	m *matcher
}

// New matcher.
func New(run, skip string) *Matcher {
	return &Matcher{
		m: newMatcher(regexp.MatchString, run, "-test.run", skip),
	}
}

// Match top-level test function.
func (m *Matcher) Match(testFunction string) bool {
	_, ok, _ := m.m.fullName(&common{}, testFunction)
	return ok
}

// common is used internally by the matcher.
type common struct {
	name  string // name of the test
	level int    // level of the test
}

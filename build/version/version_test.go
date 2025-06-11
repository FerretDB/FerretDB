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

package version

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCase1(t *testing.T) {
	assert.Regexp(t, semVerTag, info.Version)
	assert.Regexp(t, `^[0-9a-f]{40}$`, info.Commit)
	assert.NotEmpty(t, info.Branch)
	assert.NotEqual(t, unknown, info.Branch)
	assert.NotEmpty(t, info.Package)
	// package is unknown on CI for short tests where package.txt is not created

	assert.Equal(t, "7.0.77", info.MongoDBVersion)
	assert.Equal(t, [...]int32{int32(7), int32(0), int32(77), int32(0)}, info.MongoDBVersionArray)

	assert.Equal(t, runtime.Version(), info.BuildEnvironment["go.version"])
	assert.Equal(t, runtime.Version(), info.BuildEnvironment["go.runtime"])
	assert.Empty(t, info.BuildEnvironment["vcs.revision"]) // not set for unit tests
}

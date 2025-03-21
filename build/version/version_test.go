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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	runtimedebug "runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormal(t *testing.T) {
	b, err := json.MarshalIndent(info, "", "  ")
	require.NoError(t, err)
	t.Logf("info: %s", b)

	buildInfo, ok := runtimedebug.ReadBuildInfo()
	require.True(t, ok)

	b, err = json.MarshalIndent(buildInfo, "", "  ")
	require.NoError(t, err)
	t.Logf("buildInfo: %s", b)

	assert.Regexp(t, semVerTag, info.Version)
	assert.Regexp(t, `^[0-9a-f]{40}$`, info.Commit)
	assert.NotEmpty(t, info.Branch)
	assert.NotEqual(t, unknown, info.Branch)
	assert.NotEmpty(t, info.Package)
	assert.NotEqual(t, unknown, info.Package)

	assert.Equal(t, "7.0.77", info.MongoDBVersion)
	assert.Equal(t, [...]int32{int32(7), int32(0), int32(77), int32(0)}, info.MongoDBVersionArray)

	assert.Equal(t, runtime.Version(), info.BuildEnvironment["go.version"])

	assert.Equal(t, "7.0.77", info.MongoDBVersion)
	assert.Equal(t, [4]int32{7, 0, 77, 0}, info.MongoDBVersionArray)
}

func TestCompileTest(t *testing.T) {
	f := filepath.Join(t.TempDir(), "version-test.exe")

	cmd := exec.Command("go", "test", "-v", "-c", "-o", f)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err)

	b, err := exec.Command(f, "-test.run=TestNormal", "-test.v").CombinedOutput()
	t.Logf("%s", b)
	require.NoError(t, err)
	assert.NotContains(t, string(b), "no tests")
	assert.Contains(t, string(b), "PASS\n")
}

func TestGoRun(t *testing.T) {
	b, err := exec.Command("go", "run", "main.go").CombinedOutput()
	t.Logf("%s", b)
	require.NoError(t, err)

	var info Info
	err = json.Unmarshal(b, &info)
	require.NoError(t, err)

	assert.Regexp(t, semVerTag, info.Version)
	assert.Regexp(t, `^[0-9a-f]{40}$`, info.Commit)
	assert.NotEmpty(t, info.Branch)
	assert.NotEqual(t, unknown, info.Branch)
	assert.NotEmpty(t, info.Package)
	assert.NotEqual(t, unknown, info.Package)

	assert.Equal(t, "7.0.77", info.MongoDBVersion)
	assert.Equal(t, [...]int32{int32(7), int32(0), int32(77), int32(0)}, info.MongoDBVersionArray)

	assert.Equal(t, runtime.Version(), info.BuildEnvironment["go.version"])

	assert.Equal(t, "7.0.77", info.MongoDBVersion)
	assert.Equal(t, [4]int32{7, 0, 77, 0}, info.MongoDBVersionArray)
}

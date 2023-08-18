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

package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestsShard(t *testing.T) {
	t.Parallel()

	tests, err := getAllTestNames(filepath.Join("..", "..", "integration"))
	require.NoError(t, err)
	assert.Contains(t, tests, "TestQueryCompatLimit")

	// check that integration tests from subdirectories are included
	assert.Contains(t, tests, "TestCreateStress")
	assert.Contains(t, tests, "TestCommandsDiagnosticExplain")

	t.Run("ShardTestsInvalidIndex", func(t *testing.T) {
		t.Parallel()

		_, err := shardTests(0, 3, tests)
		assert.EqualError(t, err, "index must be greater than 0")

		_, err = shardTests(3, 3, tests)
		assert.NoError(t, err)

		_, err = shardTests(4, 3, tests)
		assert.EqualError(t, err, "cannot shard when index is greater to total (4 > 3)")
	})

	t.Run("ShardTestsInvalidTotal", func(t *testing.T) {
		t.Parallel()

		_, err := shardTests(3, 1000, tests[:42])
		assert.EqualError(t, err, "cannot shard when total is greater than amount of tests (1000 > 42)")
	})

	t.Run("ShardTestsValid", func(t *testing.T) {
		t.Parallel()

		res, err := shardTests(1, 3, tests)
		require.NoError(t, err)
		assert.Equal(t, tests[0], res[0])
		assert.NotEqual(t, tests[1], res[1])
		assert.NotEqual(t, tests[2], res[1])
		assert.Equal(t, tests[3], res[1])

		res, err = shardTests(3, 3, tests)
		require.NoError(t, err)
		assert.NotEqual(t, tests[0], res[0])
		assert.NotEqual(t, tests[1], res[0])
		assert.Equal(t, tests[2], res[0])
	})
}

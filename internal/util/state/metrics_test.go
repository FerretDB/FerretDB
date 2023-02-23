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

package state

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/build/version"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	p, err := NewProvider("")
	require.NoError(t, err)

	err = p.Update(func(s *State) { s.Telemetry = pointer.ToBool(true) })
	require.NoError(t, err)

	info := version.Get()

	t.Run("WithUUID", func(t *testing.T) {
		t.Parallel()

		uuid := p.Get().UUID

		mc := p.MetricsCollector(true)
		problems, err := testutil.CollectAndLint(mc)
		require.NoError(t, err)
		require.Empty(t, problems)

		//nolint:lll // it is more readable this way
		expected := fmt.Sprintf(
			`
				# HELP ferretdb_up FerretDB instance state.
				# TYPE ferretdb_up gauge
				ferretdb_up{branch=%q,commit=%q,debug="%t",dirty="%t",package="unknown",telemetry="enabled",update_available="false",uuid=%q,version=%q} 1
			`,
			info.Branch, info.Commit, info.DebugBuild, info.Dirty, uuid, info.Version,
		)
		assert.NoError(t, testutil.CollectAndCompare(mc, strings.NewReader(expected)))
	})

	t.Run("WithoutUUID", func(t *testing.T) {
		t.Parallel()

		mc := p.MetricsCollector(false)
		problems, err := testutil.CollectAndLint(mc)
		require.NoError(t, err)
		require.Empty(t, problems)

		//nolint:lll // it is more readable this way
		expected := fmt.Sprintf(
			`
				# HELP ferretdb_up FerretDB instance state.
				# TYPE ferretdb_up gauge
				ferretdb_up{branch=%q,commit=%q,debug="%t",dirty="%t",package="unknown",telemetry="enabled",update_available="false",version=%q} 1
			`,
			info.Branch, info.Commit, info.DebugBuild, info.Dirty, info.Version,
		)
		assert.NoError(t, testutil.CollectAndCompare(mc, strings.NewReader(expected)))
	})
}

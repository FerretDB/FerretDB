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

// Package setup provides integration tests setup helpers.
package setup

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNextTigrisURL(t *testing.T) {
	t.Parallel()

	t.Run("RoundRobin", func(t *testing.T) {
		urls := "127.0.0.1:8081,127.0.0.1:8082,127.0.0.1:8083,127.0.0.1:8085,127.0.0.1:8086"
		tigrisURLsF = &urls

		s := startup()
		nextURL := s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8081", nextURL)

		nextURL = s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8082", nextURL)

		nextURL = s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8083", nextURL)

		nextURL = s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8085", nextURL)

		nextURL = s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8086", nextURL)

		nextURL = s.getNextTigrisURL()
		require.NotEqual(t, "127.0.0.1:8081", nextURL)
	})
}

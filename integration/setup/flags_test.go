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

package setup

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	t.Run("ApplyOpts", func(t *testing.T) {
		testcases := map[string]struct {
			input       map[string]any
			expected    *flags
			initialFlag flags
		}{
			"TargetPort": {
				input: map[string]any{
					"target-port": 2222,
				},
				expected: &flags{targetPort: 2222},
			},
			"TargetTLS": {
				input: map[string]any{
					"target-tls": true,
				},
				expected: &flags{targetTLS: true},
			},
			"TargetUnixSocket": {
				input: map[string]any{
					"target-unix-socket": true,
				},
				expected: &flags{targetUnixSocket: true},
			},
			"ProxyAddr": {
				input: map[string]any{
					"proxy-addr": "1.2.3.4",
				},
				expected: &flags{proxyAddr: "1.2.3.4"},
			},
			"CompatPort": {
				input: map[string]any{
					"compat-port": 1111,
				},
				expected: &flags{compatPort: 1111},
			},
			"CompatTLS": {
				input: map[string]any{
					"compat-tls": true,
				},
				expected: &flags{compatTLS: true},
			},
			"PostgreSQLURL": {
				input: map[string]any{
					"postgresql-url": "pg-url",
				},
				expected: &flags{postgreSQLURL: "pg-url"},
			},
			"All": {
				input: map[string]any{
					"target-port":        1,
					"target-tls":         true,
					"handler":            "tigris",
					"target-unix-socket": true,
					"proxy-addr":         "1.2.3.4",
					"compat-port":        1111,
					"compat-tls":         true,
					"postgresql-url":     "pg url",
					"tigris-url":         "tigris url",
				},
				expected: &flags{
					targetPort:       1,
					targetTLS:        true,
					targetUnixSocket: true,
					proxyAddr:        "1.2.3.4",
					compatPort:       1111,
					compatTLS:        true,
					postgreSQLURL:    "pg url",
				},
			},
			"NoUpdateOnEmptyInput": {
				initialFlag: flags{
					targetPort:       1,
					targetTLS:        true,
					targetUnixSocket: true,
					proxyAddr:        "1.2.3.4",
					compatPort:       1111,
					compatTLS:        true,
					postgreSQLURL:    "pg url",
				},
				input: make(map[string]any, 0),
				expected: &flags{
					targetPort:       1,
					targetTLS:        true,
					targetUnixSocket: true,
					proxyAddr:        "1.2.3.4",
					compatPort:       1111,
					compatTLS:        true,
					postgreSQLURL:    "pg url",
				},
			},
		}

		for name, tc := range testcases {
			input := tc.input
			expected := tc.expected
			f := tc.initialFlag
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				updated := f.ApplyOpts(t, input)

				require.Equal(t, expected, updated)
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		f := &flags{
			targetPort:       1,
			targetTLS:        true,
			targetUnixSocket: true,
			proxyAddr:        "1.2.3.4",
			compatPort:       1111,
			compatTLS:        false,
			postgreSQLURL:    "pg url",
			startupEnv:       newStartupInitializer(t, "1.1.1.1:8080,2.2.2.2:8081"),
		}

		targetPort := f.GetTargetPort()
		require.Equal(t, f.targetPort, targetPort)

		targetTLS := f.IsTargetTLS()
		require.Equal(t, f.targetTLS, targetTLS)

		targetUnixSocket := f.IsTargetUnixSocket()
		require.Equal(t, f.targetUnixSocket, targetUnixSocket)

		proxyAddr := f.GetProxyAddr()
		require.Equal(t, f.proxyAddr, proxyAddr)

		compatPort := f.GetCompatPort()
		require.Equal(t, f.compatPort, compatPort)

		compatTLS := f.IsCompatTLS()
		require.Equal(t, f.compatTLS, compatTLS)

		postgreSQLURL := f.GetPostgreSQLURL()
		require.Equal(t, f.postgreSQLURL, postgreSQLURL)

		tigrisURL := f.GetTigrisURL()
		require.Equal(t, "1.1.1.1:8080", tigrisURL)
	})
}

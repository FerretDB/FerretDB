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

// flags are used to override flags set from cli with test setup option.
type flags struct {
	targetPort       int
	targetTLS        bool
	handler          string
	targetUnixSocket bool
	proxyAddr        string
	compatPort       int
	compatTLS        bool
	postgreSQLURL    string
	tigrisURL        string
}

// ApplyOpts applies opts to the flags to override it.
func (f *flags) ApplyOpts(tb testing.TB, opts map[string]any) *flags {
	for k, v := range opts {
		switch k {
		case "target-port":
			targetPort, ok := v.(int)
			require.True(tb, ok, "%s is not int: %T", k, v)

			f.targetPort = targetPort
		case "target-tls":
			targetTLS, ok := v.(bool)
			require.True(tb, ok, "%s is not bool: %T", v)

			f.targetTLS = targetTLS

		case "handler":
			handler, ok := v.(string)
			require.True(tb, ok, "%s is not string: %T", v)

			f.handler = handler
		case "target-unix-socket":
			targetUnixSocket, ok := v.(bool)
			require.True(tb, ok, "%s is not bool: %T", v)

			f.targetUnixSocket = targetUnixSocket
		case "proxy-addr":
			proxyAddr, ok := v.(string)
			require.True(tb, ok, "%s is not string: %T", v)

			f.proxyAddr = proxyAddr
		case "compat-port":
			compatPort, ok := v.(int)
			require.True(tb, ok, "%s is not int: %T", k, v)

			f.compatPort = compatPort
		case "compat-tls":
			compatTLS, ok := v.(bool)
			require.True(tb, ok, "%s is not bool: %T", v)

			f.compatTLS = compatTLS
		case "postgresql-url":
			postgreSQLURL, ok := v.(string)
			require.True(tb, ok, "%s is not string: %T", v)

			f.postgreSQLURL = postgreSQLURL
		case "tigris-url":
			tigrisURL, ok := v.(string)
			require.True(tb, ok, "%s is not string: %T", v)

			f.tigrisURL = tigrisURL
		default:
			tb.Errorf("unknown flag is set: %s", k)
		}
	}

	return f
}

// IsTargetTLS returns true if targetTLS is set.
func (f *flags) IsTargetTLS() bool {
	return f.targetTLS
}

// GetTargetPort returns target port number.
func (f *flags) GetTargetPort() int {
	return f.targetPort
}

// GetHandler returns the handler name.
func (f *flags) GetHandler() string {
	return f.handler
}

// IsTargetUnixSocket returns true if targetUnixSocket is set.
func (f *flags) IsTargetUnixSocket() bool {
	return f.targetUnixSocket
}

// GetProxyAddr returns proxy address.
func (f *flags) GetProxyAddr() string {
	return f.proxyAddr
}

// IsCompatTLS returns true if compatTLS is set.
func (f *flags) IsCompatTLS() bool {
	return f.compatTLS
}

// GetCompatPort returns compat port number.
func (f *flags) GetCompatPort() int {
	return f.compatPort
}

// GetPostgreSQLURL returns postgreSQL url.
func (f *flags) GetPostgreSQLURL() string {
	return f.postgreSQLURL
}

// GetTigrisURL returns tigris url.
func (f *flags) GetTigrisURL() string {
	return f.tigrisURL
}

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

package mcp_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/setup"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestListener(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	configF := setupMCP(t, ctx)

	res := askMCPHost(t, ctx, configF, "list databases")
	t.Log(res)
	//          [  ferretdb__listDatabases
	//          ]  List
	//          {"databases":[],"totalSize":{"$numberInt":"19967123"},"ok":{"$numberDouble":"1
	//          .0"}}
	res = strings.ReplaceAll(res, " ", "")
	res = strings.ReplaceAll(res, "\n", "")
	require.Contains(t, res, "ferretdb__listDatabases")
	require.Contains(t, res, `{"databases":[`)
	require.Contains(t, res, `],"totalSize":{`)
	require.Contains(t, res, `},"ok":{"$numberDouble":"1.0"}`)
}

// askMCPHost sends query to MCP host in non-interactive mode with
// the given config file and prompt and returns the output.
// Non-interactive mode without streaming is used for the ease of testing.
func askMCPHost(tb testing.TB, ctx context.Context, configF, prompt string) string {
	tb.Helper()

	// we probably should do something better with quoting
	config := `MCPHOST_CONFIG=` + configF
	args := fmt.Sprintf(`MCPHOST_ARGS=--stream=false --compact --prompt=%q`, prompt)
	cmd := exec.CommandContext(ctx, filepath.Join(testutil.BinDir, "task"), "mcphost", config, args)

	res, err := cmd.CombinedOutput()
	require.NoError(tb, err, string(res))

	return string(res)
}

// setupMCP sets up a new MCP listener and returns config file path
// that mcp host can use.
func setupMCP(tb testing.TB, ctx context.Context) string {
	tb.Helper()

	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	//exhaustruct:enforce
	res := setup.Setup(tb.Context(), &setup.SetupOpts{
		Logger:        testutil.Logger(tb),
		StateProvider: sp,
		Metrics:       middleware.NewMetrics(),

		PostgreSQLURL:          testutil.PostgreSQLURL(tb),
		Auth:                   false,
		ReplSetName:            "",
		SessionCleanupInterval: 0,

		ProxyAddr:        "",
		ProxyTLSCertFile: "",
		ProxyTLSKeyFile:  "",
		ProxyTLSCAFile:   "",

		TCPAddr:        "127.0.0.1:0",
		UnixAddr:       "",
		TLSAddr:        "",
		TLSCertFile:    "",
		TLSKeyFile:     "",
		TLSCAFile:      "",
		Mode:           middleware.NormalMode,
		TestRecordsDir: "",

		DataAPIAddr: "",

		MCPAddr: "127.0.0.1:0",
	})
	require.NotNil(tb, res)

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	runDone := make(chan struct{})

	go func() {
		defer close(runDone)
		res.Run(ctx)
	}()

	// ensure that all listener's and handler's logs are written before test ends
	tb.Cleanup(func() {
		cancel()
		<-runDone
	})

	config := fmt.Sprintf(`{
	"mcpServers": {
	  "FerretDB": {
	    "type": "remote",
	    "url": "http://%s/mcp"
	    }
	  }
	}`,
		res.MCPListener.Addr(),
	)

	configF := filepath.Join(tb.TempDir(), "mcphost.json")
	err = os.WriteFile(configF, []byte(config), 0o666)
	require.NoError(tb, err)

	return configF
}

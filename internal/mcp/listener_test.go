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

// llmModel is the model used in tests.
const llmModel = "qwen3:0.6b" // sync with Taskfile.yml

func TestBasic(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB/issues/5209
	t.Skip("https://github.com/FerretDB/FerretDB/issues/5209")

	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := t.Context()
	configF := setupMCP(t, ctx)
	db := t.Name()

	t.Run("Insert", func(t *testing.T) {
		prompt := fmt.Sprintf("Use database named %s. "+
			"Insert two documents to a collection named authors. "+
			"The first document should contain author name Jane Austen with nationality British "+
			"and the second document should contain author name Herman Melville with nationality American.",
			db,
		)
		res := askMCPHost(t, ctx, configF, prompt)

		require.Contains(t, res, "ferretdb__insert")
		require.Contains(t, res, `{"n":{"$numberInt":"2"},"ok":{"$numberDouble":"1.0"}}`)
	})

	t.Run("Find", func(t *testing.T) {
		prompt := fmt.Sprintf("Find a British author from %s database authors collection.", db)
		res := askMCPHost(t, ctx, configF, prompt)

		require.Contains(t, res, "ferretdb__find")
		require.Contains(t, res, "Jane Austen")
	})

	t.Run("ListCollections", func(t *testing.T) {
		prompt := fmt.Sprintf("List all collections in %s database.", db)
		res := askMCPHost(t, ctx, configF, prompt)

		require.Contains(t, res, "ferretdb__listCollections")
		require.Contains(t, res, `authors`)
	})

	t.Run("DropDatabase", func(t *testing.T) {
		prompt := fmt.Sprintf("Delete database named %s.", db)
		res := askMCPHost(t, ctx, configF, prompt)

		require.Contains(t, res, "ferretdb__dropDatabase")
		require.Contains(t, res, `{"ok":{"$numberDouble":"1.0"}}`)
	})
}

func TestAdmin(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB/issues/5209
	t.Skip("https://github.com/FerretDB/FerretDB/issues/5209")

	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := t.Context()
	configF := setupMCP(t, ctx)

	t.Run("ListDatabases", func(t *testing.T) {
		res := askMCPHost(t, ctx, configF, "list databases")

		require.Contains(t, res, "ferretdb__listDatabases")
		require.Contains(t, res, `{"databases":[`)
		require.Contains(t, res, `],"totalSize":{`)
		require.Contains(t, res, `},"ok":{"$numberDouble":"1.0"}`)
	})
}

// askMCPHost sends query to MCP host in non-interactive mode with
// the given config file and prompt.
// Non-interactive mode without streaming is used for the ease of testing.
func askMCPHost(tb testing.TB, ctx context.Context, configF, prompt string) string {
	tb.Helper()

	cmd := exec.CommandContext(
		ctx,
		filepath.Join(testutil.BinDir, "mcphost"),
		"--compact=true",
		"--config="+configF,
		"--model=ollama:"+llmModel,
		"--quiet=true",
		"--stream=false",
		"--temperature=0.0",
		"--prompt", prompt,
	)
	tb.Logf("%#q", cmd.Args)

	output, err := cmd.Output()
	require.NoError(tb, err)

	res := string(output)
	tb.Logf("output:\n%s", res)
	//          [  ferretdb__listDatabases
	//          ]  List
	//          {"databases":[],"totalSize":{"$numberInt":"19967123"},"ok":{"$numberDouble":"1
	//          .0"}}
	//
	// remove tabs and newlines to avoid split
	res = strings.ReplaceAll(res, "\t", " ")
	res = strings.ReplaceAll(res, "\n", " ")

	return res
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

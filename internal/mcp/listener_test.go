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

package mcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestListenerNoAuth(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	addr := setupListener(t, ctx)

	jsonConfig := fmt.Sprintf(`{
	"mcpServers": {
	  "FerretDB": {
	    "type": "remote",
	    "url": "http://%s/mcp"
	    }
	  }
	}`,
		addr.String(),
	)

	res := askMCPHost(t, ctx, jsonConfig, "list databases")
	t.Log(res)
	//          [  ferretdb__listDatabases
	//          ]  List
	//          {"databases":[],"totalSize":{"$numberInt":"19967123"},"ok":{"$numberDouble":"1
	//          .0"}}
	require.Contains(t, res, "ferretdb__listDatabases")

	res = strings.ReplaceAll(res, "\n", "")
	res = strings.ReplaceAll(res, " ", "")
	require.Contains(t, res, `{"databases":`)
	require.Contains(t, res, `"totalSize":`)
	require.Contains(t, res, `"ok":{"$numberDouble":"1.0"}`)
}

// askMCPHost runs MCP host in non-interactive mode with the given config and prompt and returns the output.
// Non-interactive mode without streaming is used for the ease of testing.
func askMCPHost(tb testing.TB, ctx context.Context, jsonConfig, prompt string) string {
	tb.Helper()

	bin := filepath.Join(testutil.BinDir, "mcphost")

	configF := filepath.Join(tb.TempDir(), "mcphost.json")
	err := os.WriteFile(configF, []byte(jsonConfig), 0o666)
	require.NoError(tb, err)

	cmd := exec.CommandContext(ctx,
		bin,
		"--config", configF,
		"--model", "ollama:qwen3:0.6b",
		"--prompt", prompt,
		"--stream=false",
		"--compact",
	)
	res, err := cmd.CombinedOutput()
	assert.NoError(tb, err)

	return string(res)
}

// setupListener sets up a new MCP listener.
func setupListener(tb testing.TB, ctx context.Context) net.Addr {
	uri := testutil.PostgreSQLURL(tb)
	l := testutil.Logger(tb)
	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	p, err := documentdb.NewPool(uri, l, sp)
	require.NoError(tb, err)

	h, err := handler.New(&handler.NewOpts{
		Pool:          p,
		L:             l,
		StateProvider: sp,
	})
	require.NoError(tb, err)

	handlerCtx, cancel := context.WithCancel(ctx)
	handlerDone := make(chan struct{})

	go func() {
		h.Run(handlerCtx)
		close(handlerDone)
	}()

	tb.Cleanup(func() {
		cancel()
		<-handlerDone
	})

	lis, err := Listen(&ListenerOpts{
		L:           l,
		Handler:     h,
		ToolHandler: NewToolHandler(h),
		TCPAddr:     "127.0.0.1:0",
	})
	require.NoError(tb, err)

	listenDone := make(chan struct{})

	go func() {
		err = lis.Run(ctx)
		assert.NoError(tb, err)
		close(listenDone)
	}()

	tb.Cleanup(func() {
		<-listenDone
	})

	return lis.lis.Addr()
}

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
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestServer(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	uri := testutil.PostgreSQLURL(t)
	l := testutil.Logger(t)
	sp, err := state.NewProvider("")
	require.NoError(t, err)

	p, err := documentdb.NewPool(uri, l, sp)
	require.NoError(t, err)

	h, err := handler.New(&handler.NewOpts{
		Pool:          p,
		L:             l,
		StateProvider: sp,
	})
	require.NoError(t, err)

	handlerCtx, cancel := context.WithCancel(ctx)
	handlerDone := make(chan struct{})

	go func() {
		defer close(handlerDone)

		h.Run(handlerCtx)
	}()

	t.Cleanup(func() {
		cancel()
		<-handlerDone
	})

	s := New(&ServerOpts{
		L:           l,
		ToolHandler: NewToolHandler(h),
		TCPAddr:     "127.0.0.1:8081",
	})

	serverDone := make(chan struct{})

	go func() {
		defer close(serverDone)

		err = s.Serve(ctx)
		assert.NoError(t, err)
	}()

	t.Cleanup(func() {
		<-serverDone
	})

	res := askMCPHost(t, ctx, "list databases")
	t.Log(res)
	//        ┃ 🔧 Calling ferretdb__listDatabases
	//        ┃  Tool Call (25 Jun 2025 11:21 AM)
	//
	//        ┃ ferretdb__listDatabases: {}
	//        ┃ {"content":[{"type":"text","text":"{\"databases\":[],\"totalSize\":{\"$numberI
	//        ┃ nt\":\"18377875\"},\"ok\":{\"$numberDouble\":\"1.0\"}}"}]}
	require.Contains(t, res, "Calling ferretdb__listDatabases")
	require.Contains(t, res, "ferretdb__listDatabases: {}")
	require.Contains(t, res, `\"ok\":{\"$numberDouble\":\"1.0\"}`)
}

// askMCPHost runs MCP host in non-interactive mode with the given prompt and returns the output.
// Non-interactive mode is used for the ease of testing.
func askMCPHost(tb testing.TB, ctx context.Context, prompt string) string {
	tb.Helper()

	bin := filepath.Join(testutil.BinDir, "mcphost")
	config := filepath.Join(testutil.RootDir, "build", "mcp", "mcphost.json")

	cmd := exec.CommandContext(ctx, bin, "--config", config, "--model", "ollama:qwen3:0.6b", "--prompt", prompt)
	res, err := cmd.CombinedOutput()
	require.NoError(tb, err)

	return string(res)
}

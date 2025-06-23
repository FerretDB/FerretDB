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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
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

	p, err := documentdb.NewPool(uri, logging.WithName(l, "pool"), sp)
	require.NoError(t, err)

	handlerOpts := &handler.NewOpts{
		Pool:          p,
		L:             logging.WithName(l, "handler"),
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
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
		TCPAddr:     "127.0.0.1:0",
	})

	serverDone := make(chan struct{})

	go func() {
		defer close(serverDone)

		_ = s.Serve(ctx)
	}()

	t.Cleanup(func() {
		err = s.Shutdown(ctx)
		assert.NoError(t, err)

		<-serverDone
	})

	res, err := askMCPHost(ctx, "list databases")
	require.NoError(t, err)

	t.Log(string(res))
	require.Contains(t, string(res), "Calling ferretdb__listDatabases")
}

func TestMCPHost(t *testing.T) {
	res, err := askMCPHost(t.Context(), "list databases")
	require.NoError(t, err)

	t.Log(string(res))
	require.Contains(t, string(res), "Calling ferretdb__listDatabases")
}

// askMCPHost runs MCP host in non-interactive mode with the given prompt and returns the output.
// Non-interactive mode is used for the ease of testing.
func askMCPHost(ctx context.Context, prompt string) ([]byte, error) {
	rootDir := getRoot()

	bin := filepath.Join(rootDir, "bin", "mcphost")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	bin, err := filepath.Abs(bin)
	if err != nil {
		return nil, err
	}

	config := filepath.Join(rootDir, "build", "mcp", "mcphost.json")
	if _, err = os.Stat(config); err != nil {
		return nil, err
	}

	config, err = filepath.Abs(config)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, bin, "--config", config, "--model", "ollama:qwen3:0.6b", "--prompt", prompt)

	return cmd.CombinedOutput()
}

// getRoot finds the repository root directory by looking for the go.mod file.
func getRoot() string {
	var dir string

	for {
		goModFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModFile); err == nil {
			return dir
		}

		dir = filepath.Join("..", dir)
	}
}

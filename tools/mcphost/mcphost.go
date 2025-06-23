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

// Package mcphost provides a host to interact with the MCP server.
package mcphost

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// AskMCPHost runs MCP host in non-interactive mode with the given prompt and returns the output.
// Non-interactive mode is used for the ease of testing.
func AskMCPHost(ctx context.Context, prompt string) ([]byte, error) {
	bin := filepath.Join("..", "..", "bin", "mcphost")
	if _, err := os.Stat(bin); err != nil {
		return nil, err
	}

	bin, err := filepath.Abs(bin)
	if err != nil {
		return nil, err
	}

	config := filepath.Join("..", "..", "build", "mcp", "mcphost.json")
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

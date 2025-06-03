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
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestHandle(t *testing.T) {
	t.Parallel()

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

	mh := NewHandler(h, logging.WithName(l, "handler"))

	ctx := conninfo.Ctx(context.Background(), conninfo.New())

	t.Run("insert", func(t *testing.T) {
		t.Parallel()

		var req mcp.CallToolRequest
		req.Params.Name = "insert"
		req.Params.Arguments = map[string]any{
			"database":   "test",
			"collection": "values",
		}

		var res *mcp.CallToolResult
		res, err = mh.insert(ctx, req)
		require.NoError(t, err)
		assert.False(t, res.IsError, res.Content)
	})

	t.Run("find", func(t *testing.T) {
		t.Parallel()

		var req mcp.CallToolRequest
		req.Params.Name = "find"
		req.Params.Arguments = map[string]any{
			"database":   "test",
			"collection": "values",
		}

		var res *mcp.CallToolResult
		res, err = mh.find(ctx, req)
		require.NoError(t, err)
		assert.False(t, res.IsError, res.Content)
	})
}

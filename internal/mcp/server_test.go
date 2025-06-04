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
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
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

	t.Cleanup(p.Close)

	handlerOpts := &handler.NewOpts{
		Pool:          p,
		L:             logging.WithName(l, "handler"),
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
	require.NoError(t, err)

	ctx := conninfo.Ctx(t.Context(), conninfo.New())
	mh := NewHandler(h, logging.WithName(l, "mcp-handler"))

	//nolint:vet // for testing
	type params struct {
		Name      string    `json:"name"`
		Arguments any       `json:"arguments,omitempty"`
		Meta      *mcp.Meta `json:"_meta,omitempty"`
	}

	for name, tc := range map[string]struct {
		req        mcp.CallToolRequest
		handleFunc server.ToolHandlerFunc
	}{
		"insertRawDocuments": {
			req: mcp.CallToolRequest{
				Params: params{
					Name: "insert",
					Arguments: map[string]any{
						"collection": "values",
						"database":   "test",
						"documents":  must.NotFail(json.Marshal([]any{map[string]any{"abc": "def"}})),
					},
				},
			},
			handleFunc: mh.insert,
		},
		"insertMapDocuments": {
			req: mcp.CallToolRequest{
				Params: params{
					Name: "insert",
					Arguments: map[string]any{
						"collection": "values",
						"database":   "test",
						"documents":  []any{map[string]any{"abc": "def"}},
					},
				},
			},
			handleFunc: mh.insert,
		},
		"find": {
			req: mcp.CallToolRequest{
				Params: params{
					Name: "find",
					Arguments: map[string]any{
						"database":   "test",
						"collection": "values",
					},
				},
			},
			handleFunc: mh.find,
		},
		"listCollections": {
			req: mcp.CallToolRequest{
				Params: params{
					Name: "list collections",
					Arguments: map[string]any{
						"database": "test",
					},
				},
			},
			handleFunc: mh.listCollections,
		},
		"listDatabases": {
			req: mcp.CallToolRequest{
				Params: params{
					Name: "list databases",
				},
			},
			handleFunc: mh.listDatabases,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// do not run parallel to avoid datarace
			var res *mcp.CallToolResult
			res, err = tc.handleFunc(ctx, tc.req)
			require.NoError(t, err)
			assert.False(t, res.IsError, res.Content)
		})
	}
}

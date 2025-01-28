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

package dataapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmokeDataAPI(t *testing.T) {
	const uri = "postgres://username:password@127.0.0.1:5432/postgres"

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	l := testutil.Logger(t)

	p, err := documentdb.NewPool(uri, l, sp)
	require.NoError(t, err)

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: false,

		L:             logging.WithName(l, "handler"),
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
	require.NoError(t, err)

	var lis *Listener
	lis, err = Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       l,
		Handler: h,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	done := make(chan struct{})

	go func() {
		lis.Run(ctx)
		close(done)
	}()

	addr := lis.lis.Addr().String()
	db := testutil.DatabaseName(t)
	coll := testutil.CollectionName(t)

	c := http.Client{}

	t.Run("Find", func(t *testing.T) {
		jb, err := json.Marshal(map[string]any{
			"database":   db,
			"collection": coll,
			"filter":     map[string]any{},
		})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/action/find", bytes.NewBuffer(jb))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := c.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		resB, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, nil, string(resB))
	})

	cancel()
	<-done // prevent panic on logging after test ends
}

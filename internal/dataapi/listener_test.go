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
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestSmokeDataAPI(t *testing.T) {
	addr, db := setup(t)
	coll := testutil.CollectionName(t)

	t.Run("Find", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {}
		}`

		res, err := request(t, "http://"+addr+"/action/find", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"documents":[]}`, string(body))
	})

	insertDoc := `{"_id":1,"foo":"bar"}`

	t.Run("InsertOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"document": ` + insertDoc + `
		}`

		res, err := request(t, "http://"+addr+"/action/insertOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"n":1}`, string(body))
	})

	t.Run("FindAfterInsertOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {}
		}`

		res, err := request(t, "http://"+addr+"/action/find", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"documents":[`+insertDoc+`]}`, string(body))
	})

	// TODO every operation
}

func request(tb testing.TB, uri, jsonBody string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer([]byte(jsonBody)))
	require.NoError(tb, err)
	req.Header.Set("Content-Type", "application/json")

	req.SetBasicAuth("username", "password")

	return http.DefaultClient.Do(req)
}

func setup(tb testing.TB) (addr string, dbName string) {
	const uri = "postgres://username:password@127.0.0.1:5432/postgres"

	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	l := testutil.Logger(tb)

	p, err := documentdb.NewPool(uri, logging.WithName(l, "pool"), sp)
	require.NoError(tb, err)

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: true,

		L:             logging.WithName(l, "handler"),
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
	require.NoError(tb, err)

	listenerOpts := clientconn.NewListenerOpts{
		Mode:    clientconn.NormalMode,
		Metrics: connmetrics.NewListenerMetrics(),
		Handler: h,
		Logger:  logging.WithName(l, "listener"),
		TCP:     "127.0.0.1:0",
	}

	lis, err := clientconn.Listen(&listenerOpts)
	require.NoError(tb, err)

	var apiLis *Listener
	apiLis, err = Listen(&ListenOpts{
		TCPAddr: "127.0.0.1:0",
		L:       logging.WithName(l, "dataapi"),
		Handler: h,
	})
	require.NoError(tb, err)

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))
	var wg sync.WaitGroup

	// ensure that all listener's and handler's logs are written before test ends
	tb.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		lis.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		apiLis.Run(ctx)
	}()

	u := &url.URL{
		Scheme: "mongodb",
		Host:   lis.TCPAddr().String(),
		Path:   "/",
		User:   url.UserPassword("username", "password"),
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	require.NoError(tb, err)

	addr = apiLis.lis.Addr().String()
	dbName = testutil.DatabaseName(tb)

	err = client.Database(dbName).Drop(ctx)
	require.NoError(tb, err)
	return
}

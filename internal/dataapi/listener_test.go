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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestSmokeDataAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	addr, db := setupDataAPI(t)
	coll := testutil.CollectionName(t)

	t.Run("FindEmpty", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/find", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"documents":[]}`, string(body))
	})

	docs := []string{
		`{"_id":1,"v":"foo"}`,
		`{"_id":2,"v":42}`,
		`{"_id":3,"v":{"foo":"bar"}}`,
	}

	t.Run("InsertOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"document": ` + docs[0] + `
		}`

		res, err := postJSON(t, "http://"+addr+"/action/insertOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"n":1}`, string(body))
	})

	t.Run("InsertMany", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"documents": [` + strings.Join(docs[1:3], ",") + `]
		}`

		res, err := postJSON(t, "http://"+addr+"/action/insertMany", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"n":1}`, string(body))
	})

	t.Run("UpdateOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":42},
			"update": {"$set":{"v":"foo"}}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/updateOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"matchedCount":1,"modifiedCount":1}`, string(body))
	})

	t.Run("UpdateMany", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":"foo"},
			"update": {"$set":{"v":42.13}}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/updateMany", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"matchedCount":2,"modifiedCount":2}`, string(body))
	})

	t.Run("FindOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":42.13}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/findOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"document":{"_id":1,"v":42.13}}`, string(body))
	})

	t.Run("Aggregate", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"pipeline": [{"$group":{"_id": "$v"}}]
		}`

		res, err := postJSON(t, "http://"+addr+"/action/aggregate", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"documents":[{"_id":42.13},{"_id":{"foo":"bar"}}]}`, string(body))
	})

	t.Run("DeleteOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":42.13}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/deleteOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"deletedCount":1}`, string(body))
	})

	t.Run("DeleteMany", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":42.13}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/deleteMany", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"deletedCount":1}`, string(body))
	})

	t.Run("Find", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/find", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"documents":[`+docs[2]+`]}`, string(body))
	})
}

// postJSON sends POST request with provided JSON to data API under provided uri.
// It handles necessary headers, as well as authentication.
func postJSON(tb testing.TB, uri, jsonBody string) (*http.Response, error) {
	tb.Helper()

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer([]byte(jsonBody)))
	require.NoError(tb, err)

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("username", "password")

	return http.DefaultClient.Do(req)
}

// setupDataAPI sets up clean database and the Data API handler.
// It returns Data API address, and database name.
func setupDataAPI(tb testing.TB) (addr string, dbName string) {
	tb.Helper()
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

	_ = client.Disconnect(ctx)

	return
}

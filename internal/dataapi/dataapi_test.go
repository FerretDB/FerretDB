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

package dataapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/setup"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestSmokeDataAPI(t *testing.T) {
	addr, db := setupDataAPI(t, true)
	coll := testutil.CollectionName(t)

	t.Parallel()

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

	t.Run("FindOneEmpty", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {}
		}`

		res, err := postJSON(t, "http://"+addr+"/action/findOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"document":null}`, string(body))
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
		assert.JSONEq(t, `{"insertedId":1}`, string(body))
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
		assert.JSONEq(t, `{"insertedIds": [2, 3]}`, string(body))
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

	t.Run("InsertOneEmptyId", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"document": ` + `{"v":"foo"}` + `
		}`

		res, err := postJSON(t, "http://"+addr+"/action/insertOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		bodyStr := strings.TrimSpace(string(body))
		assert.Regexp(t,
			`^{"insertedId":"[0-9a-f]{24}"}$`,
			bodyStr,
			"expected %q, got %q",
			`{"insertedId":"<ObjectID>"}`,
			string(body),
		)
	})

	t.Run("UpsertOne", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":"non-existent"},
			"update": {"$set":{"v":"foo"}},
			"upsert": true
		}`

		res, err := postJSON(t, "http://"+addr+"/action/updateOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		bodyStr := strings.TrimSpace(string(body))
		assert.Regexp(t,
			`^{"matchedCount":1,"modifiedCount":0,"upsertedId":".*[0-9a-f]{24}.*"}$`,
			bodyStr,
			"expected %q, got %q",
			`{"upsertedId":"<ObjectID>"}`,
			string(body),
		)
	})

	t.Run("UpsertOneStringId", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":"non-existent"},
			"update": {"$set":{"_id":"string_id","v":"foo"}},
			"upsert": true
		}`

		res, err := postJSON(t, "http://"+addr+"/action/updateOne", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.NoError(t, err)
		assert.JSONEq(t, `{"matchedCount":1,"modifiedCount":0,"upsertedId":"string_id"}`, string(body))
	})

	t.Run("UpsertMany", func(t *testing.T) {
		jsonBody := `{
			"database": "` + db + `",
			"collection": "` + coll + `",
			"filter": {"v":"non-existent"},
			"update": {"$set":{"v":"foo"}},
			"upsert": true
		}`

		res, err := postJSON(t, "http://"+addr+"/action/updateMany", jsonBody)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		bodyStr := strings.TrimSpace(string(body))
		assert.Regexp(t,
			`^{"matchedCount":1,"modifiedCount":0,"upsertedId":".*[0-9a-f]{24}.*"}$`,
			bodyStr,
			"expected %q, got %q",
			`{"upsertedId":"<ObjectID>"}`,
			string(body),
		)
	})
}

func TestOpenAPI(t *testing.T) {
	t.Parallel()

	addr, _ := setupDataAPI(t, false)

	t.Run("Spec", func(t *testing.T) {
		resp, err := http.Get("http://" + addr + "/openapi.json")
		require.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, resp.Body.Close())
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "53813", resp.Header.Get("Content-Length"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var spec map[string]any
		err = json.Unmarshal(body, &spec)
		require.NoError(t, err)

		assert.Equal(t, "3.0.0", spec["openapi"])
		assert.Contains(t, spec, "info")
		assert.Contains(t, spec, "paths")

		info := spec["info"].(map[string]any)
		assert.Equal(t, "FerretDB Data API", info["title"])
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		resp, err := http.Post("http://"+addr+"/openapi.json", "", nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, resp.Body.Close())
		})

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
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
func setupDataAPI(tb testing.TB, auth bool) (addr string, dbName string) {
	tb.Helper()

	uri := testutil.PostgreSQLURL(tb)

	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	l := testutil.Logger(tb)

	//exhaustruct:enforce
	res := setup.Setup(tb.Context(), &setup.SetupOpts{
		Logger:        l,
		StateProvider: sp,
		Metrics:       middleware.NewMetrics(),

		PostgreSQLURL:          uri,
		Auth:                   auth,
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

		DataAPIAddr: "127.0.0.1:0",
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

	u := &url.URL{
		Scheme: "mongodb",
		Host:   res.WireListener.TCPAddr().String(),
		Path:   "/",
	}

	if auth {
		u.User = url.UserPassword("username", "password")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(u.String()))
	require.NoError(tb, err)

	addr = res.DataAPIListener.Addr().String()
	dbName = testutil.DatabaseName(tb)

	err = client.Database(dbName).Drop(ctx)
	require.NoError(tb, err)

	_ = client.Disconnect(ctx)

	return
}

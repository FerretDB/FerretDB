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

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestKillAllSessions(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	collection := db.Collection(s.Collection.Name())
	cName, dbName := collection.Name(), db.Name()

	user1, user2 := t.Name()+"1", t.Name()+"2"

	user1Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user1)
	user2Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user2)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	_, err := collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "a"}},
		bson.D{{"_id", "b"}},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("OneUser", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)

		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		user2SessionID := startSession(t, ctx, user2Conn)

		user2CursorID := find(t, ctx, user2Conn, dbName, cName, user2SessionID)

		users := wirebson.MustArray(wirebson.MustDocument("db", dbName, "user", user1))

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, expectedErr)

		getMore(t, ctx, user2Conn, dbName, cName, user2SessionID, user2CursorID, nil)
	})

	t.Run("TwoUsers", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)

		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		user2SessionID := startSession(t, ctx, user2Conn)

		user2CursorID := find(t, ctx, user2Conn, dbName, cName, user2SessionID)

		users := wirebson.MustArray(
			wirebson.MustDocument("db", dbName, "user", user1),
			wirebson.MustDocument("db", dbName, "user", user2),
		)

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)

		expectedErr1 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, expectedErr1)

		expectedErr2 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user2CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, user2SessionID, user2CursorID, expectedErr2)
	})

	t.Run("OneUserAllSessions", func(t *testing.T) {
		sessionID1 := startSession(t, ctx, user1Conn)

		cursorID1 := find(t, ctx, user1Conn, dbName, cName, sessionID1)

		sessionID2 := startSession(t, ctx, user1Conn)

		cursorID2 := find(t, ctx, user1Conn, dbName, cName, sessionID2)

		users := wirebson.MustArray(wirebson.MustDocument("db", dbName, "user", user1))

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)

		expectedErr1 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID1),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID1, cursorID1, expectedErr1)

		expectedErr2 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID2),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID2, cursorID2, expectedErr2)
	})

	t.Run("KillMultipleTimes", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)

		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		users := wirebson.MustArray(wirebson.MustDocument("db", dbName, "user", user1))

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, user1Conn, dbName, sessions, nil)

		killSessions(t, ctx, user1Conn, dbName, sessions, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("NoneExistentUser", func(t *testing.T) {
		users := wirebson.MustArray(wirebson.MustDocument("db", dbName, "user", "nonexistent"))

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)
	})

	t.Run("NonExistentUID", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)

		users := wirebson.MustArray(wirebson.MustDocument("db", dbName, "user", user1))

		killAllSessions(t, ctx, user1Conn, dbName, users, nil)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		endSessions(t, ctx, user1Conn, dbName, sessions, nil)
	})
}

func TestKillAllSessionsErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	dbName := db.Name()

	user := t.Name()
	conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("NotArrayUsers", func(t *testing.T) {
		users := "invalid"

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions' is the wrong type 'string', expected type 'array'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})

	t.Run("NotDocumentUser", func(t *testing.T) {
		users := wirebson.MustArray("invalid")

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions.0' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})

	t.Run("MissingDB", func(t *testing.T) {
		users := wirebson.MustArray(wirebson.MustDocument("user", "foo"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions.db' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})

	t.Run("BoolDB", func(t *testing.T) {
		users := wirebson.MustArray(wirebson.MustDocument("db", true, "user", "foo"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions.db' is the wrong type 'bool', expected type 'string'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})

	t.Run("MissingUser", func(t *testing.T) {
		users := wirebson.MustArray(wirebson.MustDocument("db", "bar"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions.user' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})

	t.Run("BoolUser", func(t *testing.T) {
		users := wirebson.MustArray(wirebson.MustDocument("db", "bar", "user", false))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsCmd.killAllSessions.user' is the wrong type 'bool', expected type 'string'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessions(t, ctx, conn, dbName, users, expectedErr)
	})
}

func TestKillAllSessionsAllUsers(t *testing.T) {
	// do not run in parallel as this test kill sessions of other tests
	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx

	testutil.Exclusive(ctx, "this test kills sessions of other tests (except for in-process FerretDB)")

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	collection := db.Collection(s.Collection.Name())
	cName, dbName := collection.Name(), db.Name()

	user1, user2 := t.Name()+"1", t.Name()+"2"

	user1Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user1)
	user2Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user2)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	_, err := collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "a"}},
		bson.D{{"_id", "b"}},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	user1SessionID := startSession(t, ctx, user1Conn)

	user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

	user2SessionID := startSession(t, ctx, user2Conn)

	user2CursorID := find(t, ctx, user2Conn, dbName, cName, user2SessionID)

	users := wirebson.MustArray()

	killAllSessions(t, ctx, user1Conn, dbName, users, nil)

	expectedErr1 := must.NotFail(wirebson.NewDocument(
		"ok", float64(0),
		"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
		"code", int32(43),
		"codeName", "CursorNotFound",
	))

	getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, expectedErr1)

	expectedErr2 := must.NotFail(wirebson.NewDocument(
		"ok", float64(0),
		"errmsg", fmt.Sprintf("cursor id %d not found", user2CursorID),
		"code", int32(43),
		"codeName", "CursorNotFound",
	))

	getMore(t, ctx, user1Conn, dbName, cName, user2SessionID, user2CursorID, expectedErr2)
}

// killAllSessions sends a request to kill all sessions.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
func killAllSessions(t testing.TB, ctx context.Context, conn *wireclient.Conn, dbName string, users any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"killAllSessions", users,
		"$db", dbName,
	)

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	if expectedErr != nil {
		testutil.AssertEqual(t, expectedErr, res)

		return
	}

	expected := must.NotFail(wirebson.NewDocument(
		"ok", float64(1),
	))

	testutil.AssertEqual(t, expected, res)
}

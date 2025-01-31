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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestKillSessions(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	collection := db.Collection(s.Collection.Name())
	cName, dbName := collection.Name(), db.Name()

	user := t.Name()
	conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user)

	_, err := collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "a"}},
		bson.D{{"_id", "b"}},
	})
	require.NoError(t, err)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("OneSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)
	})

	t.Run("NonExistentSession", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		sessionID := wirebson.Binary{
			B:       randomUUID[:],
			Subtype: wirebson.BinaryUUID,
		}

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)
	})

	t.Run("MultipleSessions", func(t *testing.T) {
		sessionID1 := startSession(t, ctx, conn)
		sessionID2 := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(
			wirebson.MustDocument("id", sessionID1),
			wirebson.MustDocument("id", sessionID2),
		)

		killSessions(t, ctx, conn, dbName, sessions, nil)
	})

	t.Run("KillKilledSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)

		killSessions(t, ctx, conn, dbName, sessions, nil)
	})

	t.Run("ReuseKilledSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)

		cursorID := find(t, ctx, conn, dbName, cName, sessionID)

		getMore(t, ctx, conn, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("CursorOfKilledSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		cursorID := find(t, ctx, conn, dbName, cName, sessionID)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("RefreshKilledSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		cursorID := find(t, ctx, conn, dbName, cName, sessionID)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)

		refreshSessions(t, ctx, conn, dbName, nil, sessions, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("EndKilledSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		cursorID := find(t, ctx, conn, dbName, cName, sessionID)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, conn, dbName, sessions, nil)

		endSessions(t, ctx, conn, dbName, sessions, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("EmptyUUIDSessionID", func(t *testing.T) {
		sessions := wirebson.MustArray(wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryUUID}))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "uuid must be a 16-byte binary field with UUID (4) subtype",
			"code", int32(207),
			"codeName", "InvalidUUID",
		))

		killSessions(t, ctx, conn, dbName, sessions, expectedErr)
	})
}

func TestKillSessionsDifferentUser(t *testing.T) {
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

	t.Run("CannotKillOtherUserSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)

		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		killSessions(t, ctx, user2Conn, dbName, sessions, nil)

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("EmptySessions", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)
		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		user2SessionID := startSession(t, ctx, user2Conn)
		user2CursorID := find(t, ctx, user2Conn, dbName, cName, user2SessionID)

		users := wirebson.MustArray()

		killSessions(t, ctx, user2Conn, dbName, users, nil)

		getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user2CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user2Conn, dbName, cName, user2SessionID, user2CursorID, expectedErr)
	})
}

// killSessions sends a request to kill the given sessions.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
func killSessions(t testing.TB, ctx context.Context, conn *wireclient.Conn, dbName string, sessions any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"killSessions", sessions,
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

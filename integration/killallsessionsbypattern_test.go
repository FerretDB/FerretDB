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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestKillAllSessionsByPattern(t *testing.T) {
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

	t.Run("KillOwnLSID", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)

		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		user1UID := sha256Binary(user1 + "@" + dbName)

		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", sessionID,
				"uid", wirebson.Binary{B: user1UID},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("KillOtherUserLSID", func(t *testing.T) {
		sessionID := startSession(t, ctx, user2Conn)

		cursorID := find(t, ctx, user2Conn, dbName, cName, sessionID)

		user2ID := sha256Binary(user2 + "@" + dbName)

		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", sessionID,
				"uid", wirebson.Binary{B: user2ID},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user2Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("DoesNotKillOtherLSID", func(t *testing.T) {
		killSessionID := startSession(t, ctx, user1Conn)

		sessionID := startSession(t, ctx, user1Conn)

		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		userID := sha256Binary(user1 + "@" + dbName)

		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", killSessionID,
				"uid", wirebson.Binary{B: userID},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("KillMultipleLSIDs", func(t *testing.T) {
		session1ID := startSession(t, ctx, user1Conn)
		session2ID := startSession(t, ctx, user1Conn)

		cursorID1 := find(t, ctx, user1Conn, dbName, cName, session1ID)
		cursorID2 := find(t, ctx, user1Conn, dbName, cName, session2ID)

		userID := sha256Binary(user1 + "@" + dbName)

		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", session1ID,
				"uid", wirebson.Binary{B: userID},
			)),
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", session2ID,
				"uid", wirebson.Binary{B: userID},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr1 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID1),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, session1ID, cursorID1, expectedErr1)

		expectedErr2 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID2),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, session2ID, cursorID2, expectedErr2)
	})

	t.Run("DifferentUserSameSessionID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())
		sessionID := wirebson.Binary{B: randomUUID[:], Subtype: wirebson.BinaryUUID}

		user1CursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)
		user2CursorID := find(t, ctx, user2Conn, dbName, cName, sessionID)

		userID1 := sha256Binary(user1 + "@" + dbName)
		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", sessionID,
				"uid", wirebson.Binary{B: userID1},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, sessionID, user1CursorID, expectedErr)

		getMore(t, ctx, user2Conn, dbName, cName, sessionID, user2CursorID, nil)
	})

	t.Run("MultipleCursorsSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)
		cursorID1 := find(t, ctx, user1Conn, dbName, cName, sessionID)
		cursorID2 := find(t, ctx, user1Conn, dbName, cName, sessionID)

		userID1 := sha256Binary(user1 + "@" + dbName)
		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", sessionID,
				"uid", wirebson.Binary{B: userID1},
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr1 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID1),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID1, expectedErr1)

		expectedErr2 := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID2),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID2, expectedErr2)
	})

	t.Run("OwnUIDSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)
		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		user1UID := sha256Binary(user1 + "@" + dbName)
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{B: user1UID}),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("OtherUserUIDSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, user2Conn)
		cursorID := find(t, ctx, user2Conn, dbName, cName, sessionID)

		user2ID := sha256Binary(user2 + "@" + dbName)
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{B: user2ID}),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user2Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("DoesNotKillOtherUIDSessions", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)
		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		user2ID := sha256Binary(user2 + "@" + dbName)
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{B: user2ID}),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("NonExistentUIDSessions", func(t *testing.T) {
		nonExistentUID := sha256Binary("nonexistent@foo")

		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{B: nonExistentUID}),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)
	})

	t.Run("BothLSIDAndUIDPatterns", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)
		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		user1NotKillSessionID := startSession(t, ctx, user1Conn)
		user1NotKillCursorID := find(t, ctx, user1Conn, dbName, cName, user1NotKillSessionID)

		user2SessionID := startSession(t, ctx, user2Conn)
		user2CursorID := find(t, ctx, user2Conn, dbName, cName, user1SessionID)

		user1UID := sha256Binary(user1 + "@" + dbName)
		user2UID := sha256Binary(user2 + "@" + dbName)

		pattern := wirebson.MustArray(
			wirebson.MustDocument("lsid", wirebson.MustDocument(
				"id", user1SessionID,
				"uid", wirebson.Binary{B: user1UID},
			)),
			wirebson.MustDocument("uid", wirebson.Binary{B: user2UID}),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, nil)

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

		getMore(t, ctx, user1Conn, dbName, cName, user1NotKillSessionID, user1NotKillCursorID, nil)
	})
}

func TestKillAllSessionsByPatternAllUsers(t *testing.T) {
	// do not run this test in parallel as it kills sessions of other tests (except for in-process FerretDB)
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

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("EmptyPattern", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)

		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		pattern := wirebson.MustArray()

		killAllSessionsByPattern(t, ctx, user2Conn, dbName, pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("EmptyUsers", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)
		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		emptyUsersPattern := wirebson.MustArray(
			wirebson.MustDocument("users", wirebson.MustArray()),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, emptyUsersPattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)
		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		nonExistentUserPattern := wirebson.MustArray(
			wirebson.MustDocument("users", wirebson.MustArray(
				wirebson.MustDocument("db", "foo", "user", "nonexistent"),
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, nonExistentUserPattern, nil)

		// user1 does not match the pattern, but the cursor of user1 is deleted unexpectedly
		// and FerretDB exhibits the same behavior for the compatibility
		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, expectedErr)
	})

	t.Run("PatternMatchedUser", func(t *testing.T) {
		sessionID := startSession(t, ctx, user1Conn)
		cursorID := find(t, ctx, user1Conn, dbName, cName, sessionID)

		user1Pattern := wirebson.MustArray(
			wirebson.MustDocument("users", wirebson.MustArray(
				wirebson.MustDocument("db", dbName, "user", user1),
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, user1Pattern, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("PatternNotMatchedUser", func(t *testing.T) {
		user1SessionID := startSession(t, ctx, user1Conn)
		user1CursorID := find(t, ctx, user1Conn, dbName, cName, user1SessionID)

		user2Pattern := wirebson.MustArray(
			wirebson.MustDocument("users", wirebson.MustArray(
				wirebson.MustDocument("db", dbName, "user", user2),
			)),
		)

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, user2Pattern, nil)

		// sessions of user2 was killed, not user1, but the cursor of user1 is deleted unexpectedly
		// and FerretDB exhibits the same behavior for the compatibility
		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", user1CursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))
		getMore(t, ctx, user1Conn, dbName, cName, user1SessionID, user1CursorID, expectedErr)
	})
}

func TestKillAllSessionsByPatternErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	dbName := db.Name()

	user1 := t.Name()
	user1Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user1)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("NonArrayPatterns", func(t *testing.T) {
		pattern := "invalid"

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern' is the wrong type 'string', expected type 'array'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("StringPattern", func(t *testing.T) {
		pattern := wirebson.MustArray("invalid")

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.0' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("UnknownPattern", func(tt *testing.T) {
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1091")

		pattern := wirebson.MustArray(wirebson.MustDocument("invalid", "foo"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.invalid' is an unknown field.",
			"code", int32(40415),
			"codeName", "Location40415",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidString", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", "invalid"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidMissingID", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument()))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.id' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidStringID", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument("id", "invalid")))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.id' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidFunctionBinarySubtype", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{Subtype: wirebson.BinaryFunction},
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.id' is the wrong binData type 'function', expected type 'UUID'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidEmptyBinaryUUID", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{Subtype: wirebson.BinaryUUID},
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "uuid must be a 16-byte binary field with UUID (4) subtype",
			"code", int32(207),
			"codeName", "InvalidUUID",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidMissingUID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{B: randomUUID[:], Subtype: wirebson.BinaryUUID},
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.uid' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidStringUID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{B: randomUUID[:], Subtype: wirebson.BinaryUUID},
			"uid", "invalid",
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.uid' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidEmptyUID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{B: randomUUID[:], Subtype: wirebson.BinaryUUID},
			"uid", wirebson.Binary{},
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "Unsupported SHA256Block hash length: 0",
			"code", int32(12),
			"codeName", "UnsupportedFormat",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("LsidFunctionUID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		pattern := wirebson.MustArray(wirebson.MustDocument("lsid", wirebson.MustDocument(
			"id", wirebson.Binary{B: randomUUID[:], Subtype: wirebson.BinaryUUID},
			"uid", wirebson.Binary{Subtype: wirebson.BinaryFunction},
		)))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.lsid.uid' is the wrong binData type 'function', expected type 'general'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})
}

func TestKillAllSessionsByPatternUIDErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	dbName := db.Name()

	user1 := t.Name()
	user1Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user1)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("UIDString", func(t *testing.T) {
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", "invalid"),
		)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.uid' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("UIDEmpty", func(t *testing.T) {
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{}),
		)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "Unsupported SHA256Block hash length: 0",
			"code", int32(12),
			"codeName", "UnsupportedFormat",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("UIDFunctionType", func(t *testing.T) {
		pattern := wirebson.MustArray(
			wirebson.MustDocument("uid", wirebson.Binary{Subtype: wirebson.BinaryFunction}),
		)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.uid' is the wrong binData type 'function', expected type 'general'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})
}

func TestKillAllSessionsByPatternUsersErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	// username:password user is not used to avoid killing all sessions of that user

	ctx := s.Ctx

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	dbName := db.Name()

	user1 := t.Name()
	user1Conn := createKillSessionUser(t, ctx, db, s.MongoDBURI, user1)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("NotArrayUsers", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument("users", "invalid"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users' is the wrong type 'string', expected type 'array'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("NotDocumentUser", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument(
			"users", wirebson.MustArray("invalid"),
		))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users.0' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("MissingDB", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument(
			"users", wirebson.MustArray(wirebson.MustDocument("user", "foo")),
		))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users.db' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("BoolDB", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument(
			"users", wirebson.MustArray(wirebson.MustDocument("db", true, "user", "foo")),
		))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users.db' is the wrong type 'bool', expected type 'string'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("MissingUser", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument(
			"users", wirebson.MustArray(wirebson.MustDocument("db", "bar")),
		))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users.user' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})

	t.Run("BoolUser", func(t *testing.T) {
		pattern := wirebson.MustArray(wirebson.MustDocument(
			"users", wirebson.MustArray(wirebson.MustDocument("db", "bar", "user", false)),
		))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'KillAllSessionsByPatternCmd.killAllSessionsByPattern.users.user' is the wrong type 'bool', expected type 'string'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killAllSessionsByPattern(t, ctx, user1Conn, dbName, pattern, expectedErr)
	})
}

// createKillSessionUser creates a user with privileges to kill all sessions.
func createKillSessionUser(t *testing.T, ctx context.Context, db *mongo.Database, mongoDBURI, username string) *wireclient.Conn {
	roles := bson.A{}

	// TODO https://github.com/FerretDB/FerretDB/issues/3974
	if setup.IsMongoDB(t) {
		impersonateRole := username + "Role"
		roles = bson.A{"root", impersonateRole}

		t.Cleanup(func() {
			_ = db.RunCommand(ctx, bson.D{{"dropRole", impersonateRole}})
		})

		err := db.RunCommand(ctx, bson.D{
			{"createRole", impersonateRole},
			{"privileges", bson.A{
				bson.D{
					{"resource", bson.D{{"cluster", true}}},
					{"actions", bson.A{"impersonate"}},
				},
			}},
			{"roles", bson.A{}},
		}).Err()
		require.NoError(t, err)
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", username}})

	password := username + "pass"

	err := db.RunCommand(ctx, bson.D{
		{"createUser", username},
		{"roles", roles},
		{"pwd", password},
	}).Err()
	require.NoError(t, err)

	conn, err := wireclient.Connect(ctx, mongoDBURI, testutil.Logger(t))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, conn.Close())
	})

	err = conn.Login(ctx, username, password, db.Name())
	require.NoError(t, err)

	return conn
}

// killAllSessionsByPattern sends a request to kill all sessions that matches the pattern.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
func killAllSessionsByPattern(t testing.TB, ctx context.Context, conn *wireclient.Conn, dbName string, pattern any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"killAllSessionsByPattern", pattern,
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

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
	"crypto/sha256"
	"encoding/base64"
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

func TestSessionConnection(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	ctx, collection, db, conn1 := s.Ctx, s.Collection, s.Collection.Database(), s.WireConn
	cName, dbName := collection.Name(), db.Name()

	conn2, err := wireclient.Connect(ctx, s.MongoDBURI, testutil.Logger(t))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, conn2.Close())
	})

	err = conn2.Login(ctx, "username", "password", "admin")
	require.NoError(t, err)

	_, err = collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "a"}},
		bson.D{{"_id", "b"}},
	})
	require.NoError(t, err)

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("SameSessionID", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("DifferentSessionID", func(t *testing.T) {
		sessionID1 := startSession(t, ctx, conn1)
		sessionID2 := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID1)

		userHash := sha256Base64("username@admin")

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage(userHash, userHash, sessionID1, sessionID2),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, conn1, dbName, cName, sessionID2, cursorID, expectedErr)
	})

	t.Run("ConcurrentSessions", func(t *testing.T) {
		sessionID1 := startSession(t, ctx, conn1)
		sessionID2 := startSession(t, ctx, conn1)

		cursorID1 := find(t, ctx, conn1, dbName, cName, sessionID1)
		cursorID2 := find(t, ctx, conn1, dbName, cName, sessionID2)

		getMore(t, ctx, conn1, dbName, cName, sessionID1, cursorID1, nil)
		getMore(t, ctx, conn1, dbName, cName, sessionID2, cursorID2, nil)
	})

	t.Run("DifferentConnection", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		getMore(t, ctx, conn2, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("NonExistentSessionID", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		// sessionID is random UUID instead of explicitly calling startSession command
		sessionID := wirebson.Binary{
			B:       randomUUID[:],
			Subtype: 0x04,
		}

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("CloseConnection", func(t *testing.T) {
		var anotherConn *wireclient.Conn
		anotherConn, err = wireclient.Connect(ctx, s.MongoDBURI, testutil.Logger(t))
		require.NoError(t, err)

		err = anotherConn.Login(ctx, "username", "password", "admin")
		require.NoError(t, err)

		sessionID := startSession(t, ctx, anotherConn)

		cursorID := find(t, ctx, anotherConn, dbName, cName, sessionID)

		err = anotherConn.Close()
		require.NoError(t, err)

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("NoLsid", func(t *testing.T) {
		cursorID := find(t, ctx, conn1, dbName, cName, wirebson.Binary{})

		getMore(t, ctx, conn1, dbName, cName, wirebson.Binary{}, cursorID, nil)
	})

	t.Run("FindNoLsid", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, wirebson.Binary{})

		userHash := sha256Base64("username@admin")

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage(userHash, userHash, wirebson.Binary{}, sessionID),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("GetMoreNoLsid", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		userHash := sha256Base64("username@admin")

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage(userHash, userHash, sessionID, wirebson.Binary{}),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, conn1, dbName, cName, wirebson.Binary{}, cursorID, expectedErr)
	})

	t.Run("KilledCursor", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		killCursors(t, ctx, conn1, dbName, cName, cursorID, sessionID, nil)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d not found", cursorID),
			"code", int32(43),
			"codeName", "CursorNotFound",
		))

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("KillCursorInvalidSessionID", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn1)

		cursorID := find(t, ctx, conn1, dbName, cName, sessionID)

		invalidSessionID := "invalid"

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'OperationSessionInfo.lsid.id' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		killCursors(t, ctx, conn1, dbName, cName, cursorID, invalidSessionID, expectedErr)

		getMore(t, ctx, conn1, dbName, cName, sessionID, cursorID, nil)
	})
}

func TestFindLsidErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	ctx, collection, db, conn := s.Ctx, s.Collection, s.Collection.Database(), s.WireConn
	cName, dbName := collection.Name(), db.Name()

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("StringLsid", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"find", cName,
			"lsid", "invalid",
			"$db", dbName,
		)

		_, resBody, err := conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'OperationSessionInfo.lsid' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		)

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("LsidMissingID", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"find", cName,
			"lsid", wirebson.MustDocument(),
			"$db", dbName,
		)

		_, resBody, err := conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'OperationSessionInfo.lsid.id' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		)

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("LsidStringID", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"find", cName,
			"lsid", wirebson.MustDocument("id", "invalid"),
			"$db", dbName,
		)

		_, resBody, err := conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'OperationSessionInfo.lsid.id' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		)

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("LsidFunctionBinarySubtype", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"find", cName,
			"lsid", wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryFunction}),
			"$db", dbName,
		)

		_, resBody, err := conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'OperationSessionInfo.lsid.id' is the wrong binData type 'function', expected type 'UUID'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		)

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("LsidEmptyBinaryUUID", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"find", cName,
			"lsid", wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryUUID}),
			"$db", dbName,
		)

		_, resBody, err := conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", "uuid must be a 16-byte binary field with UUID (4) subtype",
			"code", int32(207),
			"codeName", "InvalidUUID",
		)

		testutil.AssertEqual(t, expected, res)
	})
}

func TestSessionConnectionDifferentUser(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})
	ctx, adminConn := s.Ctx, s.WireConn

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	db := s.Collection.Database().Client().Database("admin")
	collection := db.Collection(s.Collection.Name())
	cName, dbName := collection.Name(), db.Name()

	roles := bson.A{"readWrite"}
	if !setup.IsMongoDB(t) {
		// TODO https://github.com/FerretDB/FerretDB/issues/3974
		roles = bson.A{}
	}

	user, pass := "testsessionuser", "sessionpassword"

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", user}})

	err := db.RunCommand(ctx, bson.D{
		{"createUser", user},
		{"roles", roles},
		{"pwd", pass},
	}).Err()
	require.NoError(t, err)

	userConn, err := wireclient.Connect(ctx, s.MongoDBURI, testutil.Logger(t))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, userConn.Close())
	})

	err = userConn.Login(ctx, user, pass, dbName)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	_, err = collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "a"}},
		bson.D{{"_id", "b"}},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, collection.Drop(ctx))
	})

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("DiffUserSameSessionID", func(t *testing.T) {
		sessionID := startSession(t, ctx, adminConn)

		cursorID := find(t, ctx, adminConn, dbName, cName, sessionID)

		user1Hash := sha256Base64("username@admin")
		user2Hash := sha256Base64(user + "@" + dbName)

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage(user1Hash, user2Hash, sessionID, sessionID),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, userConn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("DiffUserStartSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, adminConn)

		cursorID := find(t, ctx, userConn, dbName, cName, sessionID)

		getMore(t, ctx, userConn, dbName, cName, sessionID, cursorID, nil)
	})

	t.Run("NoLsid", func(t *testing.T) {
		cursorID := find(t, ctx, adminConn, dbName, cName, wirebson.Binary{})

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", fmt.Sprintf("cursor id %d was not created by the authenticated user", cursorID),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, userConn, dbName, cName, wirebson.Binary{}, cursorID, expectedErr)
	})

	t.Run("FindNoLsid", func(t *testing.T) {
		sessionID := startSession(t, ctx, adminConn)

		cursorID := find(t, ctx, adminConn, dbName, cName, wirebson.Binary{})

		user2Hash := sha256Base64(user + "@" + dbName)

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage("", user2Hash, wirebson.Binary{}, sessionID),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, userConn, dbName, cName, sessionID, cursorID, expectedErr)
	})

	t.Run("GetMoreNoLsid", func(t *testing.T) {
		sessionID := startSession(t, ctx, adminConn)

		cursorID := find(t, ctx, adminConn, dbName, cName, sessionID)

		user1Hash := sha256Base64("username@admin")

		expectedErr := wirebson.MustDocument(
			"ok", float64(0),
			"errmsg", sessionErrorMessage(user1Hash, "", sessionID, wirebson.Binary{}),
			"code", int32(13),
			"codeName", "Unauthorized",
		)

		getMore(t, ctx, userConn, dbName, cName, wirebson.Binary{}, cursorID, expectedErr)
	})

	t.Run("KillDiffUserCursor", func(t *testing.T) {
		sessionID := startSession(t, ctx, adminConn)

		cursorID := find(t, ctx, adminConn, dbName, cName, sessionID)

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			// errmsg field is not compared, because as it is difficult produce exact format of document as below
			// `not authorized on admin to execute command{ killCursors: "test", cursors: [ 8541858944752455730 ],
			// lsid: { id: UUID("363dad0b-d9b8-406f-9575-b11a3779faa0") }, $db: "admin" }`
			"code", int32(13),
			"codeName", "Unauthorized",
		))

		killCursors(t, ctx, userConn, dbName, cName, cursorID, sessionID, expectedErr)

		getMore(t, ctx, adminConn, dbName, cName, sessionID, cursorID, nil)
	})
}

// startSession sends a request and returns a sessionID.
func startSession(t testing.TB, ctx context.Context, conn *wireclient.Conn) wirebson.Binary {
	msg := wire.MustOpMsg(
		"startSession", int32(1),
		"$db", "admin", // startSession is always sent to the admin database
	)

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	var res *wirebson.Document
	res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	sessionIDDoc := res.Get("id")
	require.NotNil(t, sessionIDDoc, wirebson.LogMessage(res))

	sessionID := sessionIDDoc.(*wirebson.Document).Get("id").(wirebson.Binary)

	expected := wirebson.MustDocument(
		"id", wirebson.MustDocument("id", sessionID),
		"timeoutMinutes", int32(30),
		"ok", float64(1),
	)

	testutil.AssertEqual(t, expected, res)

	return sessionID
}

// killCursors sends a request to kill the given cursor.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
// If expectedErr does not have `errmsg` field set, it compares error code only.
func killCursors(t testing.TB, ctx context.Context, conn *wireclient.Conn, dbName, cName string, cursorID, sessionID any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"killCursors", cName,
		"cursors", wirebson.MustArray(cursorID),
		"$db", dbName,
	)

	if sessionID != nil {
		msg = wire.MustOpMsg(
			"killCursors", cName,
			"cursors", wirebson.MustArray(cursorID),
			"lsid", wirebson.MustDocument("id", sessionID),
			"$db", dbName,
		)
	}

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	if expectedErr != nil {
		if expectedErr.Get("errmsg") == nil {
			require.NotNil(t, res.Get("errmsg"))
			res.Remove("errmsg")
		}

		testutil.AssertEqual(t, expectedErr, res)

		return
	}

	expected := must.NotFail(wirebson.NewDocument(
		"cursorsKilled", wirebson.MustArray(cursorID),
		"cursorsNotFound", wirebson.MakeArray(0),
		"cursorsAlive", wirebson.MakeArray(0),
		"cursorsUnknown", wirebson.MakeArray(0),
		"ok", float64(1),
	))

	testutil.AssertEqual(t, expected, res)
}

// find sends a request with a batch size of 1 and returns cursorID.
// When non-empty sessionID is provided, `lsid` field is set.
// It checks the first batch contains a document {_id: 'a'}.
func find(t testing.TB, ctx context.Context, conn *wireclient.Conn, db, coll string, sessionID wirebson.Binary) any {
	msg := wire.MustOpMsg(
		"find", coll,
		"batchSize", int32(1),
		"$db", db,
	)

	if sessionID.B != nil {
		msg = wire.MustOpMsg(
			"find", coll,
			"batchSize", int32(1),
			"lsid", wirebson.MustDocument("id", sessionID),
			"$db", db,
		)
	}

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	cursor := res.Get("cursor")
	require.NotNil(t, cursor, wirebson.LogMessage(res))

	cursorID := cursor.(*wirebson.Document).Get("id")
	require.NotZero(t, cursorID)

	expected := wirebson.MustDocument(
		"cursor", wirebson.MustDocument(
			"firstBatch", wirebson.MustArray(
				wirebson.MustDocument("_id", "a"),
			),
			"id", cursorID,
			"ns", db+"."+coll,
		),
		"ok", float64(1),
	)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/810
	if !setup.IsMongoDB(t) {
		expected = wirebson.MustDocument(
			"cursor", wirebson.MustDocument(
				"id", cursorID,
				"ns", db+"."+coll,
				"firstBatch", wirebson.MustArray(
					wirebson.MustDocument("_id", "a"),
				),
			),
			"ok", float64(1),
		)
	}

	testutil.AssertEqual(t, expected, res)

	return cursorID
}

// getMore sends a request and checks the next batch contains a document {_id: 'b'}
// When non-empty sessionID is provided, `lsid` field is set.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
func getMore(t testing.TB, ctx context.Context, conn *wireclient.Conn, db, coll string, sessionID wirebson.Binary, cursorID any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"getMore", cursorID,
		"collection", coll,
		"$db", db,
	)

	if sessionID.B != nil {
		msg = wire.MustOpMsg(
			"getMore", cursorID,
			"collection", coll,
			"lsid", wirebson.MustDocument("id", sessionID),
			"$db", db,
		)
	}

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	if expectedErr != nil {
		testutil.AssertEqual(t, expectedErr, res)

		return
	}

	expected := wirebson.MustDocument(
		"cursor", wirebson.MustDocument(
			"nextBatch", wirebson.MustArray(
				wirebson.MustDocument("_id", "b"),
			),
			"id", int64(0),
			"ns", db+"."+coll,
		),
		"ok", float64(1),
	)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/810
	if !setup.IsMongoDB(t) {
		expected = wirebson.MustDocument(
			"cursor", wirebson.MustDocument(
				"id", int64(0),
				"ns", db+"."+coll,
				"nextBatch", wirebson.MustArray(
					wirebson.MustDocument("_id", "b"),
				),
			),
			"ok", float64(1),
		)
	}

	testutil.AssertEqual(t, expected, res)
}

// sessionErrorMessage returns the expected error message from the given users' hash and
// sessionIDs used for accessing the cursor.
func sessionErrorMessage(findUserHash, getMoreUserHash string, findSessionID, getMoreSessionID wirebson.Binary) string {
	findCursorID := "none"

	if len(findSessionID.B) > 0 {
		findUUID := must.NotFail(uuid.FromBytes(findSessionID.B)).String()
		findCursorID = fmt.Sprintf("%s - %s -  - ", findUUID, findUserHash)
	}

	getMoreCursorID := "none"

	if len(getMoreSessionID.B) > 0 {
		getMoreUUID := must.NotFail(uuid.FromBytes(getMoreSessionID.B)).String()
		getMoreCursorID = fmt.Sprintf("%s - %s -  - ", getMoreUUID, getMoreUserHash)
	}

	msgBase := "Cursor session id (%s) is not the same as the operation context's session id (%s)"

	return fmt.Sprintf(msgBase, findCursorID, getMoreCursorID)
}

// sha256Binary applies SHA-256 to the input string and returns bytes.
func sha256Binary(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))

	return h.Sum(nil)
}

// sha256Base64 applies SHA-256 to the input string and returns the base64 encoded hash.
func sha256Base64(s string) string {
	return base64.StdEncoding.EncodeToString(sha256Binary(s))
}

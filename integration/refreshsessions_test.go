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
	"testing"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wireclient"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestRefreshSessions(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	ctx, db, conn := s.Ctx, s.Collection, s.WireConn
	dbName := db.Name()

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("OneSession", func(t *testing.T) {
		sessionID := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, nil)
	})

	t.Run("NonExistentSession", func(t *testing.T) {
		randomUUID := must.NotFail(uuid.NewRandom())

		sessionID := wirebson.Binary{
			B:       randomUUID[:],
			Subtype: 0x04,
		}

		sessions := wirebson.MustArray(wirebson.MustDocument("id", sessionID))

		refreshSessions(t, ctx, conn, dbName, sessionID, sessions, nil)
	})

	t.Run("MultipleSessions", func(t *testing.T) {
		sessionID1 := startSession(t, ctx, conn)
		sessionID2 := startSession(t, ctx, conn)

		sessions := wirebson.MustArray(
			wirebson.MustDocument("id", sessionID1),
			wirebson.MustDocument("id", sessionID2),
		)

		refreshSessions(t, ctx, conn, dbName, sessionID1, sessions, nil)
	})
}

func TestRefreshSessionsErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	ctx, db, conn := s.Ctx, s.Collection, s.WireConn
	dbName := db.Name()

	// test cases are not run in parallel as they use the same conn and would cause datarace

	t.Run("NotArray", func(t *testing.T) {
		sessions := "invalid"
		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'refreshSessions.refreshSessions' is the wrong type 'string', expected type 'array'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("NotSessionDocument", func(t *testing.T) {
		sessions := wirebson.MustArray("invalid")

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'refreshSessions.refreshSessionsFromClient.0' is the wrong type 'string', expected type 'object'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("MissingID", func(t *testing.T) {
		sessions := wirebson.MustArray(wirebson.MustDocument())

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'refreshSessions.refreshSessionsFromClient.id' is missing but a required field",
			"code", int32(40414),
			"codeName", "Location40414",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("WrongIDType", func(t *testing.T) {
		sessions := wirebson.MustArray(wirebson.MustDocument("id", "invalid"))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'refreshSessions.refreshSessionsFromClient.id' is the wrong type 'string', expected type 'binData'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("WrongIDSubtype", func(t *testing.T) {
		sessions := wirebson.MustArray(wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryFunction}))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "BSON field 'refreshSessions.refreshSessionsFromClient.id' is the wrong binData type 'function', expected type 'UUID'",
			"code", int32(14),
			"codeName", "TypeMismatch",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("NotUUID", func(t *testing.T) {
		sessions := wirebson.MustArray(wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryUUID}))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "uuid must be a 16-byte binary field with UUID (4) subtype",
			"code", int32(207),
			"codeName", "InvalidUUID",
		))

		refreshSessions(t, ctx, conn, dbName, nil, sessions, expectedErr)
	})

	t.Run("LsidEmptyBinaryUUID", func(t *testing.T) {
		lsid := wirebson.Binary{Subtype: wirebson.BinaryUUID}
		sessions := wirebson.MustArray(wirebson.MustDocument("id", wirebson.Binary{Subtype: wirebson.BinaryUUID}))

		expectedErr := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "uuid must be a 16-byte binary field with UUID (4) subtype",
			"code", int32(207),
			"codeName", "InvalidUUID",
		))

		refreshSessions(t, ctx, conn, dbName, lsid, sessions, expectedErr)
	})
}

// refreshSessions sends a request with given sessions.
// If expectedErr is not nil, the error is checked, otherwise it checks the response.
func refreshSessions(t testing.TB, ctx context.Context, conn *wireclient.Conn, db string, lsid, sessions any, expectedErr *wirebson.Document) {
	msg := wire.MustOpMsg(
		"refreshSessions", sessions,
		"$db", db,
	)

	if lsid != nil {
		msg = wire.MustOpMsg(
			"refreshSessions", sessions,
			"lsid", wirebson.MustDocument("id", lsid),
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

	expected := must.NotFail(wirebson.NewDocument(
		"ok", float64(1),
	))

	testutil.AssertEqual(t, expected, res)
}

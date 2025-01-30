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

package handler

import (
	"context"
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgKillAllSessionsByPattern implements `killAllSessionsByPattern` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgKillAllSessionsByPattern(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec)
	if err != nil {
		return nil, err
	}

	command := doc.Command()

	v := doc.Get(command)
	field := "KillAllSessionsByPatternCmd.killAllSessionsByPattern"

	patternV, ok := v.(wirebson.AnyArray)
	if !ok {
		m := fmt.Sprintf("BSON field '%s' is the wrong type '%T', expected type 'array'", field, v)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, command)
	}

	patternArr, err := patternV.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var allSessions bool
	var userIDs []session.UserID
	lsids := map[session.UserID][]uuid.UUID{}

	if patternArr.Len() == 0 {
		allSessions = true
	}

	for v = range patternArr.Values() {
		var patternDoc wirebson.AnyDocument

		if patternDoc, ok = v.(wirebson.AnyDocument); !ok {
			m := fmt.Sprintf("BSON field '%s.0' is the wrong type '%T', expected type 'object'", field, v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, command)
		}

		var pattern *wirebson.Document

		if pattern, err = patternDoc.Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		for k, v := range pattern.All() {
			switch k {
			case "lsid":
				var userID session.UserID
				var sessionID uuid.UUID

				if userID, sessionID, err = getLSIDParam(v, command, field); err != nil {
					return nil, err
				}

				lsids[userID] = append(lsids[userID], sessionID)

			case "uid":
				var userID session.UserID

				if userID, err = getUserIDParam(v, command, field); err != nil {
					return nil, err
				}

				userIDs = append(userIDs, userID)

			case "users":
				if _, err = getSessionUsersParam(v, command, fmt.Sprintf("%s.users", field)); err != nil {
					return nil, err
				}

				// for compatibility, all sessions of all users are deleted regardless of the pattern
				allSessions = true

			default:
				// delete sessions by roles pattern
				// TODO https://github.com/FerretDB/FerretDB/issues/3974
				msg := fmt.Sprintf("BSON field '%s.%s' is an unknown field.", field, k)
				return nil, mongoerrors.NewWithArgument(mongoerrors.ErrUnknownBsonField, msg, command)
			}
		}
	}

	var allCursorIDs []int64

	if allSessions {
		cursorIDs := h.s.DeleteAllSessions()
		allCursorIDs = append(allCursorIDs, cursorIDs...)
	}

	if len(userIDs) > 0 {
		cursorIDs := h.s.DeleteSessionsByUserIDs(userIDs)
		allCursorIDs = append(allCursorIDs, cursorIDs...)
	}

	for userID, sessionIDs := range lsids {
		cursorIDs := h.s.DeleteSessionsByIDs(userID, sessionIDs)
		allCursorIDs = append(allCursorIDs, cursorIDs...)
	}

	for _, cursorID := range allCursorIDs {
		_ = h.Pool.KillCursor(connCtx, cursorID)
	}

	return wire.MustOpMsg(
		"ok", float64(1),
	), nil
}

// getLSIDParam returns user ID and session ID from the given `v`.
// The `v` has the format `{id: <uuid>, uid: <binary>}` and
// a protocol error is returned for invalid format or value.
func getLSIDParam(v any, command, field string) (session.UserID, uuid.UUID, error) {
	lsidV, ok := v.(wirebson.AnyDocument)
	if !ok {
		msg := fmt.Sprintf("BSON field '%s.lsid' is the wrong type '%T', expected type 'object'", field, v)
		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	lsid, err := lsidV.Decode()
	if err != nil {
		return session.UserID{}, uuid.Nil, lazyerrors.Error(err)
	}

	v = lsid.Get("id")
	if v == nil {
		msg := fmt.Sprintf("BSON field '%s.lsid.id' is missing but a required field", field)
		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, msg, command)
	}

	binaryID, ok := v.(wirebson.Binary)
	if !ok {
		msg := fmt.Sprintf("BSON field '%s.lsid.id' is the wrong type '%T', expected type 'binData'", field, v)
		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	if binaryID.Subtype != wirebson.BinaryUUID {
		msg := fmt.Sprintf(
			"BSON field '%s.lsid.id' is the wrong binData type '%s', expected type 'UUID'",
			field,
			binaryID.Subtype.String(),
		)

		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	sessionID, err := uuid.FromBytes(binaryID.B)
	if err != nil {
		msg := "uuid must be a 16-byte binary field with UUID (4) subtype"
		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidUUID, msg, command)
	}

	v = lsid.Get("uid")
	if v == nil {
		msg := fmt.Sprintf("BSON field '%s.lsid.uid' is missing but a required field", field)
		return session.UserID{}, uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, msg, command)
	}

	userID, err := getUserIDParam(v, command, fmt.Sprintf("%s.lsid", field))
	if err != nil {
		return session.UserID{}, uuid.Nil, err
	}

	return userID, sessionID, nil
}

// getUserIDParam parses binary from `v` and returns user ID.
// A protocol error is returned for invalid format or value.
func getUserIDParam(v any, command, field string) (session.UserID, error) {
	binaryUserID, ok := v.(wirebson.Binary)
	if !ok {
		msg := fmt.Sprintf("BSON field '%s.uid' is the wrong type '%T', expected type 'binData'", field, v)
		return session.UserID{}, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	if binaryUserID.Subtype != wirebson.BinaryGeneric {
		msg := fmt.Sprintf(
			"BSON field '%s.uid' is the wrong binData type '%s', expected type 'general'",
			field,
			binaryUserID.Subtype.String(),
		)

		return session.UserID{}, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	var userID session.UserID

	if len(binaryUserID.B) != len(userID) {
		msg := fmt.Sprintf("Unsupported SHA256Block hash length: %d", len(binaryUserID.B))
		return session.UserID{}, mongoerrors.NewWithArgument(mongoerrors.ErrUnsupportedFormat, msg, command)
	}

	copy(userID[:], binaryUserID.B)

	return userID, nil
}

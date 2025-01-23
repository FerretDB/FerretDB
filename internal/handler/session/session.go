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

// Package session provides access to session registry.
package session

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// LogicalSessionTimeoutMinutes is the session timeout in minutes.
const LogicalSessionTimeoutMinutes = int32(30)

// UserID is the output computed by SHA256 function.
type UserID [sha256.Size]byte

// String returns the base64 encoded string.
func (s UserID) String() string {
	return base64.StdEncoding.EncodeToString(s[:])
}

// sessionInfo contains information of a session.
type sessionInfo struct {
	cursorIDs map[int64]struct{}
	created   time.Time
	lastUsed  time.Time
	ended     bool

	token *resource.Token
}

// newSession returns a new session information.
func newSessionInfo() *sessionInfo {
	now := time.Now()

	s := &sessionInfo{
		created:  now,
		lastUsed: now,
		token:    resource.NewToken(),
	}

	resource.Track(s, s.token)

	return s
}

// close untracks the session information.
func (s *sessionInfo) close() {
	s.cursorIDs = nil
	resource.Untrack(s, s.token)
}

// getSessionUUID extracts the session ID from `lsid`.
// If `lsid` field does not exist, it returns an empty uuid.
func getSessionUUID(spec wirebson.RawDocument) (uuid.UUID, error) {
	doc, err := spec.Decode()
	if err != nil {
		return uuid.Nil, lazyerrors.Error(err)
	}

	v := doc.Get("lsid")
	if v == nil {
		return uuid.Nil, nil
	}

	lsidV, ok := v.(wirebson.AnyDocument)
	if !ok {
		msg := fmt.Sprintf("BSON field 'OperationSessionInfo.lsid' is the wrong type '%T', expected type 'object'", v)
		return uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, "lsid")
	}

	lsid, err := lsidV.Decode()
	if err != nil {
		return uuid.Nil, lazyerrors.Error(err)
	}

	v = lsid.Get("id")
	if v == nil {
		msg := "BSON field 'OperationSessionInfo.lsid.id' is missing but a required field"
		return uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, msg, "lsid")
	}

	binaryID, ok := v.(wirebson.Binary)
	if !ok {
		msg := fmt.Sprintf("BSON field 'OperationSessionInfo.lsid.id' is the wrong type '%T', expected type 'binData'", v)
		return uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, "lsid")
	}

	if binaryID.Subtype != wirebson.BinaryUUID {
		msg := fmt.Sprintf(
			"BSON field 'OperationSessionInfo.lsid.id' is the wrong binData type '%s', expected type 'UUID'",
			binaryID.Subtype.String(),
		)

		return uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, "lsid")
	}

	sessionID, err := uuid.FromBytes(binaryID.B)
	if err != nil {
		msg := "uuid must be a 16-byte binary field with UUID (4) subtype"
		return uuid.Nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidUUID, msg, "lsid")
	}

	return sessionID, nil
}

// getUserID gets the username from conninfo and returns the hash of <username>@<database>.
// If there is no logged-in user, it returns a hash of an empty string.
func getUserID(ctx context.Context) UserID {
	var username, db string

	if conv := conninfo.Get(ctx).Conv(); conv != nil {
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
		username = conv.Username()
		db = "admin"
	}

	return GetUIDFromUsername(db, username)
}

// GetUIDFromUsername returns the hash of <username>@<database>.
// If the username is empty, it returns a hash of an empty string.
func GetUIDFromUsername(db, username string) UserID {
	var userAtDB string

	if username != "" {
		userAtDB = username + "@" + db
	}

	return sha256.Sum256([]byte(userAtDB))
}

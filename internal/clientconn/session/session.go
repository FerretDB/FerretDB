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
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/types"
)

// Session represents a session.
type Session struct {
	id       types.Binary
	user     string
	database string
	lastUsed time.Time
	expired  bool
}

// newSession returns a new session.
func newSession(user, db string, id uuid.UUID) *Session {
	sessionID := types.Binary{Subtype: types.BinaryUUID, B: id[:]}
	return &Session{
		id:       sessionID,
		user:     user,
		database: db,
		lastUsed: time.Now(),
	}
}

// hash returns a sha256 hash of user and database.
func hash(user, db string) string {
	str := user

	if db != "" {
		str += "@" + db
	}

	hash := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", hash)
}

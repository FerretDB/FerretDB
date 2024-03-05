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

package session

import (
	"context"
	"sync"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// timeout is the session timeout.
const timeout = 30 * time.Minute

type Registry struct {
	rw sync.RWMutex
	m  map[string]map[string]*Session

	l *zap.Logger
}

func NewRegistry(l *zap.Logger) *Registry {
	return &Registry{
		m: make(map[string]map[string]*Session),
		l: l,
	}
}

type NewParams struct {
	DB       string
	Username string
}

func (r *Registry) NewSession(ctx context.Context, params *NewParams) *Session {
	r.rw.Lock()
	defer r.rw.Unlock()

	id := uuid.New()
	sessionID := types.Binary{Subtype: types.BinaryUUID, B: id[:]}
	session := newSession(sessionID, params.DB)

	r.m[params.Username][id.String()] = session

	return session
}

func (r *Registry) GetSession(ctx context.Context, username, db string, sessionID types.Binary) *Session {
	r.rw.RLock()
	defer r.rw.RUnlock()

	id := string(sessionID.B)

	if _, ok := r.m[username][id]; !ok {
		r.m[username][id] = newSession(sessionID, db)
	}

	r.m[username][id].lastUsed = time.Now()

	return r.m[username][id]
}

func (r *Registry) KillSession(ctx context.Context, username, sessionID string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	delete(r.m[username], sessionID)
}

func (r *Registry) EndSession(ctx context.Context, username, sessionID string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.m[username][sessionID].expired = true
}

func (r *Registry) CleanExpired(ctx context.Context) {
	r.rw.Lock()
	defer r.rw.Unlock()

	for _, sessions := range r.m {
		for id, session := range sessions {
			if session.expired || time.Since(session.lastUsed) > timeout {
				delete(sessions, id)
			}
		}
	}
}

func newSession(id types.Binary, db string) *Session {
	return &Session{
		id:       id,
		db:       db,
		lastUsed: time.Now(),
	}
}

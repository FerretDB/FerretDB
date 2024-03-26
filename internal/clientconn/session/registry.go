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
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// timeout is the session timeout.
const timeout = 30 * time.Minute

// Registry stores sessions.
type Registry struct {
	rw sync.RWMutex
	m  map[string]map[string]*Session // hash -> sessionID -> session

	l *zap.Logger
}

// NewRegistry returns a new registry.
func NewRegistry(l *zap.Logger) *Registry {
	return &Registry{
		m: make(map[string]map[string]*Session),
		l: l,
	}
}

// NewSession creates a new session and adds it to the registry.
func (r *Registry) NewSession(user, db string) *Session {
	r.rw.Lock()
	defer r.rw.Unlock()

	id := uuid.New()
	session := newSession(user, db, id)
	hash := hash(user, db)

	if _, ok := r.m[hash]; !ok {
		r.m[hash] = make(map[string]*Session)
	}

	r.m[hash][id.String()] = session
	r.l.Debug(
		"New session created explicitly",
		zap.String("user", user), zap.String("db", db), zap.String("session", id.String()),
	)

	return session
}

// GetSession returns the session from the registry and updates its lastUsed time.
// If the session does not exist, it creates a new one and adds it to the registry.
// If the session ID is not a valid UUID, it returns nil.
func (r *Registry) GetSession(user, db string, sessionID string) *Session {
	r.rw.RLock()
	defer r.rw.RUnlock()

	id, err := uuid.Parse(sessionID)
	if err != nil {
		r.l.Debug(
			"Invalid session ID",
			zap.String("user", user), zap.String("db", db), zap.String("session", id.String()),
		)
		return nil
	}

	hash := hash(user, db)

	if _, ok := r.m[hash][sessionID]; !ok {
		if _, ok := r.m[hash]; !ok {
			r.m[hash] = make(map[string]*Session)
		}

		r.m[hash][sessionID] = newSession(user, db, id)
		r.l.Debug(
			"New session created implicitly",
			zap.String("user", user), zap.String("db", db), zap.String("session", id.String()),
		)
	}

	r.m[hash][sessionID].lastUsed = time.Now()

	return r.m[hash][sessionID]
}

// RefreshSession updates the last used time for the session.
// If the session does not exist, it does nothing.
func (r *Registry) RefreshSession(user, db, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	hash := hash(user, db)

	if _, ok := r.m[hash][session]; !ok {
		return
	}
	r.l.Debug(
		"Session refreshed",
		zap.String("user", user), zap.String("db", db), zap.String("session", session),
	)

	r.m[hash][session].lastUsed = time.Now()
}

// KillSession removes the session from the registry.
func (r *Registry) KillSession(user, db, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	hash := hash(user, db)

	delete(r.m[hash], session)
	r.l.Debug(
		"Session killed",
		zap.String("user", user), zap.String("db", db), zap.String("session", session),
	)
}

// EndSession marks the session as expired for the future cleanup.
func (r *Registry) EndSession(user, db, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	hash := hash(user, db)

	r.m[hash][session].expired = true
	r.l.Debug(
		"Session marked as expired",
		zap.String("user", user), zap.String("db", db), zap.String("session", session),
	)
}

// CleanExpired removes expired sessions from the registry.
func (r *Registry) CleanExpired() {
	r.rw.Lock()
	defer r.rw.Unlock()

	for username, sessions := range r.m {
		if len(sessions) == 0 {
			delete(r.m, username)
			continue
		}

		for id, session := range sessions {
			if session.expired || time.Since(session.lastUsed) > timeout {
				delete(sessions, id)
			}
		}
	}
}

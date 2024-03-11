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
	m  map[string]map[string]*Session // user -> sessionID -> session

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
func (r *Registry) NewSession(username string) *Session {
	r.rw.Lock()
	defer r.rw.Unlock()

	id := uuid.New()
	session := newSession(id)

	if _, ok := r.m[username]; !ok {
		r.m[username] = make(map[string]*Session)
	}

	r.m[username][id.String()] = session
	r.l.Debug("New session created explicitly", zap.String("username", username), zap.String("session", id.String()))

	return session
}

// GetSession returns the session from the registry and updates its lastUsed time.
// If the session does not exist, it creates a new one and adds it to the registry.
// If the session ID is not a valid UUID, it returns nil.
func (r *Registry) GetSession(username, sessionID string) *Session {
	r.rw.RLock()
	defer r.rw.RUnlock()

	id, err := uuid.Parse(sessionID)
	if err != nil {
		r.l.Debug("Invalid session ID", zap.String("username", username), zap.String("session", sessionID))
		return nil
	}

	if _, ok := r.m[username][sessionID]; !ok {
		if _, ok := r.m[username]; !ok {
			r.m[username] = make(map[string]*Session)
		}

		r.m[username][sessionID] = newSession(id)
		r.l.Debug("New session created implicitly", zap.String("username", username), zap.String("session", sessionID))
	}

	r.m[username][sessionID].lastUsed = time.Now()

	return r.m[username][sessionID]
}

// RefreshSession updates the last used time for the session.
// If the session does not exist, it does nothing.
func (r *Registry) RefreshSession(username, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	if _, ok := r.m[username][session]; !ok {
		return
	}
	r.l.Debug("Session refreshed", zap.String("username", username), zap.String("session", session))

	r.m[username][session].lastUsed = time.Now()
}

// KillSession removes the session from the registry.
func (r *Registry) KillSession(username, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	delete(r.m[username], session)
	r.l.Debug("Session killed", zap.String("username", username), zap.String("session", session))
}

// EndSession marks the session as expired for the future cleanup.
func (r *Registry) EndSession(username, session string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.m[username][session].expired = true
	r.l.Debug("Session marked as expired", zap.String("username", username), zap.String("session", session))
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

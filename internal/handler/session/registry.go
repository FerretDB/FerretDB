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
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "sessions"
)

// Registry stores sessions.
//
//nolint:vet // for readability
type Registry struct {
	rw sync.RWMutex

	// Note that different users can have sessions with the same UUID value.
	// So UUID is not really unique there.
	sessions map[UserID]map[uuid.UUID]*sessionInfo // userID -> sessionID -> sessionInfo, empty UUID for no lsid
	cursors  map[int64]cursorOwner                 // cursorID -> user ID + optional session ID pair

	timeout time.Duration

	l     *slog.Logger
	token *resource.Token

	created  *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// cursorOwner identifies the user ID and session ID that created the cursor.
// A cursor without a session ID is possible, in which case the session ID is empty.
// It happens when `find` or other cursor creating command is called without lsid field.
//
// A cursor always has a user ID, which is the hash of <username>@<database>.
// The user ID of an unauthenticated user is the hash of an empty string.
type cursorOwner struct {
	userID    UserID
	sessionID uuid.UUID // can be empty
}

// NewRegistry returns a new registry.
func NewRegistry(timeout time.Duration, l *slog.Logger) *Registry {
	r := &Registry{
		sessions: map[UserID]map[uuid.UUID]*sessionInfo{},
		cursors:  map[int64]cursorOwner{},
		timeout:  timeout,
		l:        logging.WithName(l, "session"),
		token:    resource.NewToken(),

		created: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "created_total",
				Help:      "Total number of sessions created.",
			},
			[]string{"kind"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "duration_seconds",
				Help:      "Session lifetime in seconds.",
				Buckets: []float64{
					1,
					5,
					10,
					30,
					(1 * time.Minute).Seconds(),
					(5 * time.Minute).Seconds(),
					(10 * time.Minute).Seconds(),
					(30 * time.Minute).Seconds(),
					(1 * time.Hour).Seconds(),
					(4 * time.Hour).Seconds(),
				},
			},
			[]string{"reason"},
		),
	}

	resource.Track(r, r.token)

	return r
}

// NewSession creates a new session and adds it to the registry.
func (r *Registry) NewSession(ctx context.Context) uuid.UUID {
	r.rw.Lock()
	defer r.rw.Unlock()

	sessionID := uuid.New()

	userID := getUserID(ctx)
	s := newSessionInfo()

	if _, ok := r.sessions[userID]; !ok {
		r.sessions[userID] = map[uuid.UUID]*sessionInfo{}
	}

	r.sessions[userID][sessionID] = s
	r.l.DebugContext(ctx,
		"New session created explicitly",
		slog.String("user_id", userID.String()), slog.String("session_id", sessionID.String()),
	)

	r.created.WithLabelValues("explicit").Inc()

	return sessionID
}

// EndSessions marks sessions as ended.
// If a session does not exist, it does nothing.
func (r *Registry) EndSessions(ctx context.Context, sessionIDs []uuid.UUID) {
	r.rw.Lock()
	defer r.rw.Unlock()

	userID := getUserID(ctx)

	for _, sessionID := range sessionIDs {
		if _, ok := r.sessions[userID][sessionID]; !ok {
			continue
		}

		r.sessions[userID][sessionID].ended = true
	}
}

// CreateOrUpdateByLSID fetches `lsid` field from spec and
// updates the last used time of that session.
// If the `lsid` is not a valid UUID, it returns an error.
// If a session does not exist, a new session is created implicitly.
// If `lsid` field is not present, a session is created with an empty session ID.
//
// It returns the user ID and the session ID.
func (r *Registry) CreateOrUpdateByLSID(ctx context.Context, spec wirebson.RawDocument) (UserID, uuid.UUID, error) {
	userID := getUserID(ctx)

	sessionID, err := getSessionUUID(spec)
	if err != nil {
		return UserID{}, uuid.Nil, err
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	r.createOrUpdateSessions(ctx, userID, []uuid.UUID{sessionID})

	return userID, sessionID, nil
}

// ValidateCursor checks if the cursor is created by the same session and the same user.
// If the cursor does not exist, there is nothing to check and no error is returned.
func (r *Registry) ValidateCursor(userID UserID, sessionID uuid.UUID, cursorID int64) error {
	r.rw.RLock()
	defer r.rw.RUnlock()

	owner, ok := r.cursors[cursorID]
	if !ok {
		return nil
	}

	if owner.userID != userID && owner.sessionID == uuid.Nil && sessionID == uuid.Nil {
		return mongoerrors.NewWithArgument(
			mongoerrors.ErrUnauthorized,
			fmt.Sprintf("cursor id %d was not created by the authenticated user", cursorID),
			"getMore",
		)
	}

	if owner.userID != userID || owner.sessionID != sessionID {
		cursorSession := "none"
		if owner.sessionID != uuid.Nil {
			cursorSession = fmt.Sprintf("%s - %s -  - ", owner.sessionID.String(), owner.userID.String())
		}

		currentSession := "none"
		if sessionID != uuid.Nil {
			currentSession = fmt.Sprintf("%s - %s -  - ", sessionID.String(), userID.String())
		}

		msg := fmt.Sprintf(
			"Cursor session id (%s) is not the same as the operation context's session id (%s)",
			cursorSession,
			currentSession,
		)

		return mongoerrors.NewWithArgument(mongoerrors.ErrUnauthorized, msg, "getMore")
	}

	return nil
}

// AddCursor adds the cursor with its user ID and session ID.
// If the session does not exist, a new session is created implicitly.
func (r *Registry) AddCursor(ctx context.Context, userID UserID, sessionID uuid.UUID, cursorID int64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.createOrUpdateSessions(ctx, userID, []uuid.UUID{sessionID})

	if r.sessions[userID][sessionID].cursorIDs == nil {
		r.sessions[userID][sessionID].cursorIDs = map[int64]struct{}{}
	}

	r.sessions[userID][sessionID].cursorIDs[cursorID] = struct{}{}

	r.cursors[cursorID] = cursorOwner{userID: userID, sessionID: sessionID}
}

// DeleteCursor removes the cursor.
// If the cursor does not exist, it does nothing.
// It returns an error if the cursor was not created by the same user.
func (r *Registry) DeleteCursor(userID UserID, cursorID int64, db string) error {
	r.rw.Lock()
	defer r.rw.Unlock()

	owner, ok := r.cursors[cursorID]
	if !ok {
		return nil
	}

	if owner.userID != userID {
		msg := fmt.Sprintf("not authorized on %s to execute command killCursors for cursor %d", db, cursorID)
		return mongoerrors.NewWithArgument(mongoerrors.ErrUnauthorized, msg, "killCursors")
	}

	r.deleteCursor(userID, cursorID)

	return nil
}

// deleteCursor removes the cursor.
// If the cursor was not found or created by the different user,
// it returns false and no cursor is deleted.
//
// It does not hold RWMutex, hence caller should hold RWMutex.
func (r *Registry) deleteCursor(userID UserID, cursorID int64) bool {
	owner, ok := r.cursors[cursorID]
	if !ok || owner.userID != userID {
		return false
	}

	delete(r.cursors, cursorID)

	if r.sessions[userID][owner.sessionID] != nil {
		delete(r.sessions[userID][owner.sessionID].cursorIDs, cursorID)
	}

	return true
}

// CreateOrUpdateSessions updates the last used time of the sessions.
// If a session does not exist, a new session is created implicitly.
func (r *Registry) CreateOrUpdateSessions(ctx context.Context, sessionIDs []uuid.UUID) {
	userID := getUserID(ctx)

	r.rw.Lock()
	defer r.rw.Unlock()

	r.createOrUpdateSessions(ctx, userID, sessionIDs)
}

// createOrUpdateSessions updates the last used time of the sessions.
// If a session does not exist, a new session is created implicitly.
//
// It does not hold RWMutex, hence caller should hold RWMutex.
func (r *Registry) createOrUpdateSessions(ctx context.Context, userID UserID, sessionIDs []uuid.UUID) {
	for _, sessionID := range sessionIDs {
		if _, ok := r.sessions[userID][sessionID]; ok {
			r.sessions[userID][sessionID].lastUsed = time.Now()

			r.l.DebugContext(
				ctx,
				"Session refreshed",
				slog.String("user_id", userID.String()), slog.String("session_id", sessionID.String()),
			)

			continue
		}

		if _, ok := r.sessions[userID]; !ok {
			r.sessions[userID] = map[uuid.UUID]*sessionInfo{}
		}

		r.sessions[userID][sessionID] = newSessionInfo()

		r.l.DebugContext(
			ctx,
			"Session created implicitly",
			slog.String("user_id", userID.String()), slog.String("session_id", sessionID.String()),
		)

		r.created.WithLabelValues("implicit").Inc()
	}
}

// DeleteAllSessions removes all sessions of all users and
// returns all cursors of removed sessions.
func (r *Registry) DeleteAllSessions() []int64 {
	r.rw.Lock()
	defer r.rw.Unlock()

	var cursorIDs []int64

	for _, userID := range slices.Collect(maps.Keys(r.sessions)) {
		sessionIDs := slices.Collect(maps.Keys(r.sessions[userID]))
		userCursorIDs := r.deleteSessions(userID, sessionIDs, "killed")
		cursorIDs = append(cursorIDs, userCursorIDs...)
	}

	must.BeZero(len(r.sessions))
	must.BeZero(len(r.cursors))

	r.sessions = map[UserID]map[uuid.UUID]*sessionInfo{}
	r.cursors = map[int64]cursorOwner{}

	return cursorIDs
}

// DeleteSessionsByUserIDs removes sessions of the specified user IDs and returns cursors of deleted sessions.
// If a user ID does not exist, it does nothing.
func (r *Registry) DeleteSessionsByUserIDs(userIDs []UserID) []int64 {
	r.rw.Lock()
	defer r.rw.Unlock()

	var cursorIDs []int64

	for _, userID := range userIDs {
		sessionIDs := slices.Collect(maps.Keys(r.sessions[userID]))
		userCursorIDs := r.deleteSessions(userID, sessionIDs, "killed")
		cursorIDs = append(cursorIDs, userCursorIDs...)

		must.BeTrue(r.sessions[userID] == nil)
	}

	return cursorIDs
}

// DeleteSessionsByIDs removes sessions and returns cursors of the deleted sessions.
// If a session does not exist, it does nothing.
func (r *Registry) DeleteSessionsByIDs(userID UserID, sessionIDs []uuid.UUID) []int64 {
	r.rw.Lock()
	defer r.rw.Unlock()

	return r.deleteSessions(userID, sessionIDs, "killed")
}

// deleteSessions removes given sessions of the given user and returns cursors of the deleted sessions.
// The `reason` parameter is used for the label of the Prometheus metrics.
//
// It does not hold RWMutex, hence caller should hold RWMutex.
func (r *Registry) deleteSessions(userID UserID, sessionIDs []uuid.UUID, reason string) []int64 {
	var cursorIDs []int64

	for _, sessionID := range sessionIDs {
		info := r.sessions[userID][sessionID]
		if info == nil {
			continue
		}

		for cursorID := range info.cursorIDs {
			if deleted := r.deleteCursor(userID, cursorID); deleted {
				cursorIDs = append(cursorIDs, cursorID)
			}
		}

		delete(r.sessions[userID], sessionID)

		info.close()

		r.duration.WithLabelValues(reason).Observe(time.Since(info.created).Seconds())
	}

	if len(r.sessions[userID]) == 0 {
		delete(r.sessions, userID)
	}

	return cursorIDs
}

// DeleteExpired removes ended sessions and expired session from the registry and
// returns cursors of the deleted sessions.
func (r *Registry) DeleteExpired() []int64 {
	r.rw.Lock()
	defer r.rw.Unlock()

	toEnd := map[UserID][]uuid.UUID{}
	toExpire := map[UserID][]uuid.UUID{}

	for userID, sessions := range r.sessions {
		for sessionID, s := range sessions {
			if s.ended {
				if toEnd[userID] == nil {
					toEnd[userID] = []uuid.UUID{}
				}

				toEnd[userID] = append(toEnd[userID], sessionID)

				continue
			}

			if time.Since(s.lastUsed) > r.timeout {
				if toExpire[userID] == nil {
					toExpire[userID] = []uuid.UUID{}
				}

				toExpire[userID] = append(toExpire[userID], sessionID)
			}
		}
	}

	var cursorIDs []int64

	for userID, sessionIDs := range toEnd {
		userCursorIDs := r.deleteSessions(userID, sessionIDs, "ended")
		cursorIDs = append(cursorIDs, userCursorIDs...)
	}

	for userID, sessionIDs := range toExpire {
		userCursorIDs := r.deleteSessions(userID, sessionIDs, "expired")
		cursorIDs = append(cursorIDs, userCursorIDs...)
	}

	return cursorIDs
}

// Stop stops registry and deletes all sessions.
func (r *Registry) Stop() {
	r.DeleteAllSessions()
	r.sessions = nil
	r.cursors = nil

	resource.Untrack(r, r.token)
}

// Describe implements [prometheus.Collector].
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	r.created.Describe(ch)
	r.duration.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.created.Collect(ch)
	r.duration.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)

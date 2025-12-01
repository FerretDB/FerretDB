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

// Package handler provides implementations of command handlers.
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

const (
	// Minimal supported wire protocol version.
	minWireVersion = int32(0) // needed for some apps and drivers

	// Maximal supported wire protocol version.
	maxWireVersion = int32(21)

	// Maximal supported BSON document size (enforced in DocumentDB by BSON_MAX_ALLOWED_SIZE constant).
	// TODO https://github.com/documentdb/documentdb/issues/67
	// TODO https://github.com/FerretDB/FerretDB/issues/4930
	maxBsonObjectSize = int32(16777216)

	// Maximum size of a batch for inserting data.
	maxWriteBatchSize = int32(100000)

	// Required by C# driver for `IsMaster` and `hello` op reply, without it `DPANIC` is thrown.
	connectionID = int32(42)
)

// Handler provides a set of methods to process clients' requests sent over wire protocol.
//
// The methods msgXXX handle OP_MSG commands.
// CmdQuery handles a limited subset of OP_QUERY messages.
//
// Handler instance is shared between all client connections.
type Handler struct {
	*NewOpts

	p        *documentdb.Pool
	commands map[string]*command
	s        *session.Registry

	runM   sync.Mutex
	runCtx context.Context
	runWG  sync.WaitGroup
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	PostgreSQLURL string
	Auth          bool
	TCPHost       string
	ReplSetName   string

	L             *slog.Logger
	Metrics       *middleware.Metrics
	StateProvider *state.Provider

	SessionCleanupInterval time.Duration
}

// New returns a new handler.
// [Handler.Run] must be called on the returned value.
func New(opts *NewOpts) (*Handler, error) {
	sessionTimeout := time.Duration(session.LogicalSessionTimeoutMinutes) * time.Minute

	// we rely on on it in the `getLog` implementation
	// TODO https://github.com/FerretDB/FerretDB/issues/4750
	_ = opts.L.Handler().(*logging.Handler)

	p, err := documentdb.NewPool(opts.PostgreSQLURL, logging.WithName(opts.L, "pool"), opts.StateProvider)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h := &Handler{
		NewOpts: opts,
		p:       p,
		s:       session.NewRegistry(sessionTimeout, opts.L),
	}

	h.initCommands()

	return h, nil
}

// Run implements [middleware.Handler].
//
// When this method returns, handler is stopped and pool is closed.
func (h *Handler) Run(ctx context.Context) {
	h.runM.Lock()
	h.runCtx = ctx
	h.runM.Unlock()

	defer func() {
		h.runWG.Wait()

		h.s.Stop()
		h.p.Close()
		h.L.InfoContext(ctx, "Stopped")
	}()

	sessionCleanupInterval := h.SessionCleanupInterval
	if sessionCleanupInterval == 0 {
		sessionCleanupInterval = time.Minute
	}

	ticker := time.NewTicker(sessionCleanupInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.L.InfoContext(ctx, "Stopping")
			return

		case <-ticker.C:
			cursorIDs := h.s.DeleteExpired()

			for _, cursorID := range cursorIDs {
				_ = h.p.KillCursor(ctx, cursorID)
			}
		}
	}
}

// Handle implements [middleware.Handler].
func (h *Handler) Handle(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if ctx.Err() != nil {
		return nil, lazyerrors.Error(ctx.Err())
	}

	h.runM.Lock()

	if rc := h.runCtx; rc != nil && rc.Err() != nil {
		h.runM.Unlock()
		return nil, lazyerrors.Error(rc.Err())
	}

	// we need to use Add under a lock to avoid a race with Wait in Run
	h.runWG.Add(1)
	h.runM.Unlock()

	defer h.runWG.Done()

	switch req.WireBody().(type) {
	case *wire.OpMsg:
		msgCmd := req.Document().Command()

		cmd := h.commands[msgCmd]
		if cmd == nil {
			return middleware.ResponseErr(req, mongoerrors.New(
				mongoerrors.ErrCommandNotFound,
				fmt.Sprintf("no such command: '%s'", msgCmd),
			)), nil
		}

		if cmd.handler == nil {
			return middleware.ResponseErr(req, mongoerrors.New(
				mongoerrors.ErrNotImplemented,
				fmt.Sprintf("Command %s is not implemented", msgCmd),
			)), nil
		}

		if h.Auth && !cmd.anonymous {
			conv := conninfo.Get(ctx).Conv()
			succeed := conv.Succeed()
			username := conv.Username()

			if !succeed {
				if conv == nil {
					h.L.WarnContext(ctx, "No existing conversation")
				} else {
					h.L.WarnContext(ctx, "Conversation did not succeed", slog.String("username", username))
				}

				return middleware.ResponseErr(req, mongoerrors.New(
					mongoerrors.ErrUnauthorized,
					fmt.Sprintf("Command %s requires authentication", msgCmd),
				)), nil
			}

			h.L.DebugContext(ctx, "Authentication passed", slog.String("username", username))
		}

		resp, err := cmd.handler(ctx, req)
		if err != nil {
			// TODO https://github.com/FerretDB/FerretDB/issues/4965
			resp = middleware.ResponseErr(req, mongoerrors.Make(ctx, err, "", h.L))
		}

		return resp, nil

	case *wire.OpQuery:
		resp, err := h.CmdQuery(ctx, req)
		if err != nil {
			// TODO https://github.com/FerretDB/FerretDB/issues/4965
			resp = middleware.ResponseErr(req, mongoerrors.Make(ctx, err, "", h.L))
		}

		return resp, nil

	default:
		panic("unsupported request")
	}
}

// Describe implements [prometheus.Collector].
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.p.Describe(ch)
	h.s.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.p.Collect(ch)
	h.s.Collect(ch)
}

// check interfaces
var (
	_ middleware.Handler = (*Handler)(nil)
)

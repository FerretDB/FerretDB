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
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler/operation"
	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

const (
	// Minimal supported wire protocol version.
	minWireVersion = int32(0) // needed for some apps and drivers

	// Maximal supported wire protocol version.
	maxWireVersion = int32(21)

	// Maximal supported BSON document size (enforced in DocumentDB by BSON_MAX_ALLOWED_SIZE constant).
	maxBsonObjectSize = int32(16777216)

	// Maximum size of a batch for inserting data.
	maxWriteBatchSize = int32(100000)

	// Required by C# driver for `IsMaster` and `hello` op reply, without it `DPANIC` is thrown.
	connectionID = int32(42)
)

// Handler provides a set of methods to process clients' requests sent over wire protocol.
//
// MsgXXX methods handle OP_MSG commands.
// CmdQuery handles a limited subset of OP_QUERY messages.
//
// Handler instance is shared between all client connections.
type Handler struct {
	*NewOpts

	commands map[string]*command

	operations *operation.Registry
	s          *session.Registry
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Pool *documentdb.Pool
	Auth bool

	TCPHost     string
	ReplSetName string

	L             *slog.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	SessionCleanupInterval time.Duration
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	sessionTimeout := time.Duration(session.LogicalSessionTimeoutMinutes) * time.Minute

	// we rely on on it in the `getLog` implementation
	_ = opts.L.Handler().(*logging.Handler)

	h := &Handler{
		NewOpts: opts,

		operations: operation.NewRegistry(),
		s:          session.NewRegistry(sessionTimeout, opts.L),
	}

	h.initCommands()

	return h, nil
}

// Run runs the handler until ctx is canceled.
func (h *Handler) Run(ctx context.Context) {
	defer func() {
		h.s.Stop()
		h.operations.Close()
		h.L.InfoContext(ctx, "Handler stopped")
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
			h.L.InfoContext(ctx, "Expired session deletion stopped")
			return

		case <-ticker.C:
			cursorIDs := h.s.DeleteExpired()

			for _, cursorID := range cursorIDs {
				_ = h.Pool.KillCursor(ctx, cursorID)
			}
		}
	}
}

// Describe implements [prometheus.Collector].
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.Pool.Describe(ch)
	h.s.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.Pool.Collect(ch)
	h.s.Collect(ch)
}

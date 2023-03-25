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

// Package sqlite provides SQLite handler implementation.
package sqlite

import (
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return commonerrors.NewCommandErrorMsg(commonerrors.ErrNotImplemented, "I'm a stub, not a real handler for "+command)
}

// Handler implements handlers.Interface by stubbing all methods except the following handler-independent commands:
//
//   - buildInfo;
//   - connectionStatus;
//   - debugError;
//   - getCmdLineOpts;
//   - getFreeMonitoringStatus;
//   - hostInfo;
//   - listCommands;
//   - setFreeMonitoringStatus;
//   - whatsmyuri.
type Handler struct {
	*NewOpts
}

type NewOpts struct {
	SQLiteDBPath string

	L             *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	if opts.SQLiteDBPath == "" {
		return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrNotImplemented, "SQLiteDBPath is not provided")
	}

	return &Handler{
		NewOpts: opts,
	}, nil
}

// Close implements handlers.Interface.
func (h *Handler) Close() {}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

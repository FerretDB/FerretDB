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

// Package dummy provides a basic handler implementation.
//
// The whole package can be copied to start a new handler implementation.
package dummy

import (
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return common.NewCommandErrorMsg(common.ErrNotImplemented, "I'm a stub, not a real handler for "+command)
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
type Handler struct{}

// New returns a new handler.
func New() (handlers.Interface, error) {
	return new(Handler), nil
}

// Close implements handlers.Interface.
func (h *Handler) Close() {}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)

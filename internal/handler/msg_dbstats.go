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

package handler

import (
	"context"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// MsgDBStats implements `dbStats` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgDBStats(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/9
	return nil, mongoerrors.New(mongoerrors.ErrNotImplemented, `"dbStats" is not implemented in DocumentDB yet`)
}

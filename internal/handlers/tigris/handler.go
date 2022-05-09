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

package tigris

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Handler data struct.
type Handler struct {
	client    *Client
	l         *zap.Logger
	startTime time.Time
}

// NewOpts represents handler configuration.
type NewOpts struct {
	Conn      *Client
	L         *zap.Logger
	StartTime time.Time
}

// New returns a new handler.
func New(opts *NewOpts) common.Handler {
	return &Handler{
		client:    opts.Conn,
		l:         opts.L,
		startTime: opts.StartTime,
	}
}

func (h *Handler) Close() {
	h.client.conn.Close()
}

func (h *Handler) MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgGetCmdLineOpts(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

func (h *Handler) MsgWhatsMyURI(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	panic("not implemented")
}

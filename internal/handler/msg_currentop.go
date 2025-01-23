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
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgCurrentOp implements `currentOp` command.
//
// The passed context is canceled when the client connection is closed.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3974
func (h *Handler) MsgCurrentOp(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	ops := h.operations.Operations()
	inProgress := wirebson.MakeArray(len(ops))

	for _, op := range ops {
		since := time.Since(op.CurrentOpTime)

		var ns string
		if op.DB != "" && op.Collection != "" {
			ns = op.DB + "." + op.Collection
		}

		opCommand := op.Command
		if opCommand == nil {
			opCommand = must.NotFail(wirebson.NewDocument())
		}

		doc := wirebson.MustDocument(
			"type", "op",
			"active", op.Active,
			"currentOpTime", time.Now().Format(time.RFC3339),
			"opid", op.OpID,
			"secs_running", int64(since.Truncate(time.Second).Seconds()),
			"microsecs_running", since.Microseconds(),
			"op", op.Op,
			"ns", ns,
			"command", opCommand,
		)

		must.NoError(inProgress.Add(doc))
	}

	return wire.MustOpMsg(
		"inprog", inProgress,
		"ok", float64(1),
	), nil
}

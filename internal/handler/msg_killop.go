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
	"fmt"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgKillOp implements `killOp` command.
//
// The passed context is canceled when the client connection is closed.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3974
func (h *Handler) MsgKillOp(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	if dbName != "admin" {
		m := "killOp may only be run against the admin database."
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrUnauthorized, m, "killOp")
	}

	v := doc.Get("op")
	if v == nil {
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrNoSuchKey, `Missing expected field "op"`, "killOp")
	}

	var op int32

	switch v := v.(type) {
	case float64:
		if v != float64(int32(v)) {
			m := fmt.Sprintf(`Expected field "op" to have a value exactly representable as a 64-bit integer, but found op: %g`, v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, m, "killOp")
		}

		op = int32(v)
	case int32:
		op = v
	case int64:
		if v != int64(int32(v)) {
			m := fmt.Sprintf("invalid op : %d. Op ID cannot be represented with 32 bits", v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation26823, m, "killOp")
		}

		op = int32(v)
	default:
		m := fmt.Sprintf(`Expected field "op" to have numeric type, but found %T`, v)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, "killOp")
	}

	h.operations.Kill(op)

	return wire.MustOpMsg(
		"info", "attempting to kill op",
		"ok", float64(1),
	), nil
}

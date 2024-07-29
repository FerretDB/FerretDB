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
	"errors"
	"strconv"
	"strings"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgDebugError implements `debugError` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgDebugError(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	expected, err := common.GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
	}

	// check if parameter is an error code
	if n, err := strconv.ParseInt(expected, 10, 32); err == nil {
		errCode := handlererrors.ErrorCode(n)
		return nil, errors.New(errCode.String())
	}

	switch {
	case expected == "ok":
		return bson.NewOpMsg(must.NotFail(types.NewDocument(
			"ok", float64(1),
		)))

	case strings.HasPrefix(expected, "panic"):
		panic("debugError " + expected)

	case strings.HasPrefix(expected, "lazy"):
		return nil, lazyerrors.New(expected)

	default:
		return nil, errors.New(expected)
	}
}

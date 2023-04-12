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

package common

import (
	"context"
	"errors"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDebugError is a common implementation of the debugError command.
func MsgDebugError(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	expected, err := GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
	}

	// check if parameter is an error code
	if n, err := strconv.Atoi(expected); err == nil {
		errCode := commonerrors.ErrorCode(n)
		return nil, errors.New(errCode.String())
	}

	switch expected {
	case "ok":
		var reply wire.OpMsg

		replyDoc := must.NotFail(types.NewDocument(
			"ok", float64(1),
		))

		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{replyDoc},
		}))

		return &reply, nil

	case "panic":
		panic("debugError panic")

	default:
		return nil, errors.New(expected)
	}
}

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

package middleware

import (
	"context"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"log/slog"
)

type Error struct {
	L *slog.Logger
}

func (e *Error) HandleOpMsg(next MsgHandlerFunc) MsgHandlerFunc {
	return func(ctx context.Context, req *MsgRequest) (resp *MsgResponse, err error) {
		defer func() {
			if err != nil {
				resp, err = MsgErrorResponse(ctx, err, "", e.L)
			}
		}()

		return next(ctx, req)
	}
}

func (e *Error) HandleOpReply(next QueryHandlerFunc) QueryHandlerFunc {
	return func(ctx context.Context, req *QueryRequest) (resp *ReplyResponse, err error) {
		defer func() {
			if err != nil {
				resp, err = ReplyError(ctx, err, "", e.L)
			}
		}()

		return next(ctx, req)
	}
}

// MsgErrorResponse constructs a [*MsgResponse] from an error.
func MsgErrorResponse(ctx context.Context, handlerError error, arg string, l *slog.Logger) (*MsgResponse, error) {
	err := mongoerrors.Make(ctx, handlerError, arg, l)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &MsgResponse{
		OpMsg: err.Msg(),
		Error: err,
	}, nil
}

// ReplyError constructs a [*ReplyResponse] from an error.
func ReplyError(ctx context.Context, handlerError error, arg string, l *slog.Logger) (*ReplyResponse, error) {
	err := mongoerrors.Make(ctx, handlerError, arg, l)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &ReplyResponse{
		OpReply: err.Reply(),
		Error:   err,
	}, nil
}

// check interface.
var _ Middleware = (*Error)(nil)

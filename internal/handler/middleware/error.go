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
	"log/slog"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// Error is a middleware that converts an error into a response.
type Error struct {
	l   *slog.Logger
	arg string
}

// NewError creates a new Error middleware.
func NewError(arg string, l *slog.Logger) *Error {
	return &Error{
		arg: arg,
		l:   l,
	}
}

// HandleOpMsg implements Middleware.
//
// If an error was returned from `next` handler, it converts the error to OP_MSG response.
// It also sets the Error field in the response, for the later use by such as observability.
// HandleOpMsg always returns nil error.
func (e *Error) HandleOpMsg(next MsgHandlerFunc) MsgHandlerFunc {
	return func(ctx context.Context, req *MsgRequest) (resp *MsgResponse, err error) {
		defer func() {
			if err != nil {
				mErr := mongoerrors.Make(ctx, err, e.arg, e.l)

				resp = &MsgResponse{
					OpMsg: mErr.Msg(),
					Error: mErr,
				}

				err = nil
			}
		}()

		return next(ctx, req)
	}
}

// HandleOpReply implements Middleware.
//
// If an error was returned from `next` handler, it converts the error to OP_REPLY response.
// It also sets the Error field in the response, for the later use by such as observability.
// HandleOpReply always returns nil error.
func (e *Error) HandleOpReply(next QueryHandlerFunc) QueryHandlerFunc {
	return func(ctx context.Context, req *QueryRequest) (resp *ReplyResponse, err error) {
		defer func() {
			if err != nil {
				mErr := mongoerrors.Make(ctx, err, e.arg, e.l)

				resp = &ReplyResponse{
					OpReply: mErr.Reply(),
					Error:   mErr,
				}

				err = nil
			}
		}()

		return next(ctx, req)
	}
}

// check interface.
var _ Middleware = (*Error)(nil)

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
)

func MsgError(next HandlerFunc[*MsgRequest, *MsgResponse], l *slog.Logger) HandlerFunc[*MsgRequest, *MsgResponse] {
	return func(ctx context.Context, req *MsgRequest) (resp *MsgResponse, err error) {
		defer func() {
			if err != nil {
				resp, err = MsgErrorResponse(ctx, err, "", l)
			}
		}()

		return next(ctx, req)
	}
}

func QueryError(next HandlerFunc[*QueryRequest, *ReplyResponse], l *slog.Logger) HandlerFunc[*QueryRequest, *ReplyResponse] {
	return func(ctx context.Context, req *QueryRequest) (resp *ReplyResponse, err error) {
		defer func() {
			if err != nil {
				resp, err = ReplyError(ctx, err, "", l)
			}
		}()

		return next(ctx, req)
	}
}

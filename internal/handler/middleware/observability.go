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

// OpMsgObservability is a middleware that will wrap the command handler with common observability logic.
func OpMsgObservability(next MsgHandlerFunc, l *slog.Logger) MsgHandlerFunc {
	return observability[*MsgRequest, *MsgResponse](next)
}

// OpQueryObservability is a middleware that will wrap the query handler with common observability logic.
func OpQueryObservability(next QueryHandlerFunc, l *slog.Logger) QueryHandlerFunc {
	return observability[*CmdQuery, *CmdReply](next)
}

// observability wraps the handler with logs, traces, and metrics.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4439
func observability[ReqType any, RespType any](
	next func(context.Context, ReqType) (RespType, error),
) func(context.Context, ReqType) (RespType, error) {
	return func(ctx context.Context, req ReqType) (RespType, error) {
		return next(ctx, req)
	}
}

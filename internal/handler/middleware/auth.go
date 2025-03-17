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
	"fmt"
	"log/slog"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// Auth is a middleware that wraps the command handler with authentication check.
//
// Context must contain [*conninfo.ConnInfo].
func Auth(next MsgHandlerFunc, l *slog.Logger, command string) MsgHandlerFunc {
	return func(ctx context.Context, req *MsgRequest) (*MsgResponse, error) {
		conv := conninfo.Get(ctx).Conv()
		succeed := conv.Succeed()
		username := conv.Username()

		switch {
		case conv == nil:
			l.WarnContext(ctx, "No existing conversation")

		case !succeed:
			l.WarnContext(ctx, "Conversation did not succeed", slog.String("username", username))

		default:
			l.DebugContext(ctx, "Authentication passed", slog.String("username", username))

			return next(ctx, req)
		}

		msg := fmt.Sprintf("Command %s requires authentication", command)
		return nil, mongoerrors.New(mongoerrors.ErrUnauthorized, msg)
	}
}

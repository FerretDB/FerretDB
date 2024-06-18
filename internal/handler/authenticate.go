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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
)

// authenticate validates user credentials.
// If EnableNewAuth is set, stored SCRAM conversation is validated and
// the backend authentication is bypassed.
// If EnableNewAuth not set, `PLAIN` backend authentication is used.
func (h *Handler) authenticate(ctx context.Context) error {
	_, _, mechanism, conv := conninfo.Get(ctx).Auth()

	if !h.EnableNewAuth {
		switch mechanism {
		case "PLAIN", "":
			return nil
		default:
			msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
				"See https://docs.ferretdb.io/security/authentication/ for more details."
			return handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, mechanism)
		}
	}

	conninfo.Get(ctx).SetBypassBackendAuth()

	switch mechanism {
	case "SCRAM-SHA-256", "SCRAM-SHA-1": //nolint:goconst // we don't need a constant for this
		if conv == nil || !conv.Valid() {
			return handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrAuthenticationFailed,
				"Authentication failed",
				"authenticate",
			)
		}

		return nil
	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, mechanism)
	}
}

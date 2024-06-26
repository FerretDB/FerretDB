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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
)

// checkAuthentication returns error if new authentication is enabled and SCRAM conversation is not valid.
func (h *Handler) checkAuthentication(ctx context.Context) error {
	if !h.EnableNewAuth {
		return nil
	}

	_, _, conv := conninfo.Get(ctx).Auth()
	if conv == nil || !conv.Valid() {
		return handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed",
			"checkAuthentication",
		)
	}

	return nil
}

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

package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// CmdQuery implements HandlerInterface.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	if err := common.CheckClientMetadata(ctx, query.Query); err != nil {
		return nil, lazyerrors.Error(err)
	}

	cmd := query.Query.Command()
	collection := query.FullCollectionName

	// both are valid and are allowed to be run against any database as we don't support authorization yet
	if (cmd == "ismaster" || cmd == "isMaster") && strings.HasSuffix(collection, ".$cmd") {
		return common.IsMaster()
	}

	// defaults to the database name if supplied on the connection string or $external
	if cmd == "saslStart" && strings.HasSuffix(collection, ".$cmd") {
		var emptyPayload types.Binary

		if err := common.SASLStart(ctx, query.Query); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if _, err := h.DBPool(ctx); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsInvalidAuthorizationSpecification(pgErr.Code) {
				msg := "FerretDB failed to authenticate you in PostgreSQL:\n" +
					strings.TrimSpace(pgErr.Error()) + "\n" +
					"See https://docs.ferretdb.io/security/authentication/ for more details."

				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrAuthenticationFailed,
					msg,
					"OpQuery: "+cmd,
				)
			}

			return nil, lazyerrors.Error(err)
		}

		return &wire.OpReply{
			NumberReturned: 1,
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"conversationId", int32(1),
				"done", true,
				"payload", emptyPayload,
				"ok", float64(1),
			))},
		}, nil
	}

	return nil, commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrNotImplemented,
		fmt.Sprintf("CmdQuery: unhandled command %q for collection %q", cmd, collection),
		"OpQuery: "+cmd,
	)
}

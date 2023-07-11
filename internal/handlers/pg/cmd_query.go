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
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// CmdQuery implements HandlerInterface.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	cmd := query.Query.Command()
	collection := query.FullCollectionName

	// both are valid
	if (cmd == "ismaster" || cmd == "isMaster") && collection == "admin.$cmd" {
		return common.IsMaster(ctx, query)
	}

	// defaults to the database name if supplied on the connection string or $external
	if cmd == "saslStart" &&
		(collection == "$external.$cmd" || strings.HasSuffix(collection, "$cmd")) {
		// TODO.
	}

	return nil, commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrNotImplemented,
		fmt.Sprintf("CmdQuery: unhandled command %q for collection %q", cmd, collection),
		"OpQuery: "+cmd,
	)
}

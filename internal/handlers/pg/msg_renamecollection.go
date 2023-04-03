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
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/jackc/pgx/v4"
)

// MsgRenameCollection implements HandlerInterface.
func (h *Handler) MsgRenameCollection(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "dropTarget", "writeConcern", "comment"); err != nil {
		return nil, err
	}

	// to: "<target_namespace>",
	// dropTarget: <true|false>,
	// writeConcern: <document>,
	// comment: <any> }

	var collectionParam string
	if collectionParam, err = common.GetRequiredParam[string](document, document.Command()); err != nil {
		return nil, err
	}

	var newCollection string
	if newCollection, err = common.GetRequiredParam[string](document, "to"); err != nil {
		return nil, err
	}

	sourceDB, sourceColl, err := extractFromNamespace(collectionParam)
	if err != nil {
		return nil, err
	}

	destDB, destColl, err := extractFromNamespace(newCollection)
	if err != nil {
		return nil, err
	}

	if sourceDB != destDB {
		// TODO unimplemented err
	}

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		return pgdb.RenameCollection(ctx, tx, sourceDB, sourceColl, destColl)
	})
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			// TODO
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

func extractFromNamespace(namespace string) (string, string, error) {
	s := strings.Split(namespace, ".")

	if len(s) != 2 {
		// TODO
		return "", "", fmt.Errorf("wrong namespace")
	}

	return s[0], s[1], nil
}

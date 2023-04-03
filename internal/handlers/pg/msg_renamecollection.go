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
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
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

	fromNamespace, err := common.GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
	}

	toNamespace, err := common.GetRequiredParam[string](document, "to")
	if err != nil {
		return nil, err
	}

	if fromNamespace == toNamespace {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrIllegalOperation,
			"Can't rename a collection to itself",
		)
	}

	fromDB, fromColl, err := extractFromNamespace(fromNamespace)
	if err != nil {
		return nil, err
	}

	_, toColl, err := extractFromNamespace(toNamespace)
	if err != nil {
		return nil, err
	}

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		return pgdb.RenameCollection(ctx, tx, fromDB, fromColl, toColl)
	})

	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// extractFromNamespace returns the database and collection name from a given namespace.
func extractFromNamespace(namespace string) (string, string, error) {
	split := strings.Split(namespace, ".")

	// TODO: validate namespace.
	// we assume that the given namespace contains a single dot.
	if len(split) != 2 {
		return "", "", commonerrors.NewCommandErrorMsg(
			commonerrors.ErrInvalidNamespace,
			"Invalid namespace specified "+namespace,
		)
	}

	return split[0], split[1], nil
}

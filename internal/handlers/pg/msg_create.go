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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCreate implements HandlerInterface.
func (h *Handler) MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"capped",
		"timeseries",
		"expireAfterSeconds",
		"size",
		"max",
		"validator",
		"validationLevel",
		"validationAction",
		"viewOn",
		"pipeline",
		"collation",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"autoIndexId",
		"storageEngine",
		"indexOptionDefaults",
		"writeConcern",
		"comment",
	}
	common.Ignored(document, h.L, ignoredFields...)

	command := document.Command()

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	// We use two separate transactions as there is a case when a query of the first transaction could fail,
	// and we should consider it normal: it could happen if we attempt to create databases from two parallel requests.
	// One of such requests will fail as the database was already created from another request,
	// but it's a normal situation, and we should create both collections in such a case.

	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		if err := pgdb.CreateDatabaseIfNotExists(ctx, tx, db); err != nil {
			switch {
			case errors.Is(err, pgdb.ErrAlreadyExist):
				// If the DB was created from a parallel query, it's ok.
				// However, in this case one of the transaction queries failed,
				// so we need to rollback the transaction.
				return pgdb.ErrAlreadyExist
			case errors.Is(pgdb.ErrInvalidDatabaseName, err):
				msg := fmt.Sprintf("Invalid namespace: %s.%s", db, collection)
				return common.NewCommandErrorMsg(common.ErrInvalidNamespace, msg)
			default:
				return lazyerrors.Error(err)
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, pgdb.ErrAlreadyExist) {
		return nil, err
	}

	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		if err := pgdb.CreateCollection(ctx, tx, db, collection); err != nil {
			switch {
			case errors.Is(err, pgdb.ErrAlreadyExist):
				msg := fmt.Sprintf("Collection %s.%s already exists.", db, collection)
				return common.NewCommandErrorMsg(common.ErrNamespaceExists, msg)
			case errors.Is(err, pgdb.ErrInvalidTableName):
				msg := fmt.Sprintf("Invalid collection name: '%s.%s'", db, collection)
				return common.NewCommandErrorMsg(common.ErrInvalidNamespace, msg)
			default:
				return lazyerrors.Error(err)
			}
		}
		return nil
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

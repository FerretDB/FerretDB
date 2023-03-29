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
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDropIndexes implements HandlerInterface.
func (h *Handler) MsgDropIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "writeConcern", "comment")

	command := document.Command()

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	indexNames, err := getIndexesParam(document, command)
	if err != nil {
		return nil, err
	}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		// TODO use iterator
		for _, indexName := range indexNames {
			if err := pgdb.DropIndex(ctx, tx, db, collection, indexName); err != nil {
				return err
			}
		}

		return nil
	})

	switch {
	case err == nil:
		// nothing
	case errors.Is(err, pgdb.ErrTableNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("ns not found %s.%s", db, collection),
			command,
		)
	case errors.Is(err, pgdb.ErrIndexNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			"index not found",
			command,
		)
	case errors.Is(err, pgdb.ErrIndexCannotDelete):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidOptions,
			"cannot drop _id index",
			command,
		)
	default:
		return nil, lazyerrors.Error(err)
	}

	if err != nil {
		// todo handle errors
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// getIntexesParam gets index from the document.
func getIndexesParam(doc *types.Document, command string) ([]string, error) {
	v, err := doc.Get("index")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	switch v := v.(type) {
	case *types.Array:
		var indexes []string

		for {
			iter := v.Iterator()

			_, val, err := iter.Next()
			switch {
			case err == nil:
				// nothing
			case errors.Is(err, iterator.ErrIteratorDone):
				return indexes, nil
			default:
				return nil, lazyerrors.Error(err)
			}

			index, ok := val.(string)
			if !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object']",
						pjson.GetTypeOfValue(v),
					),
					command,
				)
			}

			indexes = append(indexes, index)
		}
	case *types.Document:
	case string:
		return []string{v}, nil
	}

	return nil, commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrTypeMismatch,
		fmt.Sprintf(
			"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object']",
			pjson.GetTypeOfValue(v),
		),
		command,
	)
}

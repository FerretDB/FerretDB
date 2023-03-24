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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCreateIndexes implements HandlerInterface.
func (h *Handler) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "writeConcern", "commitQuorum", "comment")

	command := document.Command()

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	idxArr, err := common.GetRequiredParam[*types.Array](document, "indexes")
	if err != nil {
		return nil, err
	}

	iter := idxArr.Iterator()
	defer iter.Close()

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		for {
			var val any
			_, val, err = iter.Next()

			switch {
			case err == nil:
				// do nothing
			case errors.Is(err, iterator.ErrIteratorDone):
				// iterator is done, no more indexes to create
				return nil
			default:
				return lazyerrors.Error(err)
			}

			indexDoc, ok := val.(*types.Document)
			if !ok {
				return lazyerrors.Errorf("expected index document, got %T", val)
			}

			var index *pgdb.Index

			index, err = processIndexOptions(indexDoc)
			if err != nil {
				return err
			}

			if err = pgdb.CreateIndexIfNotExists(ctx, tx, db, collection, index); err != nil {
				return err
			}
		}
	})

	switch {
	case err == nil:
	// do nothing
	default:
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

func processIndexOptions(indexDoc *types.Document) (*pgdb.Index, error) {
	var index pgdb.Index

	iter := indexDoc.Iterator()

	for {
		opt, _, err := iter.Next()

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			return &index, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		switch opt {
		case "key":
			var keyDoc *types.Document

			keyDoc, err = common.GetRequiredParam[*types.Document](indexDoc, "key")
			if err != nil {
				return nil, err
			}

			index.Key, err = processIndexKey(keyDoc)
			if err != nil {
				return nil, err
			}

		case "name":
			index.Name, err = common.GetRequiredParam[string](indexDoc, "name")
			if err != nil {
				return nil, err
			}

		case "unique", "sparse", "partialFilterExpression", "expireAfterSeconds", "hidden", "storageEngine",
			"weights", "default_language", "language_override", "textIndexVersion", "2dsphereIndexVersion",
			"bits", "min", "max", "bucketSize", "collation", "wildcardProjection":
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index option %q is not implemented yet", opt),
				"createIndexes",
			)

		default:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("Index option %q is unknown", opt),
				"createIndexes",
			)
		}
	}
}

// processIndexKey processes the document containing the index key.
func processIndexKey(keyDoc *types.Document) (pgdb.IndexKey, error) {
	res := make(pgdb.IndexKey, 0, keyDoc.Len())
	keyIter := keyDoc.Iterator()

	for {
		field, order, err := keyIter.Next()

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			return res, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		var orderParam int64

		if orderParam, err = common.GetWholeNumberParam(order); err != nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index key value %q is not implemented yet", order),
				"createIndexes",
			)
		}

		var indexOrder pgdb.IndexOrder

		switch orderParam {
		case 1:
			indexOrder = pgdb.IndexOrderAsc
		case -1:
			indexOrder = pgdb.IndexOrderDesc
		default:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index key value %q is not implemented yet", orderParam),
				"createIndexes",
			)
		}

		res = append(res, pgdb.IndexKeyPair{
			Field: field,
			Order: indexOrder,
		})
	}
}

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

	"github.com/jackc/pgx/v5"

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

	var nIndexesWas int32
	var responseMsg string

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		nIndexesWas, responseMsg, err = processIndexDrop(ctx, tx, db, collection, document, command)
		return err
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
	case errors.Is(err, pgdb.ErrIndexCannotDelete):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidOptions,
			"cannot drop _id index",
			command,
		)
	default:
		return nil, lazyerrors.Error(err)
	}

	replyDoc := must.NotFail(types.NewDocument(
		"nIndexesWas", nIndexesWas,
	))

	if responseMsg != "" {
		replyDoc.Set("msg", responseMsg)
	}

	replyDoc.Set(
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

// processIndexDrop parses index doc and processes index deletion based on the provided params.
// Upon successful drop of index, it returns the number of indexes we had in the database
// before the indexes were dropped, and if any message related to particular indexes being dropped.
func processIndexDrop(ctx context.Context, tx pgx.Tx, db, collection string, doc *types.Document, command string) (int32, string, error) { //nolint:lll // for readability
	v, err := doc.Get("index")
	if err != nil {
		return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'dropIndexes.index' is missing but a required field",
			command,
		)
	}

	var nIndexesWas int32

	switch v := v.(type) {
	case *types.Document:
		// Index specification (key) is provided to drop a specific index.
		var indexKey pgdb.IndexKey

		indexKey, err = processIndexKey(v)
		if err != nil {
			return 0, "", lazyerrors.Error(err)
		}

		nIndexesWas, err = pgdb.DropIndex(ctx, tx, db, collection, &pgdb.Index{Key: indexKey})

		switch {
		case err == nil:
			return nIndexesWas, "", nil
		case errors.Is(err, pgdb.ErrIndexNotExist):
			return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrIndexNotFound,
				fmt.Sprintf("can't find index with key: %s", types.FormatAnyValue(v)),
				command,
			)
		default:
			return 0, "", lazyerrors.Error(err)
		}
	case *types.Array:
		// List of index names is provided to drop multiple indexes.
		iter := v.Iterator()

		defer iter.Close() // It's safe to defer here as the iterators reads everything.

		for {
			var val any
			_, val, err = iter.Next()

			switch {
			case err == nil:
				// nothing
			case errors.Is(err, iterator.ErrIteratorDone):
				return nIndexesWas, "", nil
			default:
				return 0, "", lazyerrors.Error(err)
			}

			index, ok := val.(string)
			if !ok {
				return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
						pjson.GetTypeOfValue(v),
					),
					command,
				)
			}

			var nsIndexesWasCurrent int32
			nsIndexesWasCurrent, err = pgdb.DropIndex(ctx, tx, db, collection, &pgdb.Index{Name: index})

			// nIndexesWas should give the number of indexes we had in the database before we start dropping indexes.
			if nIndexesWas == 0 {
				nIndexesWas = nsIndexesWasCurrent
			}

			switch {
			case err == nil:
				continue
			case errors.Is(err, pgdb.ErrIndexNotExist):
				return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrIndexNotFound,
					fmt.Sprintf("index not found with name [%s]", index),
					command,
				)

			default:
				return 0, "", lazyerrors.Error(err)
			}
		}
	case string:
		if v == "*" {
			// Drop all indexes except the _id index.
			nIndexesWas, err = pgdb.DropAllIndexes(ctx, tx, db, collection)
			if err != nil {
				return 0, "", lazyerrors.Error(err)
			}

			return nIndexesWas, "non-_id indexes dropped for collection", nil
		}

		// Index name is provided to drop a specific index.
		nIndexesWas, err = pgdb.DropIndex(ctx, tx, db, collection, &pgdb.Index{Name: v})

		switch {
		case err == nil:
			return nIndexesWas, "", nil
		case errors.Is(err, pgdb.ErrIndexNotExist):
			return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrIndexNotFound,
				fmt.Sprintf("index not found with name [%s]", v),
				command,
			)
		default:
			return 0, "", lazyerrors.Error(err)
		}
	}

	return 0, "", commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrTypeMismatch,
		fmt.Sprintf(
			"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
			pjson.GetTypeOfValue(v),
		),
		command,
	)
}

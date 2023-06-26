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

	"github.com/AlekSi/pointer"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
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

	if collection == "" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", db),
			command,
		)
	}

	v, _ := document.Get("indexes")
	if v == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'createIndexes.indexes' is missing but a required field",
			document.Command(),
		)
	}

	idxArr, ok := v.(*types.Array)
	if !ok {
		if _, ok = v.(types.NullType); ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrIndexesWrongType,
				"invalid parameter: expected an object (indexes)",
				document.Command(),
			)
		}

		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"BSON field 'createIndexes.indexes' is the wrong type, expected type 'array'",
			document.Command(),
		)
	}

	if idxArr.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"Must specify at least one index to create",
			document.Command(),
		)
	}

	iter := idxArr.Iterator()
	defer iter.Close()

	duplicateChecker := map[string]struct{}{}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		for {
			var key, val any
			key, val, err = iter.Next()

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
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf("BSON field 'createIndexes.indexes.%d' is the wrong type, expected type 'object'", key),
					document.Command(),
				)
			}

			var index *pgdb.Index

			if index, err = processIndexOptions(indexDoc); err != nil {
				return err
			}

			if index.Name == "" {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrCannotCreateIndex,
					fmt.Sprintf(
						"Error in specification %s :: caused by :: index name cannot be empty",
						types.FormatAnyValue(indexDoc),
					),
					document.Command(),
				)
			}

			if _, ok := duplicateChecker[index.Name]; ok {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrIndexAlreadyExists,
					fmt.Sprintf("Identical index already exists: %s", index.Name),
					document.Command(),
				)
			}

			duplicateChecker[index.Name] = struct{}{}

			err = pgdb.CreateIndexIfNotExists(ctx, tx, db, collection, index)
			if errors.Is(err, pgdb.ErrIndexKeyAlreadyExist) && index.Name == "_id_1" {
				// ascending _id index is created by default
				return nil
			}

			if err != nil {
				return err
			}
		}
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgdb.ErrIndexKeyAlreadyExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIndexOptionsConflict,
			"One of the specified indexes already exists with a different name",
			document.Command(),
		)
	case errors.Is(err, pgdb.ErrIndexNameAlreadyExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"One of the specified indexes already exists with a different key",
			document.Command(),
		)
	case errors.Is(err, pgdb.ErrUniqueViolation):
		// TODO: Add test for this case https://github.com/FerretDB/FerretDB/issues/2852
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrDuplicateKeyInsert,
			"Index build failed",
			document.Command(),
		)
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

// processIndexOptions processes the given indexDoc and returns a pgdb.Index.
func processIndexOptions(indexDoc *types.Document) (*pgdb.Index, error) {
	var index pgdb.Index

	iter := indexDoc.Iterator()
	defer iter.Close()

	var hasValue bool
	for {
		opt, _, err := iter.Next()

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			if !hasValue {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					fmt.Sprintf(
						"Error in specification {} :: caused by :: "+
							"The 'key' field is a required property of an index specification",
					),
					"createIndexes",
				)
			}

			return &index, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		hasValue = true

		// Process required param "key"
		var keyDoc *types.Document

		keyDoc, err = common.GetRequiredParam[*types.Document](indexDoc, "key")
		if err != nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"'key' option must be specified as an object",
				"createIndexes",
			)
		}

		if keyDoc.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrCannotCreateIndex,
				"Must specify at least one field for the index key",
				"createIndexes",
			)
		}

		// Special case: if keyDocs consists of a {"_id": -1} only, an error should be returned.
		if keyDoc.Len() == 1 {
			var val any
			var order int64

			if val, err = keyDoc.Get("_id"); err == nil {
				if order, err = commonparams.GetWholeNumberParam(val); err == nil && order == -1 {
					return nil, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrBadValue,
						"The field 'key' for an _id index must be {_id: 1}, but got { _id: -1 }",
						"createIndexes",
					)
				}
			}
		}

		index.Key, err = processIndexKey(keyDoc)
		if err != nil {
			return nil, err
		}

		v, _ := indexDoc.Get("name")
		if v == nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf(
					"Error in specification { key: %s } :: caused by :: "+
						"The 'name' field is a required property of an index specification",
					types.FormatAnyValue(keyDoc),
				),
				"createIndexes",
			)
		}

		var ok bool
		index.Name, ok = v.(string)

		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"'name' option must be specified as a string",
				"createIndexes",
			)
		}

		switch opt {
		case "key", "name":
			// already processed, do nothing

		case "unique":
			uniqueVal := must.NotFail(indexDoc.Get("unique"))

			_, ok := uniqueVal.(bool)
			if !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"Error in specification { key: %s, name: \"%s\", unique: %s } "+
							":: caused by :: "+
							"The field 'unique' has value unique: %[3]s, which is not convertible to bool",
						types.FormatAnyValue(must.NotFail(indexDoc.Get("key"))),
						index.Name, types.FormatAnyValue(uniqueVal),
					),
					"createIndexes",
				)
			}

			if len(index.Key) == 1 && index.Key[0].Field == "_id" {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrInvalidIndexSpecificationOption,
					fmt.Sprintf("The field 'unique' is not valid for an _id index specification. "+
						"Specification: { key: %s, name: \"%s\", unique: true, v: 2 }",
						types.FormatAnyValue(must.NotFail(indexDoc.Get("key"))), index.Name,
					),
					"createIndexes",
				)
			}

			index.Unique = pointer.ToBool(true)

		case "background":
			// ignore deprecated options

		case "sparse", "partialFilterExpression", "expireAfterSeconds", "hidden", "storageEngine",
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
	defer keyIter.Close()

	duplicateChecker := make(map[string]struct{}, keyDoc.Len())

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

		if _, ok := duplicateChecker[field]; ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"Error in specification %s, the field %q appears multiple times",
					types.FormatAnyValue(keyDoc), field,
				),
				"createIndexes",
			)
		}

		duplicateChecker[field] = struct{}{}

		var orderParam int64

		if orderParam, err = commonparams.GetWholeNumberParam(order); err != nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index key value %q is not implemented yet", order),
				"createIndexes",
			)
		}

		var indexOrder types.SortType

		switch orderParam {
		case 1:
			indexOrder = types.Ascending
		case -1:
			indexOrder = types.Descending
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

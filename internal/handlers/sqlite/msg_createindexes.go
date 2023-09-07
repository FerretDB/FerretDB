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

package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCreateIndexes implements HandlerInterface.
func (h *Handler) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid database specified '%s'", dbName)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
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
			fmt.Sprintf(
				"BSON field 'createIndexes.indexes' is the wrong type '%s', expected type 'array'",
				commonparams.AliasFromType(v),
			),
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

	params, err := processIndexesArray(command, idxArr)
	if err != nil {
		return nil, err
	}

	_, err = c.CreateIndexes(ctx, params)
	if err != nil {
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

// processIndexesArray processes the given array of indexes and returns a backends.CreateIndexesParams.
func processIndexesArray(command string, indexesArray *types.Array) (*backends.CreateIndexesParams, error) {
	iter := indexesArray.Iterator()
	defer iter.Close()

	params := backends.CreateIndexesParams{
		Indexes: make([]backends.IndexInfo, indexesArray.Len()),
	}

	for {
		var key, val any
		key, val, err := iter.Next()

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			return &params, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		indexDoc, ok := val.(*types.Document)
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'createIndexes.indexes.%d' is the wrong type '%s', expected type 'object'",
					key,
					commonparams.AliasFromType(val),
				),
				command,
			)
		}

		indexInfo, err := processIndex(command, indexDoc)
		if err != nil {
			return nil, err
		}

		params.Indexes[key.(int)] = *indexInfo
	}
}

// processIndex processes the given index document and returns backends.IndexInfo.
func processIndex(command string, indexDoc *types.Document) (*backends.IndexInfo, error) {
	var index backends.IndexInfo

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
					command,
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

		index.Key, err = processIndexKey(command, keyDoc)
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
			v := must.NotFail(indexDoc.Get("unique"))

			unique, ok := v.(bool)
			if !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"Error in specification { key: %s, name: \"%s\", unique: %s } "+
							":: caused by :: "+
							"The field 'unique' has value unique: %[3]s, which is not convertible to bool",
						types.FormatAnyValue(must.NotFail(indexDoc.Get("key"))),
						index.Name, types.FormatAnyValue(v),
					),
					command,
				)
			}

			if len(index.Key) == 1 && index.Key[0].Field == "_id" {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrInvalidIndexSpecificationOption,
					fmt.Sprintf("The field 'unique' is not valid for an _id index specification. "+
						"Specification: { key: %s, name: \"%s\", unique: true, v: 2 }",
						types.FormatAnyValue(must.NotFail(indexDoc.Get("key"))), index.Name,
					),
					command,
				)
			}

			if unique {
				index.Unique = true
			}

		case "background":
			// ignore deprecated options

		case "sparse", "partialFilterExpression", "expireAfterSeconds", "hidden", "storageEngine",
			"weights", "default_language", "language_override", "textIndexVersion", "2dsphereIndexVersion",
			"bits", "min", "max", "bucketSize", "collation", "wildcardProjection":
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index option %q is not implemented yet", opt),
				command,
			)

		default:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("Index option %q is unknown", opt),
				command,
			)
		}
	}
}

// processIndexKey processes the document containing the index key (set of "field-order" pairs).
func processIndexKey(command string, keyDoc *types.Document) ([]backends.IndexKeyPair, error) {
	res := make([]backends.IndexKeyPair, 0, keyDoc.Len())

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
				command,
			)
		}

		duplicateChecker[field] = struct{}{}

		var orderParam int64

		if orderParam, err = commonparams.GetWholeNumberParam(order); err != nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrIndexNotFound,
				fmt.Sprintf("can't find index with key: { %s: \"%s\" }", field, order),
				command,
			)
		}

		var descending bool

		switch orderParam {
		case 1:
			descending = false
		case -1:
			descending = true
		default:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("Index key value %q is not implemented yet", orderParam),
				command,
			)
		}

		res = append(res, backends.IndexKeyPair{
			Field:      field,
			Descending: descending,
		})
	}
}

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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCreateIndexes implements HandlerInterface.
func (h *Handler) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	/*document, err := msg.Document()
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

	iter := idxArr.Iterator()
	defer iter.Close()

	params := backends.CreateIndexesParams{
		Indexes: make([]backends.IndexInfo, 0, idxArr.Len()),
	}

	for {
		i, idx, err := iter.Next()
		if err != nil {
			//
		}

		// idx: document

		params.Indexes[i] = backends.IndexInfo{
			//	Name:   idx.Name,
			//	Unique: idx.Unique,
			//	Key:    make([]backends.IndexKeyPair, 0, len(idx.Key)),
		}
	}

	err = c.CreateIndexes(ctx, &params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	*/
	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

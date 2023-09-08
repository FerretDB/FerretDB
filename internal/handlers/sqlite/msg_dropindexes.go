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

// MsgDropIndexes implements HandlerInterface.
func (h *Handler) MsgDropIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}

	if db == nil {
		msg := fmt.Sprintf("ns not found %s.%s", dbName, collection)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceNotFound, msg, command)
	}

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}

	if c == nil {
		msg := fmt.Sprintf("ns not found %s.%s", dbName, collection)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceNotFound, msg, command)
	}

	options, err := processDropIndexOptions(command, document)
	if err != nil {
		return nil, err
	}

	res, err := c.DropIndexes(ctx, options)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	replyDoc := must.NotFail(types.NewDocument(
		"nIndexesWas", res.NIndexesWas,
	))

	if res.Msg != "" {
		replyDoc.Set("msg", res.Msg)
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

// processDropIndexOptions parses index doc and returns backends.DropIndexesParams.
func processDropIndexOptions(command string, doc *types.Document) (*backends.DropIndexesParams, error) { //nolint:lll // for readability
	var params backends.DropIndexesParams

	v, err := doc.Get("index")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'dropIndexes.index' is missing but a required field",
			command,
		)
	}

	switch v := v.(type) {
	case *types.Document:
		// Index specification (key) is provided to drop a specific index.
		//	var indexKey []backends.IndexKeyPair

		//	indexKey, err = processIndexKey(v)
		//	if err != nil {
		//		return nil, lazyerrors.Error(err)
		//	}
	case *types.Array:
		// List of index names is provided to drop multiple indexes.
		iter := v.Iterator()

		defer iter.Close() // it's safe to defer here as the iterators reads everything

		for {
			var val any
			_, val, err = iter.Next()

			switch {
			case err == nil:
				// do nothing
			case errors.Is(err, iterator.ErrIteratorDone):
				return &params, nil
			default:
				return nil, lazyerrors.Error(err)
			}

			index, ok := val.(string)
			if !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
						commonparams.AliasFromType(v),
					),
					command,
				)
			}
		}
	case string:
		if v == "*" {
			// Drop all indexes except the _id index.
		}

	}

	return nil, commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrTypeMismatch,
		fmt.Sprintf(
			"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
			commonparams.AliasFromType(v),
		),
		command,
	)
}

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
	"strings"

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
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}

	indexValue, err := document.Get("index")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'dropIndexes.index' is missing but a required field",
			command,
		)
	}

	beforeDrop, err := c.ListIndexes(ctx, nil)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
			msg := fmt.Sprintf("ns not found %s.%s", dbName, collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceNotFound, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}

	toDrop, dropAll, err := processDropIndexOptions(command, indexValue, beforeDrop.Indexes)
	if err != nil {
		return nil, err
	}

	_, err = c.DropIndexes(ctx, &backends.DropIndexesParams{Indexes: toDrop})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	replyDoc := must.NotFail(types.NewDocument(
		"nIndexesWas", int32(len(beforeDrop.Indexes)),
	))

	if dropAll {
		replyDoc.Set("msg", "non-_id indexes dropped for collection")
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

// processDropIndexOptions parses index doc and returns the list of indexes to delete
// and true if a flag to drop all indexes except _id_ was set.
func processDropIndexOptions(command string, v any, existing []backends.IndexInfo) ([]string, bool, error) { //nolint:lll // for readability
	switch v := v.(type) {
	case *types.Document:
		// Index specification (key) is provided to drop a specific index.
		var spec []backends.IndexKeyPair

		spec, err := processIndexKey(command, v)
		if err != nil {
			return nil, false, lazyerrors.Error(err)
		}

		if len(spec) == 1 && spec[0].Field == "_id" && !spec[0].Descending {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidOptions, "cannot drop _id index", command,
			)
		}

		for _, index := range existing {
			matches := true

			for i, key := range index.Key {
				if key.Field != spec[i].Field || key.Descending != spec[i].Descending {
					matches = false
					break
				}
			}

			if matches {
				return []string{index.Name}, false, nil
			}
		}

		formattedSpec := make([]string, len(spec))

		for i, key := range spec {
			order := 1
			if key.Descending {
				order = -1
			}

			formattedSpec[i] = fmt.Sprintf("%s: %d", key.Field, order)
		}

		msg := fmt.Sprintf("can't find index with key: { %v }", strings.Join(formattedSpec, ", "))

		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrIndexNotFound, msg, command)

	case *types.Array:
		// List of index names is provided to drop multiple indexes.
		iter := v.Iterator()

		defer iter.Close() // it's safe to defer here as the iterators reads everything

		toDrop := make([]string, 0, v.Len())

		for {
			var val any

			_, val, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					return toDrop, false, nil
				}

				return nil, false, lazyerrors.Error(err)
			}

			index, ok := val.(string)
			if !ok {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
						commonparams.AliasFromType(v),
					),
					command,
				)
			}

			if index == "_id_" {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrInvalidOptions, "cannot drop _id index", command,
				)
			}

			var found bool

			for _, existingIndex := range existing {
				if index == existingIndex.Name {
					toDrop = append(toDrop, index)
					found = true

					break
				}
			}

			if !found {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrIndexNotFound,
					fmt.Sprintf("index not found with name [%s]", index),
					command,
				)
			}
		}

	case string:
		if v == "*" {
			toDrop := make([]string, 0, len(existing))

			// Drop all indexes except the _id index.
			for _, index := range existing {
				if index.Name == "_id_" {
					continue
				}

				toDrop = append(toDrop, index.Name)
			}

			return toDrop, true, nil
		}

		if v == "_id_" {
			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidOptions, "cannot drop _id index", command,
			)
		}

		return []string{v}, false, nil
	}

	return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
		commonerrors.ErrTypeMismatch,
		fmt.Sprintf(
			"BSON field 'dropIndexes.index' is the wrong type '%s', expected types '[string, object]'",
			commonparams.AliasFromType(v),
		),
		command,
	)
}

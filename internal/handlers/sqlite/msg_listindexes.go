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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListIndexes implements HandlerInterface.
func (h *Handler) MsgListIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	res, err := c.ListIndexes(ctx, nil)
	if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseDoesNotExist) ||
		backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
		msg := fmt.Sprintf("ns does not exist: %s.%s", dbName, collection)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceNotFound, msg, document.Command())
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(res.Indexes))

	for _, index := range res.Indexes {
		indexKey := must.NotFail(types.NewDocument())

		for _, key := range index.Key {
			indexKey.Set(key.Field, int32(key.Order))
		}

		indexDoc := must.NotFail(types.NewDocument(
			"v", int32(2), // for compatibility, the meaning of this field is not documented
			"key", indexKey,
			"name", index.Name,
		))

		// only non-default unique indexes should have unique field in the response
		if index.Unique != nil && *index.Unique && index.Name != "_id_" {
			indexDoc.Set("unique", *index.Unique)
		}

		firstBatch.Append(indexDoc)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", fmt.Sprintf("%s.%s", dbName, collection),
				"firstBatch", firstBatch,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

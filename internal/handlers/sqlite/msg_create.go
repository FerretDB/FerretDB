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

// MsgCreate implements HandlerInterface.
func (h *Handler) MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"timeseries",
		"expireAfterSeconds",
		"size", // TODO https://github.com/FerretDB/FerretDB/issues/3458
		"max",  // TODO https://github.com/FerretDB/FerretDB/issues/3458
		"validator",
		"validationLevel",
		"validationAction",
		"viewOn",
		"pipeline",
		"collation",
	}
	if err = common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3458
	if err = common.UnimplementedNonDefault(document, "capped", func(v any) bool {
		b, ok := v.(bool)
		return ok && !b
	}); err != nil {
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

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collectionName, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collectionName)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "create")
		}

		return nil, lazyerrors.Error(err)
	}

	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
		Name: collectionName,
	})

	switch {
	case err == nil:
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"ok", float64(1),
			))},
		}))

		return &reply, nil

	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid):
		msg := fmt.Sprintf("Invalid collection name: %s", collectionName)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "create")

	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionAlreadyExists):
		msg := fmt.Sprintf("Collection %s.%s already exists.", dbName, collectionName)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceExists, msg, "create")

	default:
		return nil, lazyerrors.Error(err)
	}
}

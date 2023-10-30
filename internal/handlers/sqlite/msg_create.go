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
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
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

	params := backends.CreateCollectionParams{
		Name: collectionName,
	}

	var capped bool
	if v, _ := document.Get("capped"); v != nil {
		capped, err = commonparams.GetBoolOptionalParam("capped", v)
		if err != nil {
			return nil, err
		}
	}

	if capped {
		if !h.EnableOplog {
			return nil, common.Unimplemented(document, "capped")
		}

		size, _ := document.Get("size")
		if _, ok := size.(types.NullType); size == nil || ok {
			msg := "the 'size' field is required when 'capped' is true"
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidOptions, msg, "create")
		}

		params.CappedSize, err = commonparams.GetValidatedNumberParamWithMinValue(document.Command(), "size", size, 1)
		if err != nil {
			return nil, err
		}

		if params.CappedSize%256 != 0 {
			params.CappedSize = (params.CappedSize/256 + 1) * 256
		}

		if max, _ := document.Get("max"); max != nil {
			params.CappedDocuments, err = commonparams.GetValidatedNumberParamWithMinValue(document.Command(), "max", max, 0)
			if err != nil {
				return nil, err
			}
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collectionName)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "create")
		}

		return nil, lazyerrors.Error(err)
	}

	err = db.CreateCollection(ctx, &params)

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

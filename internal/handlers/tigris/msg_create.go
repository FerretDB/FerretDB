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

package tigris

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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
		"capped",
		"timeseries",
		"expireAfterSeconds",
		"size",
		"max",
		"validationLevel",
		"validationAction",
		"viewOn",
		"pipeline",
		"collation",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
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

	var db, collection string

	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	// Validator is required for Tigris as we always need to set schema to create a collection.
	schema, err := getJSONSchema(document)
	if err != nil {
		return nil, err
	}

	b := must.NotFail(json.Marshal(schema))

	created, err := h.db.CreateCollectionIfNotExist(ctx, db, collection, b)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if tigrisdb.IsInvalidArgument(err) {
			return nil, common.NewCommandError(common.ErrBadValue, err)
		}

		// Tigris returns Code_ABORTED if concurrent create collection request is detected.
		if tigrisdb.IsAborted(err) {
			msg := fmt.Sprintf("Collection %s.%s already exists.", db, collection)
			return nil, common.NewCommandErrorMsg(common.ErrNamespaceExists, msg)
		}

		return nil, lazyerrors.Error(err)
	default:
		return nil, lazyerrors.Error(err)
	}

	if !created {
		msg := fmt.Sprintf("Collection %s.%s already exists.", db, collection)
		return nil, common.NewCommandErrorMsg(common.ErrNamespaceExists, msg)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

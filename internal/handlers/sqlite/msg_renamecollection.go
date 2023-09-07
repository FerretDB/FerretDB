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

// MsgRenameCollection implements HandlerInterface.
func (h *Handler) MsgRenameCollection(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var err error

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// implement dropTarget param
	// TODO https://github.com/FerretDB/FerretDB/issues/2565
	if err = common.UnimplementedNonDefault(document, "dropTarget", func(v any) bool {
		b, ok := v.(bool)
		return ok && !b
	}); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"writeConcern",
		"comment",
	}
	common.Ignored(document, h.L, ignoredFields...)

	command := document.Command()

	oldName, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		from, _ := document.Get(command)

		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", commonparams.AliasFromType(from)),
			command,
		)
	}

	newName, err := common.GetRequiredParam[string](document, "to")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"'to' must be of type String",
			command,
		)
	}

	oldDBName, oldCName, err := commonparams.SplitNamespace(oldName, command)
	if err != nil {
		return nil, err
	}

	newDBName, newCName, err := commonparams.SplitNamespace(newName, command)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid target namespace: %s", newName),
			command,
		)
	}

	// support cross-database rename
	// TODO https://github.com/FerretDB/FerretDB/issues/2563
	if oldDBName != newDBName {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"Command renameCollection does not support cross-database rename",
			command,
		)
	}

	if oldCName == newCName {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			"Can't rename a collection to itself",
			command,
		)
	}

	db, err := h.b.Database(oldDBName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s'", oldName)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, command)
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	err = db.RenameCollection(ctx, &backends.RenameCollectionParams{
		OldName: oldCName,
		NewName: newCName,
	})

	switch {
	case err == nil:
	// do nothing
	case backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseDoesNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("Database %s does not exist or is drop pending", oldDBName),
			command,
		)
	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionAlreadyExists):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceExists,
			"target namespace exists",
			command,
		)
	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("Source collection %s does not exist", oldName),
			command,
		)
	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			fmt.Sprintf("error with target namespace: Invalid collection name: %s", newCName),
			command,
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

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
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgRenameCollection implements HandlerInterface.
func (h *Handler) MsgRenameCollection(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var err error

	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO Implement dropTarget param: https://github.com/FerretDB/FerretDB/issues/2565
	if err = common.Unimplemented(document, "dropTarget"); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"writeConcern",
		"comment",
	}
	common.Ignored(document, h.L, ignoredFields...)

	command := document.Command()

	namespaceFrom, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	namespaceTo, err := common.GetRequiredParam[string](document, "to")
	if err != nil {
		return nil, err
	}

	dbFrom, collectionFrom, err := extractFromNamespace(namespaceFrom)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s'", namespaceFrom),
			command,
		)
	}

	dbTo, collectionTo, err := extractFromNamespace(namespaceTo)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid target namespace: %s", namespaceTo),
			command,
		)
	}

	// TODO Support cross-database rename: https://github.com/FerretDB/FerretDB/issues/2563
	if dbFrom != dbTo {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"Command renameCollection does not support cross-database rename",
			command,
		)
	}

	if collectionFrom == collectionTo {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			"Can't rename a collection to itself",
			command,
		)
	}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.RenameCollection(ctx, tx, dbFrom, collectionFrom, collectionTo)
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgdb.ErrAlreadyExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceExists,
			"target namespace exists",
			command,
		)
	case errors.Is(err, pgdb.ErrTableNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("Source collection %s does not exist", namespaceFrom),
			command,
		)
	case errors.Is(err, pgdb.ErrInvalidCollectionName):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			fmt.Sprintf("Invalid collection name: '%s'", namespaceTo),
			command,
		)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// extractFromNamespace returns the database and collection name from a given namespace.
//
// The namespace must be in the format of "database.collection".
// If the namespace is invalid, an error is returned
func extractFromNamespace(namespace string) (string, string, error) {
	split := strings.Split(namespace, ".")

	if len(split) != 2 || split[0] == "" || split[1] == "" {
		return "", "", errors.New("invalid namespace")
	}

	return split[0], split[1], nil
}

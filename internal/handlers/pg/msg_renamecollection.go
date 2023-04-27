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

	if err = common.Unimplemented(document, "writeConcern", "comment", "dropTarget"); err != nil {
		return nil, err
	}

	sourceNamespace, err := common.GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
	}

	sourceDB, collection, err := extractFromNamespace(sourceNamespace)
	if err != nil {
		return nil, err
	}

	targetField, err := document.Get("to")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'renameCollection.to' is missing but a required field",
			document.Command(),
		)
	}

	targetNamespace, ok := targetField.(string)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"BSON field 'renameCollection.to' is the wrong type 'bool', expected type 'string'",
			document.Command(),
		)
	}

	// we assume we cannot move a collection between databases, yet.
	targetDB, targetCollection, err := extractFromNamespace(targetNamespace)
	if err != nil {
		return nil, err
	}

	if targetDB != sourceDB {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			"Can't rename a collection to another database",
			document.Command(),
		)
	}

	if sourceNamespace == targetNamespace {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			"Can't rename a collection to itself",
			document.Command(),
		)
	}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.RenameCollection(ctx, tx, sourceDB, collection, targetCollection)
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgdb.ErrAlreadyExist):
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrNamespaceExists,
			"target namespace exists",
		)
	case errors.Is(err, pgdb.ErrTableNotExist):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("Source collection %s does not exist", sourceNamespace),
			document.Command(),
		)
	case errors.Is(err, pgdb.ErrInvalidCollectionName):
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrIllegalOperation,
			fmt.Sprintf(
				"error with target namespace: Fully qualified namespace is too long. Namespace: %s Max: 255",
				targetNamespace,
			),
			document.Command(),
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
func extractFromNamespace(namespace string) (string, string, error) {
	split := strings.Split(namespace, ".")

	// we assume that the given namespace contains a single dot.
	if len(split) != 2 {
		return "", "", commonerrors.NewCommandErrorMsg(
			commonerrors.ErrInvalidNamespace,
			"Invalid namespace specified "+namespace,
		)
	}

	return split[0], split[1], nil
}

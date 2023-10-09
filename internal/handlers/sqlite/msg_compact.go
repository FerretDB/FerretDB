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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func (h *Handler) MsgCompact(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidNamespace,
				fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection),
				command,
			)
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidNamespace,
				fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection),
				command,
			)
		}

		return nil, lazyerrors.Error(err)
	}

	// TODO: handle compact parameters

	// TODO: call collstats before

	// TODO: call compact
	_, err = c.Compact(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO: cal
	// call collstats after

	return nil, nil
}

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
	"fmt"
	"os"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/jackc/pgx/v4"
)

// MsgExplain implements HandlerInterface.
func (h *Handler) MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	// todo extract for each pg command a part which is before Query Document
	// make map [command] func of it
	// run it before query document regarding a command
	// in query document "simple explain analyse"
	// it's not beautiful.
	// what if only for selects?
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	verbosity := "allPlansExecution"
	if verbosity, err = common.GetOptionalParam(document, "verbosity", verbosity); err != nil {
		return nil, err
	}
	if verbosity == "executionStats" || verbosity == "serverParameters" {
		err = fmt.Errorf("verbosity: support for value %q is not implemented yet", verbosity)
		return nil, common.NewError(common.ErrNotImplemented, err)
	}

	var sp pgdb.SQLParam
	if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	commandParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	fmt.Printf("%#v", commandParam)
	if _, ok := commandParam.(*types.Document); !ok {
		return nil, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("has invalid type %s", common.AliasFromType(commandParam)),
		)
	}

	os.Exit(1)

	resDocs := make([]*types.Document, 0, 16)
	err = h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		fetchedChan, err := h.pgPool.QueryDocuments(ctx, tx, sp)
		if err != nil {
			return err
		}
		defer func() {
			// Drain the channel to prevent leaking goroutines.
			// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
			for range fetchedChan {
			}
		}()
		for fetchedItem := range fetchedChan {
			if fetchedItem.Err != nil {
				return fetchedItem.Err
			}
			resDocs = append(resDocs, fetchedItem.Docs...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

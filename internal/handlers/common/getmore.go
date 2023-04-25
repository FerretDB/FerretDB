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

package common

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// GetMore is a part of common implementation of the getMore command.
func GetMore(ctx context.Context, msg *wire.OpMsg, registry *cursor.Registry) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db, err := commonparams.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := commonparams.GetRequiredParam[string](document, "collection")
	if err != nil {
		return nil, err
	}

	cursorID, err := commonparams.GetRequiredParam[int64](document, document.Command())
	if err != nil {
		return nil, err
	}

	// TODO maxTimeMS, comment

	username, _ := conninfo.Get(ctx).Auth()

	cursor := registry.Cursor(username, cursorID)
	if cursor == nil {
		return nil, lazyerrors.Errorf("no cursor %d", cursorID)
	}

	// TODO this logic should be tested
	batchSize, _ := commonparams.GetOptionalParam(document, "batchSize", cursor.BatchSize)
	if batchSize < 0 {
		batchSize = 101
	}

	if cursor.DB != db || cursor.Collection != collection {
		return nil, lazyerrors.Errorf("cursor %d is for %s.%s, not %s.%s", cursorID, cursor.DB, cursor.Collection, db, collection)
	}

	resDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor.Iter), int(batchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	nextBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		nextBatch.Append(doc)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"nextBatch", nextBatch,
				"id", cursorID,
				"ns", db+"."+collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

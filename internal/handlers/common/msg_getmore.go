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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetMore is a common implementation of the `getMore` command.
func MsgGetMore(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db, err := GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := GetRequiredParam[string](document, "collection")
	if err != nil {
		return nil, err
	}

	cursorID, err := GetRequiredParam[int64](document, document.Command())
	if err != nil {
		return nil, err
	}

	batchSize, err := GetRequiredParam[int32](document, "batchSize")
	if err != nil {
		return nil, err
	}

	// TODO maxTimeMS, comment

	cursor := conninfo.Get(ctx).Cursor(cursorID)
	if cursor == nil {
		return nil, lazyerrors.Errorf("no cursor %d", cursorID)
	}

	if cursor.DB != db || cursor.Collection != collection {
		return nil, lazyerrors.Errorf("cursor %d is for %s.%s, not %s.%s", cursorID, cursor.DB, cursor.Collection, db, collection)
	}

	if cursor.BatchSize != batchSize {
		return nil, lazyerrors.Errorf("cursor %d has batch size %d, not %d", cursorID, cursor.BatchSize, batchSize)
	}

	resDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor.Iter), int(cursor.BatchSize))
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

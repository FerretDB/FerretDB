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
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
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

	if err = Unimplemented(document, "comment", "maxTimeMS"); err != nil {
		return nil, err
	}

	db, err := GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := GetRequiredParam[string](document, "collection")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrBadValue, `required parameter "collection" is missing`)
	}

	cursorIDValue, err := document.Get("getMore")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrBadValue, `required parameter "getMore" is missing`)
	}

	var cursorID int64
	var ok bool

	if cursorID, ok = cursorIDValue.(int64); !ok {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				`BSON field 'getMore.getMore' is the wrong type '%s', expected type 'long'`,
				AliasFromType(cursorIDValue),
			),
		)
	}

	if cursorID <= 0 {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
		)
	}

	batchSize, err := getBatchSize(document)
	if err != nil {
		return nil, err
	}

	connInfo := conninfo.Get(ctx)

	cursor := connInfo.Cursor(cursorID)
	if cursor == nil {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
		)
	}

	resDocs := types.MakeArray(0)
	iter := cursor.Iter

	var done bool

	for i := 0; i < int(batchSize); i++ {
		var doc any

		_, doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				done = true
				break
			}

			return nil, lazyerrors.Error(err)
		}

		var matches bool

		matches, err = FilterDocument(document, cursor.Filter)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs.Append(doc)
	}

	if done {
		cursorID = 0
	}

	var reply wire.OpMsg

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"nextBatch", resDocs,
				"id", cursorID,
				"ns", db+"."+collection,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// getBatchSize returns the batch size from the document.
func getBatchSize(doc *types.Document) (int64, error) {
	batchSizeValue, err := doc.Get("batchSize")
	if err != nil {
		return 0, nil
	}

	batchSize, err := GetWholeNumberParam(batchSizeValue)
	if err != nil {
		if errors.Is(err, errUnexpectedType) {
			return 0, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'batchSize' is the wrong type '%s', expected type 'long'",
					AliasFromType(batchSizeValue),
				),
			)
		}
	}

	if batchSize < 0 {
		return 0, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrBatchSizeNegative,
			"BSON field 'batchSize' value must be >= 0, actual value '-1'",
		)
	}

	return batchSize, nil
}

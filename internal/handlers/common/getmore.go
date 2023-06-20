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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
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

	db, err := GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	// TODO: Use ExtractParam https://github.com/FerretDB/FerretDB/issues/2859
	v, err := document.Get("collection")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'getMore.collection' is missing but a required field",
			document.Command(),
		)
	}

	collection, ok := v.(string)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				"BSON field 'getMore.collection' is the wrong type '%s', expected type 'string'",
				commonparams.AliasFromType(v),
			),
			document.Command(),
		)
	}

	if collection == "" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			"Collection names cannot be empty",
			document.Command(),
		)
	}

	cursorID, err := GetRequiredParam[int64](document, document.Command())
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
			document.Command(),
		)
	}

	// TODO maxTimeMS, comment

	username, _ := conninfo.Get(ctx).Auth()

	// TODO: Use ExtractParam https://github.com/FerretDB/FerretDB/issues/2859
	cursor := registry.Cursor(username, cursorID)
	if cursor == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
			document.Command(),
		)
	}

	v, err = document.Get("batchSize")
	if err != nil || types.Compare(v, int32(0)) == types.Equal {
		// TODO: Use 16MB batchSize limit https://github.com/FerretDB/FerretDB/issues/2824
		// unlimited default batchSize is used for unset batchSize and zero values,
		// set 250 assuming it is small enough not to crash FerretDB.
		v = int32(250)
	}

	batchSize, err := commonparams.GetValidatedNumberParamWithMinValue(document.Command(), "batchSize", v, 0)
	if err != nil {
		return nil, err
	}

	if cursor.DB != db || cursor.Collection != collection {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrUnauthorized,
			fmt.Sprintf("Requested getMore on namespace '%s.%s', but cursor belongs to a different namespace %s.%s",
				db,
				collection,
				cursor.DB,
				cursor.Collection,
			),
			document.Command(),
		)
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

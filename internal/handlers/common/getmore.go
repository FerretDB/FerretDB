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
	"math"

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

	// Use ExtractParam.
	// TODO https://github.com/FerretDB/FerretDB/issues/2859
	v, _ := document.Get("collection")
	if v == nil {
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

	// TODO https://github.com/FerretDB/FerretDB/issues/2984
	v, _ = document.Get("maxTimeMS")
	if v == nil {
		v = int64(0)
	}

	// cannot use other existing commonparams function, they return different error codes
	maxTimeMS, err := commonparams.GetWholeNumberParam(v)
	if err != nil {
		switch {
		case errors.Is(err, commonparams.ErrUnexpectedType):
			if _, ok = v.(types.NullType); ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"maxTimeMS must be a number",
					document.Command(),
				)
			}

			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field 'getMore.maxTimeMS' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
					commonparams.AliasFromType(v),
				),
				document.Command(),
			)
		case errors.Is(err, commonparams.ErrNotWholeNumber):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"maxTimeMS has non-integral value",
				document.Command(),
			)
		case errors.Is(err, commonparams.ErrLongExceededPositive) || errors.Is(err, commonparams.ErrLongExceededNegative):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("%s value for maxTimeMS is out of range", types.FormatAnyValue(v)),
				document.Command(),
			)
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	if maxTimeMS < int64(0) || maxTimeMS > math.MaxInt32 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("%v value for maxTimeMS is out of range", v),
			document.Command(),
		)
	}

	// Handle comment.
	// TODO https://github.com/FerretDB/FerretDB/issues/2986

	username, _ := conninfo.Get(ctx).Auth()

	// Use ExtractParam.
	// TODO https://github.com/FerretDB/FerretDB/issues/2859
	cursor := registry.Get(cursorID)
	if cursor == nil || cursor.Username != username {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
			document.Command(),
		)
	}

	v, _ = document.Get("batchSize")
	if v == nil || types.Compare(v, int32(0)) == types.Equal {
		// Use 16MB batchSize limit.
		// Unlimited default batchSize is used for missing batchSize and zero values,
		// set 250 assuming it is small enough not to crash FerretDB.
		// TODO https://github.com/FerretDB/FerretDB/issues/2824
		v = int32(250)
	}

	batchSize, err := commonparams.GetValidatedNumberParamWithMinValue(document.Command(), "batchSize", v, 0)
	if err != nil {
		return nil, err
	}

	if cursor.DB != db || cursor.Collection != collection {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrUnauthorized,
			fmt.Sprintf(
				"Requested getMore on namespace '%s.%s', but cursor belongs to a different namespace %s.%s",
				db,
				collection,
				cursor.DB,
				cursor.Collection,
			),
			document.Command(),
		)
	}

	resDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor), int(batchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	nextBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		nextBatch.Append(doc)
	}

	if nextBatch.Len() < int(batchSize) {
		// Cursor ID 0 lets the client know that there are no more results.
		// Cursor is already closed and removed from the registry by this point.
		cursorID = 0
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

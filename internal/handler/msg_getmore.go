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

package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetMore implements `getMore` command.
func (h *Handler) MsgGetMore(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	// Use ExtractParam.
	// TODO https://github.com/FerretDB/FerretDB/issues/2859
	v, _ := document.Get("collection")
	if v == nil {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrMissingField,
			"BSON field 'getMore.collection' is missing but a required field",
			document.Command(),
		)
	}

	collection, ok := v.(string)
	if !ok {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrTypeMismatch,
			fmt.Sprintf(
				"BSON field 'getMore.collection' is the wrong type '%s', expected type 'string'",
				handlerparams.AliasFromType(v),
			),
			document.Command(),
		)
	}

	if collection == "" {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrInvalidNamespace,
			"Collection names cannot be empty",
			document.Command(),
		)
	}

	cursorID, err := common.GetRequiredParam[int64](document, document.Command())
	if err != nil {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrTypeMismatch,
			"BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
			document.Command(),
		)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2984
	v, _ = document.Get("maxTimeMS")
	if v == nil {
		v = int64(1000)
	}

	// cannot use other existing handlerparams function, they return different error codes
	maxTimeMS, err := handlerparams.GetWholeNumberParam(v)
	if err != nil {
		switch {
		case errors.Is(err, handlerparams.ErrUnexpectedType):
			if _, ok = v.(types.NullType); ok {
				return nil, handlererrors.NewCommandErrorMsgWithArgument(
					handlererrors.ErrBadValue,
					"maxTimeMS must be a number",
					document.Command(),
				)
			}

			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field 'getMore.maxTimeMS' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
					handlerparams.AliasFromType(v),
				),
				document.Command(),
			)
		case errors.Is(err, handlerparams.ErrNotWholeNumber):
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				"maxTimeMS has non-integral value",
				document.Command(),
			)
		case errors.Is(err, handlerparams.ErrLongExceededPositive) || errors.Is(err, handlerparams.ErrLongExceededNegative):
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				fmt.Sprintf("%s value for maxTimeMS is out of range", types.FormatAnyValue(v)),
				document.Command(),
			)
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	if maxTimeMS < int64(0) || maxTimeMS > math.MaxInt32 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			fmt.Sprintf("%v value for maxTimeMS is out of range", v),
			document.Command(),
		)
	}

	// Handle comment.
	// TODO https://github.com/FerretDB/FerretDB/issues/2986

	username := conninfo.Get(ctx).Username()

	// Use ExtractParam.
	// TODO https://github.com/FerretDB/FerretDB/issues/2859
	c := h.cursors.Get(cursorID)
	if c == nil || c.Username != username {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrCursorNotFound,
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

	batchSize, err := handlerparams.GetValidatedNumberParamWithMinValue(document.Command(), "batchSize", v, 0)
	if err != nil {
		return nil, err
	}

	if c.DB != db || c.Collection != collection {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrUnauthorized,
			fmt.Sprintf(
				"Requested getMore on namespace '%s.%s', but cursor belongs to a different namespace %s.%s",
				db,
				collection,
				c.DB,
				c.Collection,
			),
			document.Command(),
		)
	}

	nextBatch, err := h.makeNextBatch(c, batchSize)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	switch c.Type {
	case cursor.Normal:
		if nextBatch.Len() < int(batchSize) {
			// The cursor is already closed and removed;
			// let the client know that there are no more results.
			cursorID = 0
		}

	case cursor.Tailable:
		if nextBatch.Len() == 0 {
			// The previous iterator is already closed there.

			data := c.Data.(*findCursorData)

			queryRes, err := data.coll.Query(ctx, data.qp)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			closer := iterator.NewMultiCloser()

			iter, err := h.makeFindIter(queryRes.Iter, closer, data.findParams)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err = c.Reset(iter); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if nextBatch.Len() == 0 {
				nextBatch, err = h.makeNextBatch(c, batchSize)
				if err != nil {
					return nil, lazyerrors.Error(err)
				}
			}
		}

	case cursor.TailableAwait:
		if nextBatch.Len() == 0 {
			nextBatch, err = h.awaitData(ctx, &awaitDataParams{
				cursor:    c,
				batchSize: batchSize,
				maxTimeMS: maxTimeMS,
			})

			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

	default:
		panic(fmt.Sprintf("unknown cursor type %s", c.Type))
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

// makeNextBatch returns the next batch of documents from the cursor.
func (h *Handler) makeNextBatch(c *cursor.Cursor, batchSize int64) (*types.Array, error) {
	docs, err := iterator.ConsumeValuesN(c, int(batchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.L.Debug(
		"Got next batch", zap.Int64("cursor_id", c.ID), zap.Stringer("type", c.Type),
		zap.Int("count", len(docs)), zap.Int64("batch_size", batchSize),
	)

	nextBatch := types.MakeArray(len(docs))
	for _, doc := range docs {
		nextBatch.Append(doc)
	}

	return nextBatch, nil
}

// awaitDataParams contains parameters that can be passed to awaitData function.
type awaitDataParams struct {
	cursor    *cursor.Cursor
	maxTimeMS int64
	batchSize int64
}

// awaitData stops the goroutine, and waits for a new data for the cursor.
// If there's a new document, or the maxTimeMS have passed it returns the nextBatch.
func (h *Handler) awaitData(ctx context.Context, params *awaitDataParams) (resBatch *types.Array, err error) {
	c := params.cursor
	data := c.Data.(*findCursorData)

	closer := iterator.NewMultiCloser()

	sleepDur := time.Duration(params.maxTimeMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, sleepDur)

	defer func() {
		c.Close()
		cancel()

		if err == nil {
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			err = nil
			return
		}

		err = lazyerrors.Error(err)
		resBatch = types.MakeArray(0)
	}()

	for {
		var queryRes *backends.QueryResult

		queryRes, err = data.coll.Query(ctx, data.qp)
		if err != nil {
			return
		}

		var iter types.DocumentsIterator

		iter, err = h.makeFindIter(queryRes.Iter, closer, data.findParams)
		if err != nil {
			return
		}

		if err = c.Reset(iter); err != nil {
			return
		}

		if resBatch.Len() != 0 {
			return
		}

		resBatch, err = h.makeNextBatch(c, params.batchSize)
		if err != nil {
			return
		}

		if params.maxTimeMS > 10 {
			ctxutil.Sleep(ctx, 10*time.Millisecond)
		}
	}
}

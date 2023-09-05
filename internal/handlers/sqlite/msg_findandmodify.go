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
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// findAndModifyResult represents information about modified document.
type findAndModifyResult struct {
	updateExisting any
	upserted       any
	value          any
	writeErrors    *types.Array
	modified       int32
}

// MsgFindAndModify implements HandlerInterface.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetFindAndModifyParams(document, h.L)
	if err != nil {
		return nil, err
	}

	if params.Update != nil {
		if err = common.ValidateUpdateOperators(document.Command(), params.Update); err != nil {
			return nil, err
		}
	}

	res, err := h.findAndModifyDocument(ctx, params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	lastError := must.NotFail(types.NewDocument(
		"n", res.modified,
	))

	if res.updateExisting != nil {
		lastError.Set("updatedExisting", res.updateExisting)
	}

	if res.upserted != nil {
		lastError.Set("upserted", res.upserted)
	}

	resDoc := must.NotFail(types.NewDocument(
		"lastErrorObject", lastError,
		"value", res.value,
	))

	if res.writeErrors.Len() > 0 {
		resDoc.Set("writeErrors", res.writeErrors)
	}

	resDoc.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	}))

	return &reply, nil
}

// findAndModifyDocument finds and modifies a single document.
func (h *Handler) findAndModifyDocument(ctx context.Context, params *common.FindAndModifyParams) (*findAndModifyResult, error) {
	var modified int32
	var updateExisting, upserted, value any

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "findAndModify")
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "findAndModify")
		}

		return nil, lazyerrors.Error(err)
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))
	defer closer.Close()

	queryRes, err := c.Query(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	closer.Add(queryRes.Iter)

	iter := common.FilterIterator(queryRes.Iter, closer, params.Query)

	iter, err = common.SortIterator(iter, closer, params.Sort)
	if err != nil {
		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrPathContainsEmptyElement,
				"FieldPath field names may not be empty strings.",
				"findAndModify",
			)
		}

		return nil, lazyerrors.Error(err)
	}

	// findAndModify modifies a single document
	iter = common.LimitIterator(iter, closer, 1)

	writeErrors := types.MakeArray(0)

	_, v, err := iter.Next()
	if errors.Is(err, iterator.ErrIteratorDone) {
		value = types.Null

		if params.Upsert {
			doc := params.Update
			if params.HasUpdateOperators {
				doc = must.NotFail(types.NewDocument())
				if _, err = common.UpdateDocument("findAndModify", doc, params.Update); err != nil {
					return nil, err
				}
			}

			upserted, _ = doc.Get("_id")
			if upserted == nil {
				upserted, err = params.Query.Get("_id")
				if err != nil {
					upserted = types.NewObjectID()
				}

				idDoc, ok := upserted.(*types.Document)
				if ok {
					var hasOp bool

					if hasOp, err = common.HasQueryOperator(idDoc); err != nil {
						return nil, err
					}

					if hasOp {
						upserted = types.NewObjectID()
					}
				}

				doc.Set("_id", upserted)
			}

			if err = doc.ValidateData(); err != nil {
				var we *writeError

				if we, err = handleValidationError(err); err != nil {
					return nil, err
				}

				writeErrors.Append(we.Document())
			}

			if _, err = c.InsertAll(ctx, &backends.InsertAllParams{
				Docs: []*types.Document{doc},
			}); err != nil {
				if backends.ErrorCodeIs(err, backends.ErrorCodeInsertDuplicateID) {
					we := &writeError{
						index:  int32(0),
						code:   commonerrors.ErrDuplicateKeyInsert,
						errmsg: fmt.Sprintf(`E11000 duplicate key error collection: %s.%s`, params.DB, params.Collection),
					}
					writeErrors.Append(we.Document())
				}

				return nil, lazyerrors.Error(err)
			}

			if params.ReturnNewDocument {
				value = doc
			}

			modified = 1
		}

		if !params.Remove {
			updateExisting = false
		}

		return &findAndModifyResult{
			modified:       modified,
			updateExisting: updateExisting,
			upserted:       upserted,
			value:          value,
			writeErrors:    writeErrors,
		}, nil
	}

	value = v

	if params.Remove {
		var delRes *backends.DeleteAllResult

		if delRes, err = c.DeleteAll(ctx, &backends.DeleteAllParams{IDs: []any{must.NotFail(v.Get("_id"))}}); err != nil {
			return nil, lazyerrors.Error(err)
		}

		modified = delRes.Deleted
	} else {
		// TODO https://github.com/FerretDB/FerretDB/issues/3040
		doc := params.Update
		if params.HasUpdateOperators {
			doc = v.DeepCopy()
			if _, err = common.UpdateDocument("findAndModify", doc, params.Update); err != nil {
				return nil, err
			}
		}

		id := must.NotFail(v.Get("_id"))
		updateID, _ := doc.Get("_id")
		if updateID == nil {
			doc.Set("_id", id)
		}

		if updateID != nil && updateID != id {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrImmutableField,
				fmt.Sprintf(
					`Plan executor error during findAndModify :: caused `+
						`by :: After applying the update, the (immutable) field `+
						`'_id' was found to have been altered to _id: "%s"`,
					updateID,
				),
				"findAndModify",
			)
		}

		if err = doc.ValidateData(); err != nil {
			var we *writeError

			if we, err = handleValidationError(err); err != nil {
				return nil, err
			}

			writeErrors.Append(we.Document())
		}

		updateRes, err := c.UpdateAll(ctx, &backends.UpdateAllParams{Docs: []*types.Document{doc}})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if params.ReturnNewDocument {
			value = doc
		}

		modified = updateRes.Updated
		updateExisting = true
	}

	return &findAndModifyResult{
		modified:       modified,
		updateExisting: updateExisting,
		upserted:       upserted,
		value:          value,
		writeErrors:    writeErrors,
	}, nil
}

// handleValidationError checks validation error code and returns *writeError.
func handleValidationError(err error) (*writeError, error) {
	var ve *types.ValidationError

	if !errors.As(err, &ve) {
		return nil, lazyerrors.Error(err)
	}

	var code commonerrors.ErrorCode

	switch ve.Code() {
	case types.ErrValidation, types.ErrIDNotFound:
		code = commonerrors.ErrBadValue
	case types.ErrWrongIDType:
		code = commonerrors.ErrInvalidID
	default:
		panic(fmt.Sprintf("Unknown error code: %v", ve.Code()))
	}

	return &writeError{
		index:  int32(0),
		code:   code,
		errmsg: ve.Error(),
	}, nil
}

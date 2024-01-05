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
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// findAndModifyResult represents information about modification made.
type findAndModifyResult struct {
	updateExisting any
	upserted       any
	value          any
	modified       int32
}

// MsgFindAndModify implements `findAndModify` command.
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

	var we *mongo.WriteError
	var resDoc *types.Document

	res, err := h.findAndModifyDocument(ctx, params)
	if err != nil {
		we, err = handleUpdateError(params.DB, params.Collection, err)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		resDoc = must.NotFail(types.NewDocument(
			"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0))),
			"value", types.Null,
			"writeErrors", must.NotFail(types.NewArray(WriteErrorDocument(we))),
		))
	} else {
		lastError := must.NotFail(types.NewDocument(
			"n", res.modified,
		))

		if res.updateExisting != nil {
			lastError.Set("updatedExisting", res.updateExisting)
		}

		if res.upserted != nil {
			lastError.Set("upserted", res.upserted)
		}

		resDoc = must.NotFail(types.NewDocument(
			"lastErrorObject", lastError,
			"value", res.value,
		))
	}

	resDoc.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		resDoc,
	)))

	return &reply, nil
}

// findAndModifyDocument finds and modifies a single document.
// Upon finding a document, if `remove` flag is set that document is removed,
// otherwise it updates the document applying operators if any.
// When no document is found, a document is inserted if `upsert` flag is set.
func (h *Handler) findAndModifyDocument(ctx context.Context, params *common.FindAndModifyParams) (*findAndModifyResult, error) {
	db, err := h.b.Database(params.DB)
	if err != nil {
		// TODO https://github.com/FerretDB/FerretDB/issues/2168
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "findAndModify")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		// TODO https://github.com/FerretDB/FerretDB/issues/2168
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "findAndModify")
		}

		return nil, lazyerrors.Error(err)
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		// TODO https://github.com/FerretDB/FerretDB/issues/2168
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))
	defer closer.Close()

	var qp backends.QueryParams
	if !h.DisablePushdown {
		qp.Filter = params.Query
	}

	queryRes, err := c.Query(ctx, &qp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	closer.Add(queryRes.Iter)

	iter := common.FilterIterator(queryRes.Iter, closer, params.Query)

	iter, err = common.SortIterator(iter, closer, params.Sort)
	if err != nil {
		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrPathContainsEmptyElement,
				"FieldPath field names may not be empty strings.",
				"findAndModify",
			)
		}

		return nil, lazyerrors.Error(err)
	}

	// findAndModify modifies a single document
	iter = common.LimitIterator(iter, closer, 1)

	result := &findAndModifyResult{
		value: types.Null,
	}

	if params.Remove {
		_, doc, err := iter.Next()
		if err != nil && !errors.Is(err, iterator.ErrIteratorDone) {
			return nil, lazyerrors.Error(err)
		}

		if doc != nil {
			if _, err = c.DeleteAll(ctx, &backends.DeleteAllParams{IDs: []any{must.NotFail(doc.Get("_id"))}}); err != nil {
				return nil, lazyerrors.Error(err)
			}
			result.modified = 1
			result.value = doc
		}

		return result, nil
	}

	// handle update and upsert

	update := &common.Update{
		Filter:             params.Query,
		Update:             params.Update,
		Upsert:             params.Upsert,
		HasUpdateOperators: params.HasUpdateOperators,
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2168
	updateRes, err := common.UpdateDocument(ctx, c, "findAndModify", iter, update)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	result.updateExisting = false

	if updateRes.Upserted.Doc != nil {
		doc := updateRes.Upserted.Doc
		result.modified = 1
		result.upserted = must.NotFail(doc.Get("_id"))
		if params.ReturnNewDocument {
			result.value = doc
		}
	} else if updateRes.Matched.Count > 0 {
		result.modified = 1
		result.updateExisting = true
		result.value = updateRes.Matched.Doc
		if params.ReturnNewDocument && updateRes.Modified.Doc != nil {
			result.value = updateRes.Modified.Doc
		}
	}

	return result, nil
}

// handleUpdateError process backend/validation error returned from update operation.
// It returns *mongo.WriteError if updateErr is an expected error from update operation, otherwise it returns error.
func handleUpdateError(db, coll string, updateErr error) (*mongo.WriteError, error) {
	var we *mongo.WriteError
	var be *backends.Error
	var ve *types.ValidationError

	if errors.As(updateErr, &be) && be.Code() == backends.ErrorCodeInsertDuplicateID {
		we = &mongo.WriteError{
			Index:   0,
			Code:    int(handlererrors.ErrDuplicateKeyInsert),
			Message: fmt.Sprintf(`E11000 duplicate key error collection: %s.%s`, db, coll),
		}
	} else if errors.As(updateErr, &ve) {
		we = convertValidationErrToWriteErr(ve)
	} else {
		return nil, lazyerrors.Error(updateErr)
	}

	return we, nil
}

// convertValidationErrToWriteErr converts validation error and returns *mongo.WriteError.
func convertValidationErrToWriteErr(err error) *mongo.WriteError {
	ve, ok := err.(*types.ValidationError)
	if !ok {
		panic(fmt.Sprintf("unexpected error type %T", err))
	}

	var code handlererrors.ErrorCode

	switch ve.Code() {
	case types.ErrValidation, types.ErrIDNotFound:
		code = handlererrors.ErrBadValue
	case types.ErrWrongIDType:
		code = handlererrors.ErrInvalidID
	default:
		panic(fmt.Sprintf("unknown error code: %v", ve.Code()))
	}

	return &mongo.WriteError{
		Index:   0,
		Code:    int(code),
		Message: ve.Error(),
	}
}

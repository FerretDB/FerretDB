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

	var resDoc *types.Document

	res, err := h.findAndModifyDocument(ctx, params)
	if err != nil {
		return nil, handleUpdateError(params.DB, params.Collection, "findAndModify", err)
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

	resDoc = must.NotFail(types.NewDocument(
		"lastErrorObject", lastError,
		"value", res.value,
	))

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
		var doc *types.Document

		_, doc, err = iter.Next()
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

// handleUpdateError coverts backend/validation error returned from update operation
// into CommandError or WriteError based on the command.
func handleUpdateError(db, coll, command string, err error) error {
	var be *backends.Error
	var ve *types.ValidationError

	if errors.As(err, &be) && be.Code() == backends.ErrorCodeInsertDuplicateID {
		err = common.NewUpdateError(
			handlererrors.ErrDuplicateKeyInsert,
			fmt.Sprintf(`E11000 duplicate key error collection: %s.%s`, db, coll),
			command,
		)
	} else if errors.As(err, &ve) {
		err = validationErrToUpdateErr(command, ve)
	}

	return err
}

// validationErrToUpdateErr converts validation error into CommandError or WriteError based on the command.
func validationErrToUpdateErr(command string, ve *types.ValidationError) error {
	var code handlererrors.ErrorCode

	switch ve.Code() {
	case types.ErrValidation, types.ErrIDNotFound:
		code = handlererrors.ErrBadValue
	case types.ErrWrongIDType:
		code = handlererrors.ErrInvalidID
	default:
		panic(fmt.Sprintf("unknown error code: %v", ve.Code()))
	}

	return common.NewUpdateError(code, ve.Error(), command)
}

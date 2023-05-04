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
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// DeleteParams represents parameters for delete operation.
type DeleteParams struct {
	DB, Collection string
	Comment        string
	Deletes        []Delete
	Ordered        bool
}

// Delete represents single delete operation parameters.
type Delete struct {
	Filter  *types.Document
	Comment string
	Limited bool
}

// GetDeleteParams returns parameters for delete operation.
func GetDeleteParams(document *types.Document, l *zap.Logger) (*DeleteParams, error) {
	var err error

	if err = Unimplemented(document, "let"); err != nil {
		return nil, err
	}

	Ignored(document, l, "writeConcern")

	var deletesArray *types.Array
	if deletesArray, err = GetOptionalParam(document, "deletes", deletesArray); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	var db, collection string

	if db, err = GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", commonparams.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	// get comment from options.Delete().SetComment() method
	var comment string
	if comment, err = GetOptionalParam(document, "comment", comment); err != nil {
		return nil, err
	}

	var deletes []Delete

	for i := 0; i < deletesArray.Len(); i++ {
		// get document with filter
		deleteDoc, err := AssertType[*types.Document](must.NotFail(deletesArray.Get(i)))
		if err != nil {
			return nil, err
		}

		deleteFilter, limited, err := prepareDeleteParams(deleteDoc, l)
		if err != nil {
			return nil, err
		}

		// get comment from query, e.g. db.collection.DeleteOne({"_id":"string", "$comment: "test"})
		if comment, err = GetOptionalParam(deleteFilter, "$comment", comment); err != nil {
			return nil, err
		}

		deletes = append(deletes, Delete{
			Filter:  deleteFilter,
			Limited: limited,
			Comment: comment,
		})
	}

	return &DeleteParams{
		DB:         db,
		Collection: collection,
		Ordered:    ordered,
		Comment:    comment,
		Deletes:    deletes,
	}, nil
}

// prepareDeleteParams returns filter and limit parameters for delete operation.
func prepareDeleteParams(deleteDoc *types.Document, l *zap.Logger) (*types.Document, bool, error) {
	var err error

	if err = Unimplemented(deleteDoc, "collation"); err != nil {
		return nil, false, err
	}

	Ignored(deleteDoc, l, "hint")

	// get filter from document
	var filter *types.Document
	if filter, err = GetOptionalParam(deleteDoc, "q", filter); err != nil {
		return nil, false, err
	}

	// TODO use `GetLimitParam`
	// https://github.com/FerretDB/FerretDB/issues/2255
	limitValue, err := deleteDoc.Get("limit")
	if err != nil {
		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'delete.deletes.limit' is missing but a required field",
			"limit",
		)
	}

	var limit int64
	if limit, err = commonparams.GetWholeNumberParam(limitValue); err != nil || limit < 0 || limit > 1 {
		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFailedToParse,
			fmt.Sprintf("The limit field in delete objects must be 0 or 1. Got %v", limitValue),
			"limit",
		)
	}

	return filter, limit == 1, nil
}

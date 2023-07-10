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
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../bin/stringer -linecomment -type UpsertOperation

// UpsertOperation represents operation type of upsert.
type UpsertOperation uint8

const (
	_ UpsertOperation = iota

	// UpsertOperationInsert indicates that upsert is an insert operation.
	UpsertOperationInsert

	// UpsertOperationUpdate indicates that upsert is a update operation.
	UpsertOperationUpdate
)

// FindAndModifyParams represent parameters for the findAndModify command.
type FindAndModifyParams struct {
	DB                string          `ferretdb:"$db"`
	Collection        string          `ferretdb:"collection"`
	Comment           string          `ferretdb:"comment,opt"`
	Query             *types.Document `ferretdb:"query,opt"`
	Sort              *types.Document `ferretdb:"sort,opt"`
	UpdateValue       any             `ferretdb:"update,opt"`
	Remove            bool            `ferretdb:"remove,opt"`
	Upsert            bool            `ferretdb:"upsert,opt"`
	ReturnNewDocument bool            `ferretdb:"new,opt,numericBool"`
	MaxTimeMS         int64           `ferretdb:"maxTimeMS,opt,wholePositiveNumber"`

	Update      *types.Document `ferretdb:"-"`
	Aggregation *types.Array    `ferretdb:"-"`

	HasUpdateOperators bool `ferretdb:"-"`

	Let          *types.Document `ferretdb:"let,unimplemented"`
	Collation    *types.Document `ferretdb:"collation,unimplemented"`
	Fields       *types.Document `ferretdb:"fields,unimplemented"`
	ArrayFilters *types.Array    `ferretdb:"arrayFilters,unimplemented"`

	Hint                     string          `ferretdb:"hint,ignored"`
	WriteConcern             *types.Document `ferretdb:"writeConcern,ignored"`
	BypassDocumentValidation bool            `ferretdb:"bypassDocumentValidation,ignored"`
}

// UpsertParams represents parameters for upsert, if the document exists UpdateParams is set.
// Otherwise, Insert is set. It returns ReturnValue to return to the client.
type UpsertParams struct {
	// ReturnValue is the value set on the command response.
	// It returns original document for update operation, and null for insert operation.
	// If FindAndModifyParams.ReturnNewDocument is true, it returns upserted document.
	ReturnValue any

	// Upsert is a document used for insert or update operation.
	Upsert *types.Document

	// Operation is the type of upsert to perform.
	Operation UpsertOperation
}

// GetFindAndModifyParams returns `findAndModifyParams` command parameters.
func GetFindAndModifyParams(doc *types.Document, l *zap.Logger) (*FindAndModifyParams, error) {
	var params FindAndModifyParams

	err := commonparams.ExtractParams(doc, "findAndModify", &params, l)
	if err != nil {
		return nil, err
	}

	if params.Collection == "" {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", params.DB),
		)
	}

	if params.UpdateValue == nil && !params.Remove {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrFailedToParse,
			"Either an update or remove=true must be specified",
		)
	}

	if params.ReturnNewDocument && params.Remove {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrFailedToParse,
			"Cannot specify both new=true and remove=true; 'remove' always returns the deleted document",
		)
	}

	if params.UpdateValue != nil {
		switch updateParam := params.UpdateValue.(type) {
		case *types.Document:
			params.Update = updateParam
		case *types.Array:
			// TODO aggregation pipeline stages metrics
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Aggregation pipelines are not supported yet",
				"update",
			)
		default:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"Update argument must be either an object or an array",
				"update",
			)
		}
	}

	if params.Update != nil && params.Remove {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrFailedToParse,
			"Cannot specify both an update and remove=true",
		)
	}

	if params.Upsert && params.Remove {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrFailedToParse,
			"Cannot specify both upsert=true and remove=true",
		)
	}

	hasUpdateOperators, err := HasSupportedUpdateModifiers("findAndModify", params.Update)
	if err != nil {
		return nil, err
	}

	params.HasUpdateOperators = hasUpdateOperators

	return &params, nil
}

// PrepareDocumentForUpsert prepares the document used for upsert operation.
// If docs is empty it prepares a document for insert using params.
// Otherwise, it takes the first document of docs and prepare document for update.
// It sets the value to return on the command response using ReturnNewDocument param.
func PrepareDocumentForUpsert(docs []*types.Document, params *FindAndModifyParams) (*UpsertParams, error) {
	res := new(UpsertParams)
	var err error

	if len(docs) == 0 {
		res.Operation = UpsertOperationInsert
		res.Upsert, err = prepareDocumentForInsert(params)

		// insert operation returns null since no document existed before upsert.
		res.ReturnValue = types.Null
		if params.ReturnNewDocument {
			// if return ReturnNewDocument is set, return newly inserted doc.
			res.ReturnValue = res.Upsert
		}

		return res, err
	}

	res.Operation = UpsertOperationUpdate
	res.Upsert, err = prepareDocumentForUpdate(docs, params)

	// update operation returns the document before updated was applied.
	res.ReturnValue = docs[0]
	if params.ReturnNewDocument {
		// if return ReturnNewDocument is set, return updated doc.
		res.ReturnValue = res.Upsert
	}

	return res, err
}

// prepareDocumentForInsert creates an insert document from the parameter.
// When inserting new document we must check that `_id` is present, so we must extract `_id`
// from query or generate a new one.
func prepareDocumentForInsert(params *FindAndModifyParams) (*types.Document, error) {
	insert := types.NewEmptyDocument()

	if params.HasUpdateOperators {
		if _, err := UpdateDocument("findAndModify", insert, params.Update); err != nil {
			return nil, err
		}
	} else {
		insert = params.Update
	}

	if !insert.Has("_id") {
		id, err := getUpsertID(params.Query)
		if err != nil {
			return nil, err
		}

		insert.Set("_id", id)
	}

	return insert, nil
}

// prepareDocumentForUpdate takes the first document of docs and apply update params.
func prepareDocumentForUpdate(docs []*types.Document, params *FindAndModifyParams) (*types.Document, error) {
	update := docs[0].DeepCopy()

	if params.HasUpdateOperators {
		if _, err := UpdateDocument("findAndModify", update, params.Update); err != nil {
			return nil, err
		}

		return update, nil
	}

	for _, k := range params.Update.Keys() {
		v := must.NotFail(params.Update.Get(k))
		if k == "_id" {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrImmutableField,
				fmt.Sprintf(
					"Plan executor error during findAndModify :: caused "+
						"by :: After applying the update, the (immutable) field "+
						"'_id' was found to have been altered to _id: \"%s\"",
					v,
				),
				"findAndModify",
			)
		}

		update.Set(k, v)
	}

	return update, nil
}

// getUpsertID gets the _id to use for upsert document. If query contains _id,
// that _id is assigned unless _id contains operator. Otherwise, it generates an ID.
func getUpsertID(query *types.Document) (any, error) {
	id, err := query.Get("_id")
	if err != nil {
		return types.NewObjectID(), nil
	}

	idDoc, ok := id.(*types.Document)
	if !ok {
		return id, nil
	}

	_, hasOp, err := hasFilterOperator(idDoc)
	if err != nil {
		return nil, err
	}

	if hasOp {
		// if there is an operator in the query, the _id of the query cannot be used.
		// generate a new one.
		return types.NewObjectID(), nil
	}

	return id, nil
}

// hasFilterOperator returns true if query contains filter operator among with its key.
// When sub query contains any key/operator prefixed with $, it returns error.
func hasFilterOperator(query *types.Document) (string, bool, error) {
	iter := query.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			return "", false, nil
		}

		if strings.HasPrefix(k, "$") {
			return k, true, nil
		}

		doc, ok := v.(*types.Document)
		if !ok {
			continue
		}

		opKey, hasOp, err := hasFilterOperator(doc)
		if err != nil {
			return "", false, err
		}

		if hasOp {
			return "", false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrDollarPrefixedFieldName,
				fmt.Sprintf(
					"Plan executor error during findAndModify :: "+
						"caused by :: _id fields may not contain '$'-prefixed "+
						"fields: %s is not valid for storage.",
					opKey,
				),
				"findAndModify",
			)
		}
	}
}

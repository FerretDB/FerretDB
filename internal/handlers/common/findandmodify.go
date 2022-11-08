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

	"github.com/FerretDB/FerretDB/internal/types"
)

// FindAndModifyParams represent all findAndModify requests' fields.
// It's filled by calling prepareFindAndModifyParams.
type FindAndModifyParams struct {
	DB, Collection, Comment               string
	Query, Sort, Update                   *types.Document
	Remove, Upsert                        bool
	ReturnNewDocument, HasUpdateOperators bool
	MaxTimeMS                             int32
}

// PrepareFindAndModifyParams prepares findAndModify request fields.
func PrepareFindAndModifyParams(document *types.Document) (*FindAndModifyParams, error) {
	var err error

	command := document.Command()

	db, err := GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	if collection == "" {
		return nil, NewCommandErrorMsg(
			ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", db),
		)
	}

	remove, err := GetBoolOptionalParam(document, "remove")
	if err != nil {
		return nil, err
	}

	returnNewDocument, err := GetBoolOptionalParam(document, "new")
	if err != nil {
		return nil, err
	}

	upsert, err := GetBoolOptionalParam(document, "upsert")
	if err != nil {
		return nil, err
	}

	query, err := GetOptionalParam(document, "query", new(types.Document))
	if err != nil {
		return nil, err
	}

	sort, err := GetOptionalParam(document, "sort", new(types.Document))
	if err != nil {
		return nil, err
	}

	maxTimeMS, err := GetOptionalPositiveNumber(document, "maxTimeMS")
	if err != nil {
		return nil, err
	}

	var update *types.Document

	updateParam, err := document.Get("update")
	if err != nil && !remove {
		return nil, NewCommandErrorMsg(ErrFailedToParse, "Either an update or remove=true must be specified")
	}

	if err == nil {
		switch updateParam := updateParam.(type) {
		case *types.Document:
			update = updateParam
		case *types.Array:
			// TODO aggregation pipeline stages metrics
			return nil, NewCommandErrorMsgWithArgument(ErrNotImplemented, "Aggregation pipelines are not supported yet", "update")
		default:
			return nil, NewCommandErrorMsgWithArgument(
				ErrFailedToParse,
				"Update argument must be either an object or an array",
				"update",
			)
		}
	}

	if update != nil && remove {
		return nil, NewCommandErrorMsg(ErrFailedToParse, "Cannot specify both an update and remove=true")
	}

	if upsert && remove {
		return nil, NewCommandErrorMsg(ErrFailedToParse, "Cannot specify both upsert=true and remove=true")
	}

	if returnNewDocument && remove {
		return nil, NewCommandErrorMsg(
			ErrFailedToParse,
			"Cannot specify both new=true and remove=true; 'remove' always returns the deleted document",
		)
	}

	hasUpdateOperators, err := HasSupportedUpdateModifiers(update)
	if err != nil {
		return nil, err
	}

	var comment string
	// get comment from a "comment" field
	if comment, err = GetOptionalParam(document, "comment", comment); err != nil {
		return nil, err
	}

	// get comment from query, e.g. db.collection.FindAndModify({"_id":"string", "$comment: "test"},{$set:{"v":"foo""}})
	if comment, err = GetOptionalParam(query, "$comment", comment); err != nil {
		return nil, err
	}

	return &FindAndModifyParams{
		DB:                 db,
		Collection:         collection,
		Comment:            comment,
		Query:              query,
		Update:             update,
		Sort:               sort,
		Remove:             remove,
		Upsert:             upsert,
		ReturnNewDocument:  returnNewDocument,
		HasUpdateOperators: hasUpdateOperators,
		MaxTimeMS:          maxTimeMS,
	}, nil
}

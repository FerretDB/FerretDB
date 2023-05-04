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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdatesParams represents parameters for update command.
type UpdatesParams struct {
	DB         string
	Collection string
	Updates    []UpdateParams
}

// UpdateParams represents a single update operation parameters.
type UpdateParams struct {
	Filter  *types.Document
	Update  *types.Document
	Comment string
	Multi   bool
	Upsert  bool
}

// GetUpdateParams returns parameters for update command.
func GetUpdateParams(document *types.Document, l *zap.Logger) (*UpdatesParams, error) {
	var err error

	if err = Unimplemented(document, "let"); err != nil {
		return nil, err
	}

	Ignored(document, l, "ordered", "writeConcern", "bypassDocumentValidation")

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
			fmt.Sprintf("collection name has invalid type %s", AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	var updatesArray *types.Array
	if updatesArray, err = GetOptionalParam(document, "updates", updatesArray); err != nil {
		return nil, err
	}

	var updates []UpdateParams

	for i := 0; i < updatesArray.Len(); i++ {
		update, err := AssertType[*types.Document](must.NotFail(updatesArray.Get(i)))
		if err != nil {
			return nil, err
		}

		if err = Unimplemented(update, "c", "collation", "arrayFilters"); err != nil {
			return nil, err
		}

		Ignored(update, l, "hint")

		var q, u *types.Document

		if q, err = GetOptionalParam(update, "q", q); err != nil {
			return nil, err
		}

		if u, err = GetOptionalParam(update, "u", u); err != nil {
			// TODO check if u is an array of aggregation pipeline stages
			return nil, err
		}

		var comment string

		// get comment from options.UpdateParams().SetComment() method
		if comment, err = GetOptionalParam(document, "comment", comment); err != nil {
			return nil, err
		}

		// get comment from query, e.g. db.collection.UpdateOne({"_id":"string", "$comment: "test"},{$set:{"v":"foo""}})
		if comment, err = GetOptionalParam(q, "$comment", comment); err != nil {
			return nil, err
		}

		if u != nil {
			if err = ValidateUpdateOperators(document.Command(), u); err != nil {
				return nil, err
			}
		}

		var upsert, multi bool

		if upsert, err = GetOptionalParam(update, "upsert", upsert); err != nil {
			return nil, err
		}

		if multi, err = GetOptionalParam(update, "multi", multi); err != nil {
			return nil, err
		}

		updates = append(updates, UpdateParams{
			Filter:  q,
			Update:  u,
			Upsert:  upsert,
			Multi:   multi,
			Comment: comment,
		})
	}

	return &UpdatesParams{
		DB:         db,
		Collection: collection,
		Updates:    updates,
	}, nil
}

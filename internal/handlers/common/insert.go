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
)

// InsertParams represents the parameters for an insert command.
type InsertParams struct {
	Docs           *types.Array
	DB, Collection string
	Ordered        bool
}

// GetInsertParams returns the parameters for an insert command.
func GetInsertParams(document *types.Document, l *zap.Logger) (*InsertParams, error) {
	var err error

	Ignored(document, l, "writeConcern", "bypassDocumentValidation", "comment")

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

	var docs *types.Array
	if docs, err = GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	return &InsertParams{
		DB:         db,
		Collection: collection,
		Ordered:    ordered,
		Docs:       docs,
	}, nil
}

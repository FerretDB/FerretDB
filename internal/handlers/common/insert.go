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

// InsertParams represents the parameters for an insert command.
type InsertParams struct {
	Docs       *types.Array `ferretdb:"documents,opt"`
	DB         string       `ferretdb:"$db"`
	Collection string       `ferretdb:"insert,collection"`
	Ordered    bool         `ferretdb:"ordered,opt"`

	WriteConcern             any    `ferretdb:"writeConcern,ignored"`
	BypassDocumentValidation bool   `ferretdb:"bypassDocumentValidation,ignored"`
	Comment                  string `ferretdb:"comment,ignored"`
	LSID                     any    `ferretdb:"lsid,ignored"`
}

// GetInsertParams returns the parameters for an insert command.
func GetInsertParams(document *types.Document, l *zap.Logger) (*InsertParams, error) {
	params := InsertParams{
		Ordered: true,
	}

	err := commonparams.ExtractParams(document, "insert", &params, l)
	if err != nil {
		return nil, err
	}

	for i := 0; i < params.Docs.Len(); i++ {
		doc := must.NotFail(params.Docs.Get(i))

		if _, ok := doc.(*types.Document); !ok {
			return nil, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'insert.documents.%d' is the wrong type '%s', expected type 'object'",
					i,
					commonparams.AliasFromType(doc),
				),
			)
		}
	}

	return &params, nil
}

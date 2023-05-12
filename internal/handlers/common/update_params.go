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
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// UpdatesParams represents parameters for update command.
type UpdatesParams struct {
	DB         string         `ferretdb:"$db"`
	Collection string         `ferretdb:"collection"`
	Updates    []UpdateParams `ferretdb:"updates"`

	Comment string `ferretdb:"comment,opt"`

	Let *types.Document `ferretdb:"let,unimplemented"`

	Ordered                  bool            `ferretdb:"ordered,ignored"`
	BypassDocumentValidation bool            `ferretdb:"bypassDocumentValidation,ignored"`
	WriteConcern             *types.Document `ferretdb:"writeConcern,ignored"`
}

// UpdateParams represents a single update operation parameters.
type UpdateParams struct {
	// TODO: query could have a comment that we probably should extract.
	// get comment from query, e.g. db.collection.UpdateOne({"_id":"string", "$comment: "test"},{$set:{"v":"foo""}})
	Filter *types.Document `ferretdb:"q,opt"`
	Update *types.Document `ferretdb:"u,opt"`
	Multi  bool            `ferretdb:"multi,opt"`
	Upsert bool            `ferretdb:"upsert,opt"`

	C            *types.Document `ferretdb:"c,unimplemented"`
	Collation    *types.Document `ferretdb:"collation,unimplemented"`
	ArrayFilters *types.Array    `ferretdb:"arrayFilters,unimplemented"`

	Hint string `ferretdb:"hint,ignored"`
}

// GetUpdateParams returns parameters for update command.
func GetUpdateParams(document *types.Document, l *zap.Logger) (*UpdatesParams, error) {
	var params UpdatesParams

	err := commonparams.ExtractParams(document, "update", &params, l)
	if err != nil {
		return nil, err
	}

	for _, update := range params.Updates {
		if err := ValidateUpdateOperators(document.Command(), update.Update); err != nil {
			return nil, err
		}
	}

	return &params, nil
}

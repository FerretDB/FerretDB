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

// UpdateParams represents parameters for the update command.
//
//nolint:vet // for readability
type UpdateParams struct {
	DB         string `ferretdb:"$db"`
	Collection string `ferretdb:"update,collection"`

	Updates []Update `ferretdb:"updates"`

	Comment string `ferretdb:"comment,opt"`

	Let *types.Document `ferretdb:"let,unimplemented"`

	Ordered                  bool            `ferretdb:"ordered,ignored"`
	BypassDocumentValidation bool            `ferretdb:"bypassDocumentValidation,ignored"`
	WriteConcern             *types.Document `ferretdb:"writeConcern,ignored"`
	LSID                     any             `ferretdb:"lsid,ignored"`
}

// Update represents a single update operation parameters.
//
//nolint:vet // for readability
type Update struct {
	Filter *types.Document `ferretdb:"q,opt"`
	Update *types.Document `ferretdb:"u,opt"` // TODO https://github.com/FerretDB/FerretDB/issues/2742
	Multi  bool            `ferretdb:"multi,opt"`
	Upsert bool            `ferretdb:"upsert,opt,numericBool"`

	C            *types.Document `ferretdb:"c,unimplemented"`
	Collation    *types.Document `ferretdb:"collation,unimplemented"`
	ArrayFilters *types.Array    `ferretdb:"arrayFilters,unimplemented"`

	Hint string `ferretdb:"hint,ignored"`
}

// GetUpdateParams returns parameters for update command.
func GetUpdateParams(document *types.Document, l *zap.Logger) (*UpdateParams, error) {
	var params UpdateParams

	err := commonparams.ExtractParams(document, "update", &params, l)
	if err != nil {
		return nil, err
	}

	if len(params.Updates) > 0 {
		for _, update := range params.Updates {
			if update.Update == nil {
				continue
			}

			if err := ValidateUpdateOperators(document.Command(), update.Update); err != nil {
				return nil, err
			}
		}
	}

	return &params, nil
}

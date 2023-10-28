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

// DeleteParams represents parameters for the delete command.
//
//nolint:vet // for readability
type DeleteParams struct {
	DB         string `ferretdb:"$db"`
	Collection string `ferretdb:"delete,collection"`

	Deletes []Delete `ferretdb:"deletes,opt"`
	Comment string   `ferretdb:"comment,opt"`
	Ordered bool     `ferretdb:"ordered,opt"`

	Let *types.Document `ferretdb:"let,unimplemented"`

	WriteConcern *types.Document `ferretdb:"writeConcern,ignored"`
	LSID         any             `ferretdb:"lsid,ignored"`
}

// Delete represents single delete operation parameters.
//
//nolint:vet // for readability
type Delete struct {
	Filter  *types.Document `ferretdb:"q"`
	Limited bool            `ferretdb:"limit,zeroOrOneAsBool"`

	Collation *types.Document `ferretdb:"collation,unimplemented"`

	Hint string `ferretdb:"hint,ignored"`
}

// GetDeleteParams returns parameters for delete operation.
func GetDeleteParams(document *types.Document, l *zap.Logger) (*DeleteParams, error) {
	params := DeleteParams{
		Ordered: true,
	}

	err := commonparams.ExtractParams(document, "delete", &params, l)
	if err != nil {
		return nil, err
	}

	return &params, nil
}

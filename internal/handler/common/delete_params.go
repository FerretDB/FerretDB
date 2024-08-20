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
	"log/slog"

	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
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

	MaxTimeMS      int64           `ferretdb:"maxTimeMS,ignored"`
	WriteConcern   *types.Document `ferretdb:"writeConcern,ignored"`
	LSID           any             `ferretdb:"lsid,ignored"`
	TxnNumber      int64           `ferretdb:"txnNumber,ignored"`
	ClusterTime    any             `ferretdb:"$clusterTime,ignored"`
	ReadPreference *types.Document `ferretdb:"$readPreference,ignored"`

	ApiVersion           string `ferretdb:"apiVersion,ignored"`
	ApiStrict            bool   `ferretdb:"apiStrict,ignored"`
	ApiDeprecationErrors bool   `ferretdb:"apiDeprecationErrors,ignored"`
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
func GetDeleteParams(document *types.Document, l *slog.Logger) (*DeleteParams, error) {
	params := DeleteParams{
		Ordered: true,
	}

	err := handlerparams.ExtractParams(document, "delete", &params, l)
	if err != nil {
		return nil, err
	}

	return &params, nil
}

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

// FindParams contains `find` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type FindParams struct {
	DB          string          `ferretdb:"$db"`
	Collection  string          `ferretdb:"collection"`
	Filter      *types.Document `ferretdb:"filter,opt"`
	Sort        *types.Document `ferretdb:"sort,opt"`
	Projection  *types.Document `ferretdb:"projection,opt"`
	Skip        int64           `ferretdb:"skip,opt"`
	Limit       int64           `ferretdb:"limit,opt"`
	BatchSize   int64           `ferretdb:"batchSize,opt"`
	SingleBatch bool            `ferretdb:"singleBatch,opt"`
	Comment     string          `ferretdb:"comment,opt"`
	MaxTimeMS   int32           `ferretdb:"maxTimeMS,opt"`

	Collation any `ferretdb:"collation,unimplemented"`
	Let       any `ferretdb:"let,unimplemented"`

	AllowDiskUse any `ferretdb:"allowDiskUse,ignored"`
	ReadConcern  any `ferretdb:"readConcern,ignored"`
	Max          any `ferretdb:"max,ignored"`
	Min          any `ferretdb:"min,ignored"`
	Hint         any `ferretdb:"hint,ignored"`

	ReturnKey           any `ferretdb:"returnKey,non-default"`
	ShowRecordId        any `ferretdb:"showRecordId,non-default"`
	Tailable            any `ferretdb:"tailable,non-default"`
	OplogReplay         any `ferretdb:"oplogReplay,non-default"`
	NoCursorTimeout     any `ferretdb:"noCursorTimeout,non-default"`
	AwaitData           any `ferretdb:"awaitData,non-default"`
	AllowPartialResults any `ferretdb:"allowPartialResults,non-default"`
}

// GetFindParams returns `find` command parameters.
func GetFindParams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
	var params FindParams

	err := commonparams.ExtractParams(doc, "find", &params, l)
	if err != nil {
		return nil, err
	}

	// TODO: process default values somehow
	params.BatchSize = 101

	return &params, nil
}

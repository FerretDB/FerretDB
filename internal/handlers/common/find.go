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
	"errors"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// FindParams represents parameters for the find command.
//
//nolint:vet // for readability
type FindParams struct {
	DB          string          `ferretdb:"$db"`
	Collection  string          `ferretdb:"collection"`
	Filter      *types.Document `ferretdb:"filter,opt"`
	Sort        *types.Document `ferretdb:"sort,opt"`
	Projection  *types.Document `ferretdb:"projection,opt"`
	Skip        int64           `ferretdb:"skip,opt,positiveNumber"`
	Limit       int64           `ferretdb:"limit,opt,positiveNumber"`
	BatchSize   int64           `ferretdb:"batchSize,opt,positiveNumber"`
	SingleBatch bool            `ferretdb:"singleBatch,opt"`
	Comment     string          `ferretdb:"comment,opt"`
	MaxTimeMS   int64           `ferretdb:"maxTimeMS,opt,wholePositiveNumber"`

	Collation *types.Document `ferretdb:"collation,unimplemented"`
	Let       *types.Document `ferretdb:"let,unimplemented"`

	AllowDiskUse bool            `ferretdb:"allowDiskUse,ignored"`
	ReadConcern  *types.Document `ferretdb:"readConcern,ignored"`
	Max          *types.Document `ferretdb:"max,ignored"`
	Min          *types.Document `ferretdb:"min,ignored"`
	Hint         any             `ferretdb:"hint,ignored"`
	LSID         any             `ferretdb:"lsid,ignored"`

	ReturnKey           bool `ferretdb:"returnKey,unimplemented-non-default"`
	ShowRecordId        bool `ferretdb:"showRecordId,unimplemented-non-default"`
	Tailable            bool `ferretdb:"tailable,unimplemented-non-default"`
	OplogReplay         bool `ferretdb:"oplogReplay,unimplemented-non-default"`
	NoCursorTimeout     bool `ferretdb:"noCursorTimeout,unimplemented-non-default"`
	AwaitData           bool `ferretdb:"awaitData,unimplemented-non-default"`
	AllowPartialResults bool `ferretdb:"allowPartialResults,unimplemented-non-default"`
}

// GetFindParams returns `find` command parameters.
func GetFindParams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
	params := FindParams{
		BatchSize: 101,
	}

	err := commonparams.ExtractParams(doc, "find", &params, l)

	var ce *commonerrors.CommandError
	if errors.As(err, &ce) {
		if ce.Code() == commonerrors.ErrInvalidNamespace {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrBadValue, ce.Unwrap().Error(), "find")
		}
	}

	if err != nil {
		return nil, err
	}

	return &params, nil
}

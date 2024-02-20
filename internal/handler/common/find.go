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

	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// FindParams represents parameters for the find command.
//
//nolint:vet // for readability
type FindParams struct {
	DB           string          `ferretdb:"$db"`
	Collection   string          `ferretdb:"find,collection"`
	Filter       *types.Document `ferretdb:"filter,opt"`
	Sort         *types.Document `ferretdb:"sort,opt"`
	Projection   *types.Document `ferretdb:"projection,opt"`
	Skip         int64           `ferretdb:"skip,opt,positiveNumber"`
	Limit        int64           `ferretdb:"limit,opt,positiveNumber"`
	BatchSize    int64           `ferretdb:"batchSize,opt,positiveNumber"`
	SingleBatch  bool            `ferretdb:"singleBatch,opt"`
	Comment      string          `ferretdb:"comment,opt"`
	MaxTimeMS    int64           `ferretdb:"maxTimeMS,opt,wholePositiveNumber"`
	ShowRecordId bool            `ferretdb:"showRecordId,opt"`
	Tailable     bool            `ferretdb:"tailable,opt"`
	AwaitData    bool            `ferretdb:"awaitData,opt"`

	Collation *types.Document `ferretdb:"collation,unimplemented"`
	Let       *types.Document `ferretdb:"let,unimplemented"`

	AllowDiskUse     bool            `ferretdb:"allowDiskUse,ignored"`
	ReadConcern      *types.Document `ferretdb:"readConcern,ignored"`
	Max              *types.Document `ferretdb:"max,ignored"`
	Min              *types.Document `ferretdb:"min,ignored"`
	Hint             any             `ferretdb:"hint,ignored"`
	LSID             any             `ferretdb:"lsid,ignored"`
	TxnNumber        int64           `ferretdb:"txnNumber,ignored"`
	StartTransaction bool            `ferretdb:"startTransaction,ignored"`
	Autocommit       bool            `ferretdb:"autocommit,ignored"`
	ClusterTime      any             `ferretdb:"$clusterTime,ignored"`
	ReadPreference   *types.Document `ferretdb:"$readPreference,ignored"`

	ReturnKey           bool `ferretdb:"returnKey,unimplemented-non-default"`
	OplogReplay         bool `ferretdb:"oplogReplay,ignored"`
	AllowPartialResults bool `ferretdb:"allowPartialResults,unimplemented-non-default"`

	// TODO https://github.com/FerretDB/FerretDB/issues/4035
	NoCursorTimeout bool `ferretdb:"noCursorTimeout,unimplemented-non-default"`
}

// GetFindParams returns `find` command parameters.
func GetFindParams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
	params := FindParams{
		BatchSize: 101,
	}

	err := handlerparams.ExtractParams(doc, "find", &params, l)

	var ce *handlererrors.CommandError
	if errors.As(err, &ce) {
		if ce.Code() == handlererrors.ErrInvalidNamespace {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, ce.Err().Error(), "find")
		}
	}

	if err != nil {
		return nil, err
	}

	if params.AwaitData && !params.Tailable {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrFailedToParse,
			"Cannot set 'awaitData' without also setting 'tailable'",
			"find",
		)
	}

	return &params, nil
}

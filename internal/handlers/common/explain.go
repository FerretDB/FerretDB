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

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ExplainParams represents the parameters for the explain command.
type ExplainParams struct {
	DB         string `ferretdb:"$db"`
	Collection string `ferretdb:"collection"`

	Explain *types.Document `ferretdb:"explain"`

	Filter *types.Document `ferretdb:"filter,opt"`
	Sort   *types.Document `ferretdb:"sort,opt"`
	Skip   int64           `ferretdb:"skip,opt"`
	Limit  int64           `ferretdb:"limit,opt"`

	StagesDocs []any           `ferretdb:"-"`
	Aggregate  bool            `ferretdb:"-"`
	Command    *types.Document `ferretdb:"-"`

	Verbosity string `ferretdb:"verbosity,ignored"`
}

// GetExplainParams returns the parameters for the explain command.
func GetExplainParams(document *types.Document, l *zap.Logger) (*ExplainParams, error) {
	var err error

	var db, collection string

	if db, err = GetRequiredParam[string](document, "$db"); err != nil {
		return nil, lazyerrors.Error(err)
	}

	Ignored(document, l, "verbosity")

	var cmd *types.Document

	cmd, err = GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if collection, err = GetRequiredParam[string](cmd, cmd.Command()); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var explain, filter, sort *types.Document

	cmd, err = GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	explain, err = GetRequiredParam[*types.Document](document, "explain")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err = GetOptionalParam(explain, "filter", filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sort, err = GetOptionalParam(explain, "sort", sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var limit, skip int64

	if limit, err = GetLimitParam(explain); err != nil {
		return nil, err
	}

	if limit, err = commonparams.GetValidatedNumberParamWithMinValue("explain", "limit", limit, 0); err != nil {
		return nil, err
	}

	if skip, err = GetOptionalParam(explain, "skip", skip); err != nil {
		return nil, err
	}

	if skip, err = commonparams.GetValidatedNumberParamWithMinValue("explain", "skip", skip, 0); err != nil {
		return nil, err
	}

	var stagesDocs []any

	if cmd.Command() == "aggregate" {
		var pipeline *types.Array

		pipeline, err = GetRequiredParam[*types.Array](explain, "pipeline")
		if err != nil {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrMissingField,
				"BSON field 'aggregate.pipeline' is missing but a required field",
				document.Command(),
			)
		}

		stagesDocs = must.NotFail(iterator.ConsumeValues(pipeline.Iterator()))
		for _, d := range stagesDocs {
			if _, ok := d.(*types.Document); !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					"Each element of the 'pipeline' array must be an object",
					document.Command(),
				)
			}
		}
	}

	return &ExplainParams{
		DB:         db,
		Collection: collection,
		Filter:     filter,
		Sort:       sort,
		Skip:       skip,
		Limit:      limit,
		StagesDocs: stagesDocs,
		Aggregate:  cmd.Command() == "aggregate",
		Command:    cmd,
	}, nil
}

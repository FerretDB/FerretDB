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

package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handler/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handler/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"go.uber.org/zap"
)

// MsgExplain implements `explain` command.
func (h *Handler) MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := GetExplainParams(document, h.L)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	serverInfo := must.NotFail(types.NewDocument(
		"host", hostname,
		"version", version.Get().MongoDBVersion,
		"gitVersion", version.Get().Commit,

		// our extensions
		"ferretdbVersion", version.Get().Version,
	))

	cmd := params.Command
	cmd.Set("$db", params.DB)

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	coll, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	qp := backends.ExplainParams{
		Filter: params.Filter,
	}

	if params.Aggregate {
		qp.Filter, params.Sort = aggregations.GetPushdownQuery(params.StagesDocs)
	}

	// Limit pushdown is not applied if:
	//  - `filter` is set, it must fetch all documents to filter them in memory;
	//  - `sort` is set, it must fetch all documents and sort them in memory;
	//  - `skip` is non-zero value, skip pushdown is not supported yet.
	if params.Filter.Len() == 0 && params.Sort.Len() == 0 && params.Skip == 0 {
		qp.Limit = params.Limit
	}

	if h.DisableFilterPushdown {
		qp.Filter = nil
	}

	res, err := coll.Explain(ctx, &qp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"queryPlanner", res.QueryPlanner,
			"explainVersion", "1",
			"command", cmd,
			"serverInfo", serverInfo,

			// our extensions
			// TODO https://github.com/FerretDB/FerretDB/issues/3235
			"filterPushdown", res.FilterPushdown,
			"sortPushdown", res.SortPushdown,
			"limitPushdown", res.LimitPushdown,

			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

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

	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, l, "verbosity")

	var cmd *types.Document

	cmd, err = common.GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if collection, err = common.GetRequiredParam[string](cmd, cmd.Command()); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var explain, filter, sort *types.Document

	cmd, err = common.GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	explain, err = common.GetRequiredParam[*types.Document](document, "explain")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err = common.GetOptionalParam(explain, "filter", filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sort, err = common.GetOptionalParam(explain, "sort", sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var limit, skip int64

	if limit, err = common.GetLimitParam(explain); err != nil {
		return nil, err
	}

	if limit, err = commonparams.GetValidatedNumberParamWithMinValue("explain", "limit", limit, 0); err != nil {
		return nil, err
	}

	if skip, err = common.GetOptionalParam(explain, "skip", skip); err != nil {
		return nil, err
	}

	if skip, err = commonparams.GetValidatedNumberParamWithMinValue("explain", "skip", skip, 0); err != nil {
		return nil, err
	}

	var stagesDocs []any

	if cmd.Command() == "aggregate" {
		var pipeline *types.Array

		pipeline, err = common.GetRequiredParam[*types.Array](explain, "pipeline")
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

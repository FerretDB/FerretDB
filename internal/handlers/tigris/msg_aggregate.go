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

package tigris

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1892
	common.Ignored(document, h.L, "cursor", "lsid")

	if err = common.Unimplemented(document, "explain", "collation", "let"); err != nil {
		return nil, err
	}

	common.Ignored(
		document, h.L,
		"allowDiskUse", "maxTimeMS", "bypassDocumentValidation", "readConcern", "hint", "comment", "writeConcern",
	)

	var db string

	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	// TODO handle collection-agnostic pipelines ({aggregate: 1})
	// https://github.com/FerretDB/FerretDB/issues/1890
	var ok bool
	var collection string

	if collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFailedToParse,
			"Invalid command format: the 'aggregate' field must specify a collection name or 1",
			document.Command(),
		)
	}

	pipeline, err := common.GetRequiredParam[*types.Array](document, "pipeline")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"'pipeline' option must be specified as an array",
			document.Command(),
		)
	}

	stages := must.NotFail(iterator.ConsumeValues(pipeline.Iterator()))
	stagesDocuments := make([]aggregations.Stage, 0, len(stages))
	stagesStats := make([]aggregations.Stage, 0, len(stages))

	for _, d := range stages {
		d, ok := d.(*types.Document)
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"Each element of the 'pipeline' array must be an object",
				document.Command(),
			)
		}

		var s aggregations.Stage

		if s, err = aggregations.NewStage(d); err != nil {
			return nil, err
		}

		switch s.Type() {
		case aggregations.StageTypeDocuments:
			stagesDocuments = append(stagesDocuments, s)
		case aggregations.StageTypeStats:
			stagesStats = append(stagesStats, s)
		default:
			panic(fmt.Sprintf("unknown stage type: %v", s.Type()))
		}
	}

	var resDocs []*types.Document

	if len(stagesDocuments) > 0 {
		qp := tigrisdb.QueryParams{
			DB:         db,
			Collection: collection,
			Filter:     aggregations.GetPushdownQuery(stages),
		}

		qp.Filter = aggregations.GetPushdownQuery(stages)

		if resDocs, err = processStagesDocuments(ctx, &stagesDocumentsParams{
			dbPool, &qp, stagesDocuments,
		}); err != nil {
			return nil, err
		}
	}

	if len(stagesStats) > 0 {
		statistics := aggregations.GetStatistics(stagesStats)

		if resDocs, err = processStagesStats(ctx, &stagesStatsParams{
			dbPool, db, collection, statistics, stagesStats,
		}); err != nil {
			return nil, err
		}
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1892
	firstBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		firstBatch.Append(doc)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0),
				"ns", db+"."+collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// stagesDocumentsParams contains the parameters for processStagesDocuments.
type stagesDocumentsParams struct {
	dbPool *tigrisdb.TigrisDB
	qp     *tigrisdb.QueryParams
	stages []aggregations.Stage
}

// processStagesDocuments retrieves the documents from the database and then processes them through the stages.
func processStagesDocuments(ctx context.Context, p *stagesDocumentsParams) ([]*types.Document, error) { //nolint:lll // for readability
	var docs []*types.Document

	iter, err := p.dbPool.QueryDocuments(ctx, p.qp)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	docs, err = iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))
	if err != nil {
		return nil, err
	}

	for _, s := range p.stages {
		if docs, err = s.Process(ctx, docs); err != nil {
			return nil, err
		}
	}

	return docs, nil
}

// stagesStatsParams contains the parameters for processStagesStats.
type stagesStatsParams struct {
	dbPool     *tigrisdb.TigrisDB
	db         string
	collection string
	statistics map[aggregations.Statistic]struct{}
	stages     []aggregations.Stage
}

// processStagesStats retrieves the statistics from the database and then processes them through the stages.
func processStagesStats(ctx context.Context, p *stagesStatsParams) ([]*types.Document, error) {
	// Clarify what needs to be retrieved from the database and retrieve it.
	_, hasCount := p.statistics[aggregations.StatisticCount]
	_, hasStorage := p.statistics[aggregations.StatisticStorage]

	var host string
	var err error

	host, err = os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc := must.NotFail(types.NewDocument(
		"ns", p.db+"."+p.collection,
		"host", host,
		"localTime", time.Now().UTC().Format(time.RFC3339),
	))

	var dbStats *tigrisdb.CollectionStats

	if hasCount || hasStorage {
		var exists bool

		if exists, err = p.dbPool.CollectionExists(ctx, p.db, p.collection); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !exists {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNamespaceNotFound,
				fmt.Sprintf("ns not found: %s.%s", p.db, p.collection),
				"aggregate",
			)
		}

		querier := p.dbPool.Driver.UseDatabase(p.db)
		dbStats, err = tigrisdb.FetchStats(ctx, querier, p.collection)

		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if hasStorage {
		var avgObjSize float64
		if dbStats.NumObjects > 0 {
			avgObjSize = float64(dbStats.Size) / float64(dbStats.NumObjects)
		}

		doc.Set(
			"storageStats", must.NotFail(types.NewDocument(
				"size", dbStats.Size,
				"count", dbStats.NumObjects,
				"avgObjSize", avgObjSize,
				"storageSize", dbStats.Size,
				"freeStorageSize", float64(0), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"capped", false, // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"wiredTiger", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"nindexes", int64(0), // Not supported for Tigris
				"indexDetails", must.NotFail(types.NewDocument()), // Not supported for Tigris
				"indexBuilds", must.NotFail(types.NewDocument()), // Not supported for Tigris
				"totalIndexSize", int64(0), // Not supported for Tigris
				"indexSizes", must.NotFail(types.NewDocument()), // Not supported for Tigris
			)),
		)
	}

	if hasCount {
		doc.Set(
			"count", dbStats.NumObjects,
		)
	}

	// Process the retrieved statistics through the stages.
	var res []*types.Document

	for _, s := range p.stages {
		if res, err = s.Process(ctx, []*types.Document{doc}); err != nil {
			return nil, err
		}
	}

	return res, nil
}

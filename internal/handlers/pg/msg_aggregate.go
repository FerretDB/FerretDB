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

package pg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
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

	for i, d := range stages {
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
			stagesStats = append(stagesStats, s) // It's possible to apply "documents" stages to statistics
		case aggregations.StageTypeStats:
			if i > 0 {
				// TODO Add a test to cover this error: https://github.com/FerretDB/FerretDB/issues/2349
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrCollStatsIsNotFirstStage,
					"$collStats is only valid as the first stage in a pipeline",
					document.Command(),
				)
			}
			stagesStats = append(stagesStats, s)
		default:
			panic(fmt.Sprintf("unknown stage type: %v", s.Type()))
		}
	}

	var resDocs []*types.Document

	// At this point we have a list of stages to apply to the documents or stats.
	// If stagesStats contains the same stages as stagesDocuments, we apply aggregation to documents fetched from the DB.
	// If stagesStats contains more stages than stagesDocuments, we apply aggregation to statistics fetched from the DB.
	if len(stagesStats) == len(stagesDocuments) {
		// only documents stages or no stages - fetch documents from the DB and apply stages to them
		qp := pgdb.QueryParams{
			DB:         db,
			Collection: collection,
			Filter:     aggregations.GetPushdownQuery(stages),
		}

		resDocs, err = processStagesDocuments(ctx, &stagesDocumentsParams{dbPool, &qp, stagesDocuments})
	} else {
		// stats stages are provided - fetch stats from the DB and apply stages to them
		statistics := aggregations.GetStatistics(stagesStats)

		resDocs, err = processStagesStats(ctx, &stagesStatsParams{
			dbPool, db, collection, statistics, stagesStats,
		})
	}

	if err != nil {
		return nil, err
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
	dbPool *pgdb.Pool
	qp     *pgdb.QueryParams
	stages []aggregations.Stage
}

// processStagesDocuments retrieves the documents from the database and then processes them through the stages.
func processStagesDocuments(ctx context.Context, p *stagesDocumentsParams) ([]*types.Document, error) { //nolint:lll // for readability
	var docs []*types.Document

	if err := p.dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		iter, getErr := pgdb.QueryDocuments(ctx, tx, p.qp)
		if getErr != nil {
			return getErr
		}

		var err error
		docs, err = iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))
		return err
	}); err != nil {
		return nil, err
	}

	for _, s := range p.stages {
		var err error
		if docs, err = s.Process(ctx, docs); err != nil {
			return nil, err
		}
	}

	return docs, nil
}

// stagesStatsParams contains the parameters for processStagesStats.
type stagesStatsParams struct {
	dbPool     *pgdb.Pool
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

	// var dbStats *pgdb.DBStats

	if hasCount || hasStorage {
		// dbStats, err = p.dbPool.(ctx, p.db, p.collection)

		switch {
		case err == nil:
		// do nothing
		case errors.Is(err, pgdb.ErrTableNotExist):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNamespaceNotFound,
				fmt.Sprintf("ns not found: %s.%s", p.db, p.collection),
				"aggregate",
			)
		default:
			return nil, err
		}
	}

	if hasStorage {
		//var avgObjSize int32
		//if dbStats.CountRows > 0 {
		//	avgObjSize = int32(dbStats.SizeRelation) / dbStats.CountRows
		//}
		//
		//doc.Set(
		//	"storageStats", must.NotFail(types.NewDocument(
		//		"size", int32(dbStats.SizeTotal),
		//		"count", dbStats.CountRows,
		//		"avgObjSize", avgObjSize,
		//		"storageSize", int32(dbStats.SizeRelation),
		//		"freeStorageSize", int32(0), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//		"capped", false, // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//		"wiredTiger", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//		"nindexes", dbStats.CountIndexes,
		//		"indexDetails", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//		"indexBuilds", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//		"totalIndexSize", int32(dbStats.SizeIndexes),
		//		"totalSize", int32(dbStats.SizeTotal),
		//		"indexSizes", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		//	)),
		//	)
	}

	if hasCount {
		//doc.Set(
		//	//"count", dbStats.CountRows,
		//)
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

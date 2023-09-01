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

package sqlite

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/stages"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "lsid")

	if err = common.Unimplemented(document, "explain", "collation", "let"); err != nil {
		return nil, err
	}

	common.Ignored(
		document, h.L,
		"allowDiskUse", "bypassDocumentValidation", "readConcern", "hint", "comment", "writeConcern",
	)

	var dbName string

	if dbName, err = common.GetRequiredParam[string](document, "$dbName"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	// handle collection-agnostic pipelines ({aggregate: 1})
	// TODO https://github.com/FerretDB/FerretDB/issues/1890
	var ok bool
	var collection string

	if collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFailedToParse,
			"Invalid command format: the 'aggregate' field must specify a collection name or 1",
			document.Command(),
		)
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	username, _ := conninfo.Get(ctx).Auth()

	v, _ := document.Get("maxTimeMS")
	if v == nil {
		v = int64(0)
	}

	// cannot use other existing commonparams function, they return different error codes
	maxTimeMS, err := commonparams.GetWholeNumberParam(v)
	if err != nil {
		switch {
		case errors.Is(err, commonparams.ErrUnexpectedType):
			if _, ok = v.(types.NullType); ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					"maxTimeMS must be a number",
					document.Command(),
				)
			}

			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field 'aggregate.maxTimeMS' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
					commonparams.AliasFromType(v),
				),
				document.Command(),
			)
		case errors.Is(err, commonparams.ErrNotWholeNumber):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"maxTimeMS has non-integral value",
				document.Command(),
			)
		case errors.Is(err, commonparams.ErrLongExceededPositive):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("%s value for maxTimeMS is out of range", types.FormatAnyValue(v)),
				document.Command(),
			)
		case errors.Is(err, commonparams.ErrLongExceededNegative):
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrValueNegative,
				fmt.Sprintf("BSON field 'maxTimeMS' value must be >= 0, actual value '%s'", types.FormatAnyValue(v)),
				document.Command(),
			)
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	if maxTimeMS < int64(0) {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrValueNegative,
			fmt.Sprintf("BSON field 'maxTimeMS' value must be >= 0, actual value '%s'", types.FormatAnyValue(v)),
			document.Command(),
		)
	}

	if maxTimeMS > math.MaxInt32 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("%v value for maxTimeMS is out of range", v),
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

	aggregationStages := must.NotFail(iterator.ConsumeValues(pipeline.Iterator()))
	stagesDocuments := make([]aggregations.Stage, 0, len(aggregationStages))
	collStatsDocuments := make([]aggregations.Stage, 0, len(aggregationStages))

	for i, v := range aggregationStages {
		var d *types.Document

		if d, ok = v.(*types.Document); !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"Each element of the 'pipeline' array must be an object",
				document.Command(),
			)
		}

		var s aggregations.Stage

		if s, err = stages.NewStage(d); err != nil {
			return nil, err
		}

		switch d.Command() {
		case "$collStats":
			if i > 0 {
				// Add a test to cover this error.
				// TODO https://github.com/FerretDB/FerretDB/issues/2349
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrCollStatsIsNotFirstStage,
					"$collStats is only valid as the first stage in a pipeline",
					document.Command(),
				)
			}

			collStatsDocuments = append(collStatsDocuments, s)
		default:
			stagesDocuments = append(stagesDocuments, s)
			collStatsDocuments = append(collStatsDocuments, s) // It's possible to apply any stage after $collStats stage
		}
	}

	// validate cursor after validating pipeline stages to keep compatibility
	v, _ = document.Get("cursor")
	if v == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFailedToParse,
			"The 'cursor' option is required, except for aggregate with the explain argument",
			document.Command(),
		)
	}

	cursorDoc, ok := v.(*types.Document)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				`BSON field 'cursor' is the wrong type '%s', expected type 'object'`,
				commonparams.AliasFromType(v),
			),
			document.Command(),
		)
	}

	v, _ = cursorDoc.Get("batchSize")
	if v == nil {
		v = int32(101)
	}

	batchSize, err := commonparams.GetValidatedNumberParamWithMinValue(document.Command(), "batchSize", v, 0)
	if err != nil {
		return nil, err
	}

	cancel := func() {}
	if maxTimeMS != 0 {
		// It is not clear if maxTimeMS affects only aggregate, or both aggregate and getMore (as the current code does).
		// TODO https://github.com/FerretDB/FerretDB/issues/2983
		ctx, cancel = context.WithTimeout(ctx, time.Duration(maxTimeMS)*time.Millisecond)
	}

	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))

	var iter iterator.Interface[struct{}, *types.Document]

	if len(collStatsDocuments) == len(stagesDocuments) {
		// TODO https://github.com/FerretDB/FerretDB/issues/3235
		// TODO https://github.com/FerretDB/FerretDB/issues/3181
		iter, err = processStagesDocuments(ctx, closer, &stagesDocumentsParams{c, stagesDocuments})
	} else {
		// TODO https://github.com/FerretDB/FerretDB/issues/2423
		statistics := stages.GetStatistics(collStatsDocuments)

		iter, err = processStagesStats(ctx, closer, &stagesStatsParams{
			c, dbName, collection, statistics, collStatsDocuments,
		})
	}

	if err != nil {
		closer.Close()
		return nil, err
	}

	closer.Add(iter)

	cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
		Iter:       iterator.WithClose(iter, closer.Close),
		DB:         dbName,
		Collection: collection,
		Username:   username,
	})

	cursorID := cursor.ID

	firstBatchDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor), int(batchSize))
	if err != nil {
		cursor.Close()
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(firstBatchDocs))
	for _, doc := range firstBatchDocs {
		firstBatch.Append(doc)
	}

	if firstBatch.Len() < int(batchSize) {
		// let the client know that there are no more results
		cursorID = 0

		cursor.Close()
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", cursorID,
				"ns", dbName+"."+collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// stagesDocumentsParams contains the parameters for processStagesDocuments.
type stagesDocumentsParams struct {
	c      backends.Collection
	stages []aggregations.Stage
}

// processStagesDocuments retrieves the documents from the database and then processes them through the stages.
func processStagesDocuments(ctx context.Context, closer *iterator.MultiCloser, p *stagesDocumentsParams) (types.DocumentsIterator, error) { //nolint:lll // for readability
	queryRes, err := p.c.Query(ctx, nil)
	if err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	closer.Add(queryRes.Iter)

	iter := queryRes.Iter

	for _, s := range p.stages {
		if iter, err = s.Process(ctx, iter, closer); err != nil {
			return nil, err
		}
	}

	return iter, nil
}

// stagesStatsParams contains the parameters for processStagesStats.
type stagesStatsParams struct {
	c              backends.Collection
	dbName         string
	collectionName string
	statistics     map[stages.Statistic]struct{}
	stages         []aggregations.Stage
}

// processStagesStats retrieves the statistics from the database and then processes them through the stages.
//
// Move $collStats specific logic to its stage.
// TODO https://github.com/FerretDB/FerretDB/issues/2423
func processStagesStats(ctx context.Context, closer *iterator.MultiCloser, p *stagesStatsParams) (types.DocumentsIterator, error) { //nolint:lll // for readability
	// Clarify what needs to be retrieved from the database and retrieve it.
	_, hasCount := p.statistics[stages.StatisticCount]
	_, hasStorage := p.statistics[stages.StatisticStorage]

	var host string
	var err error

	host, err = os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc := must.NotFail(types.NewDocument(
		"ns", p.dbName+"."+p.collectionName,
		"host", host,
		"localTime", time.Now().UTC().Format(time.RFC3339),
	))

	var collStats *backends.CollectionStatsResult

	if hasCount || hasStorage {
		collStats, err = p.c.Stats(ctx, new(backends.CollectionStatsParams))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if hasStorage {
		var avgObjSize int64
		if collStats.CountObjects > 0 {
			avgObjSize = collStats.SizeCollection / collStats.CountObjects
		}

		doc.Set(
			"storageStats", must.NotFail(types.NewDocument(
				"size", collStats.SizeTotal,
				"count", collStats.CountObjects,
				"avgObjSize", avgObjSize,
				"storageSize", collStats.SizeCollection,
				"freeStorageSize", int64(0), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"capped", false, // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"wiredTiger", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"nindexes", collStats.CountIndexes,
				"indexDetails", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"indexBuilds", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"totalIndexSize", collStats.SizeIndexes,
				"totalSize", collStats.SizeTotal,
				"indexSizes", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
			)),
		)
	}

	if hasCount {
		doc.Set(
			"count", collStats.CountObjects,
		)
	}

	// Process the retrieved statistics through the stages.
	iter := iterator.Values(iterator.ForSlice([]*types.Document{doc}))
	closer.Add(iter)

	for _, s := range p.stages {
		if iter, err = s.Process(ctx, iter, closer); err != nil {
			return nil, err
		}
	}

	return iter, nil
}

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
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/stages"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
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

	common.Ignored(document, h.L, "lsid")

	if err = common.Unimplemented(document, "explain", "collation", "let"); err != nil {
		return nil, err
	}

	common.Ignored(
		document, h.L,
		"allowDiskUse", "bypassDocumentValidation", "readConcern", "hint", "comment", "writeConcern",
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
	previousStages := make([]string, 0, len(aggregationStages))

	filter, sort := aggregations.GetPushdownQuery(aggregationStages)
	// only documents stages or no stages - fetch documents from the DB and apply stages to them
	qp := &pgdb.QueryParams{
		DB:         db,
		Collection: collection,
	}

	if !h.DisableFilterPushdown {
		qp.Filter = filter
	}

	if h.EnableSortPushdown {
		qp.Sort = sort
	}

	q := &aggregationQuery{
		dbPool:     dbPool,
		qp:         qp,
		db:         db,
		collection: collection,
	}

	for _, d := range aggregationStages {
		d, ok := d.(*types.Document)
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"Each element of the 'pipeline' array must be an object",
				document.Command(),
			)
		}

		var s aggregations.Stage

		if s, err = stages.NewStage(d, db, collection, previousStages, q); err != nil {
			return nil, err
		}

		stagesDocuments = append(stagesDocuments, s)
		previousStages = append(previousStages, d.Command())
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

	for i, s := range stagesDocuments {
		if i == 0 {
			if iter, err = s.FetchDocuments(ctx, closer); err != nil {
				return nil, err
			}
		}

		if iter, err = s.Process(ctx, iter, closer); err != nil {
			return nil, err
		}
	}

	closer.Add(iter)

	cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
		Iter:       iterator.WithClose(iter, closer.Close),
		DB:         db,
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
				"ns", db+"."+collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// aggregationQuery queries database using set parameters.
type aggregationQuery struct {
	dbPool     *pgdb.Pool
	qp         *pgdb.QueryParams
	db         string
	collection string
}

// QueryDocuments implements AggregateQuery interface.
func (p *aggregationQuery) QueryDocuments(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error) {
	var keepTx pgx.Tx
	var iter types.DocumentsIterator

	if err := p.dbPool.InTransactionKeep(ctx, func(tx pgx.Tx) error {
		keepTx = tx

		var err error
		iter, _, err = pgdb.QueryDocuments(ctx, tx, p.qp)
		if err != nil {
			return lazyerrors.Error(err)
		}

		closer.Add(iter)

		return nil
	}); err != nil {
		return nil, err
	}

	closer.Add(iterator.CloserFunc(func() {
		// It does not matter if we commit or rollback the read transaction,
		// but we should close it.
		// ctx could be cancelled already.
		_ = keepTx.Rollback(context.Background())
	}))

	return iter, nil
}

// QueryCollStats implements AggregateQuery interface.
func (p *aggregationQuery) QueryCollStats(ctx context.Context, closer *iterator.MultiCloser) (*aggregations.CollStats, error) {
	var collStats *pgdb.CollStats

	if err := p.dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		var err error
		var exists bool

		if exists, err = pgdb.CollectionExists(ctx, tx, p.db, p.collection); err != nil {
			return err
		}

		if !exists {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNamespaceNotFound,
				fmt.Sprintf("ns not found: %s.%s", p.db, p.collection),
				"aggregate",
			)
		}

		collStats, err = pgdb.CalculateCollStats(ctx, tx, p.db, p.collection)
		if errors.Is(err, pgdb.ErrTableNotExist) {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNamespaceNotFound,
				fmt.Sprintf("ns not found: %s.%s", p.db, p.collection),
				"aggregate",
			)
		}

		return err
	}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &aggregations.CollStats{
		CountObjects:   collStats.CountObjects,
		CountIndexes:   collStats.CountIndexes,
		SizeTotal:      collStats.SizeTotal,
		SizeIndexes:    collStats.SizeIndexes,
		SizeCollection: collStats.SizeCollection,
	}, nil
}

// check interfaces
var (
	_ aggregations.AggregateQuery = (*aggregationQuery)(nil)
)

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

// Package handler provides a universal handler implementation for all backends.
package handler

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/decorators/oplog"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "handler"
)

// Handler provides a set of methods to process clients' requests sent over wire protocol.
//
// MsgXXX methods handle OP_MSG commands.
// CmdQuery handles a limited subset of OP_QUERY messages.
//
// Handler instance is shared between all client connections.
type Handler struct {
	*NewOpts

	b backends.Backend

	cursors  *cursor.Registry
	commands map[string]command

	cappedCleanupPercentage       uint8
	cleanupCappedCollectionsDocs  *prometheus.CounterVec
	cleanupCappedCollectionsBytes *prometheus.CounterVec
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Backend backends.Backend

	L             *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// test options
	DisablePushdown         bool
	EnableOplog             bool
	CappedCleanupInterval   time.Duration
	CappedCleanupPercentage uint8
	EnableNewAuth           bool
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	b := opts.Backend

	if opts.EnableOplog {
		b = oplog.NewBackend(b, opts.L.Named("oplog"))
	}

	if opts.CappedCleanupPercentage > 100 {
		opts.CappedCleanupPercentage = 100
	}

	h := &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),

		cappedCleanupPercentage: opts.CappedCleanupPercentage,
		cleanupCappedCollectionsDocs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cleanup_capped_docs",
				Help:      "Total number of documents deleted in capped collections during cleanup.",
			},
			[]string{"db", "collection"},
		),
		cleanupCappedCollectionsBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cleanup_capped_bytes",
				Help:      "Total number of bytes freed in capped collections during cleanup.",
			},
			[]string{"db", "collection"},
		),
	}

	h.initCommands()

	if opts.EnableOplog {
		// FIXME: how to test
		// mongosh --port=27017 --eval="db.createCollection('coll', {capped: true, size: 1000, max: 100})" cleanup
		//
		// for in in (seq 100)
		//   mongosh --port=27017 --eval="db.coll.insert({data: 'document $i'})" cleanup
		// end
		//
		// fixme:
		// - Run cleanup in embedded?(flag)

		// fixme: stop goroutine if close is called
		go func() {
			ticker := time.NewTicker(opts.CappedCleanupInterval)
			for {
				<-ticker.C
				if err := h.cleanupCappedCollections(context.Background()); err != nil {
					opts.L.Error("Failed to cleanup capped collections", zap.Error(err))
				}
			}
		}()
	}

	return h, nil
}

// Close gracefully shutdowns handler.
// It should be called after listener closes all client connections and stops listening.
func (h *Handler) Close() {
	h.cursors.Close()
	// stop goroutine
}

// Describe implements prometheus.Collector interface.
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.b.Describe(ch)
	h.cursors.Describe(ch)
	h.cleanupCappedCollectionsDocs.Describe(ch)
	h.cleanupCappedCollectionsBytes.Describe(ch)
}

// Collect implements prometheus.Collector interface.
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.b.Collect(ch)
	h.cursors.Collect(ch)
	h.cleanupCappedCollectionsDocs.Collect(ch)
	h.cleanupCappedCollectionsBytes.Collect(ch)
}

// cleanupCappedCollections drops the given percent of documents from all capped collections.
func (h *Handler) cleanupCappedCollections(ctx context.Context) error {
	h.L.Debug("cleanupCappedCollections: started", zap.Uint8("percentage", h.cappedCleanupPercentage))
	defer h.L.Debug("cleanupCappedCollections: finished")

	connInfo := conninfo.New()
	connInfo.BypassAuth = true
	ctx = conninfo.Ctx(ctx, connInfo)

	dbs, err := h.b.ListDatabases(ctx, nil)
	if err != nil {
		return lazyerrors.Error(err)
	}

	for _, db := range dbs.Databases {
		database, err := h.b.Database(db.Name)
		if err != nil {
			return lazyerrors.Error(err)
		}

		if database == nil {
			continue
		}

		freed, err := h.cleanupCappedCollection(ctx, db.Name, "")
		if err != nil {
			return lazyerrors.Error(err)
		}

		h.L.Debug("cleanupCappedCollections: bytes freed in the database",
			zap.String("db", db.Name), zap.Int64("bytes", freed),
		)
	}

	return nil
}

// cleanupCappedCollection cleanup the given collection in the given database.
// If collection name is empty, it cleanups all the collections in the given database.
// It returns the number of bytes freed.
// If the given database or collection does not exist, it doesn't return an error.
func (h *Handler) cleanupCappedCollection(ctx context.Context, dbName, collName string) (int64, error) {
	db, err := h.b.Database(dbName)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	if db == nil {
		return 0, nil
	}

	var cList *backends.ListCollectionsResult

	if cList, err = db.ListCollections(ctx, nil); err != nil {
		return 0, lazyerrors.Error(err)
	}

	var collections []backends.CollectionInfo

	if collName == "" {
		collections = cList.Collections
	} else {
		var cInfo backends.CollectionInfo

		// TODO https://github.com/FerretDB/FerretDB/issues/3601
		if i, found := slices.BinarySearchFunc(cList.Collections, collName, func(e backends.CollectionInfo, t string) int {
			return cmp.Compare(e.Name, t)
		}); found {
			cInfo = cList.Collections[i]
		}

		collections = []backends.CollectionInfo{cInfo}
	}

	var totalBytesFreed int64

	for _, col := range collections {
		if !col.Capped() {
			continue
		}

		collection, err := db.Collection(collName)
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		statsBefore, err := collection.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
		if err != nil {
			if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
				continue
			}

			return 0, lazyerrors.Error(err)
		}

		if statsBefore.SizeCollection < col.CappedSize && statsBefore.CountDocuments < col.CappedDocuments {
			continue
		}

		params := backends.QueryParams{
			Limit:         int64(float64(statsBefore.CountDocuments) * float64(h.cappedCleanupPercentage) / 100),
			OnlyRecordIDs: true,
		}

		res, err := collection.Query(ctx, &params)
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		var recordIDs []int64

		for {
			_, doc, err := res.Iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return 0, lazyerrors.Error(err)
			}

			recordIDs = append(recordIDs, doc.RecordID())
		}

		deleted, err := collection.DeleteAll(ctx, &backends.DeleteAllParams{RecordIDs: recordIDs})
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		h.cleanupCappedCollectionsDocs.WithLabelValues(dbName, col.Name).Add(float64(deleted.Deleted))
		h.L.Debug("cleanupCappedCollection: documents deleted",
			zap.String("db", dbName), zap.String("collection", col.Name),
			zap.Int32("deleted", deleted.Deleted),
		)

		if _, err := collection.Compact(ctx, &backends.CompactParams{Full: false}); err != nil {
			return 0, lazyerrors.Error(err)
		}

		statsAfter, err := collection.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
		if err != nil {
			if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
				continue
			}

			return 0, lazyerrors.Error(err)
		}

		bytesFreed := statsBefore.SizeTotal - statsAfter.SizeTotal

		// There's a possibility that the size of a collection might be greater at the
		// end of a compact operation if the collection is being actively written to at
		// the time of compaction.
		if bytesFreed < 0 {
			bytesFreed = 0
		}

		h.cleanupCappedCollectionsBytes.WithLabelValues(dbName, col.Name).Add(float64(bytesFreed))
		h.L.Debug("cleanupCappedCollection: bytes freed",
			zap.String("db", dbName), zap.String("collection", col.Name),
			zap.Int64("bytes", bytesFreed),
		)

		totalBytesFreed += bytesFreed
	}

	return totalBytesFreed, nil
}

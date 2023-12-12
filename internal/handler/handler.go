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

	cappedCleanupPercentage    uint8
	cleanupedCappedCollections *prometheus.CounterVec
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
		cleanupedCappedCollections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cleanuped_capped",
				Help:      "Total number of documents cleanuped in capped collections.",
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
	h.cleanupedCappedCollections.Describe(ch)
}

// Collect implements prometheus.Collector interface.
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.b.Collect(ch)
	h.cursors.Collect(ch)
	h.cleanupedCappedCollections.Collect(ch)
}

// cleanupCappedCollections drops the given percent of documents from all capped collections.
func (h *Handler) cleanupCappedCollections(ctx context.Context) error {
	h.L.Debug("CleanupCappedCollections: started", zap.Uint8("percentage", h.cappedCleanupPercentage))
	defer h.L.Debug("CleanupCappedCollections: finished")

	// fixme: conninfo must be set, otherwise backend panics
	connInfo := conninfo.New()
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

		if err = h.cleanupCappedCollection(ctx, db.Name, ""); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// cleanupCappedCollection cleanup the given collection in the given database.
//
// If collection is empty, it cleanups all the collections in the given database.
func (h *Handler) cleanupCappedCollection(ctx context.Context, dbName, collName string) error {
	db, err := h.b.Database(dbName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if db == nil {
		return nil
	}

	var cList *backends.ListCollectionsResult

	if cList, err = db.ListCollections(ctx, nil); err != nil {
		return lazyerrors.Error(err)
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

	for _, col := range collections {
		if !col.Capped() {
			continue
		}

		collection, err := db.Collection(collName)
		if err != nil {
			return lazyerrors.Error(err)
		}

		stats, err := collection.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
		if err != nil {
			if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
				return nil
			}

			return lazyerrors.Error(err)
		}

		if stats.SizeCollection < col.CappedSize && stats.CountDocuments < col.CappedDocuments {
			return nil
		}

		params := backends.QueryParams{
			Limit:         int64(float64(stats.CountDocuments) * float64(h.cappedCleanupPercentage) / 100),
			OnlyRecordIDs: true,
		}

		res, err := collection.Query(ctx, &params)
		if err != nil {
			return lazyerrors.Error(err)
		}

		var recordIDs []int64

		for {
			_, doc, err := res.Iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return lazyerrors.Error(err)
			}

			recordIDs = append(recordIDs, doc.RecordID())
		}

		deleted, err := collection.DeleteAll(ctx, &backends.DeleteAllParams{RecordIDs: recordIDs})
		if err != nil {
			return lazyerrors.Error(err)
		}

		h.cleanupedCappedCollections.WithLabelValues(dbName, col.Name).Add(float64(deleted.Deleted))
		h.L.Debug("CleanupCappedCollections: documents deleted",
			zap.String("db", dbName), zap.String("collection", col.Name),
			zap.Int32("deleted", deleted.Deleted),
		)

		if _, err := collection.Compact(ctx, &backends.CompactParams{Full: false}); err != nil {
			return lazyerrors.Error(err)
		}

		// check stats again - how many documents left + log it
	}

	return nil
}

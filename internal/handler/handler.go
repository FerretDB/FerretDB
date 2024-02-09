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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/decorators/oplog"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
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
	wg       sync.WaitGroup

	cappedCleanupStop             chan struct{}
	cleanupCappedCollectionsDocs  *prometheus.CounterVec
	cleanupCappedCollectionsBytes *prometheus.CounterVec
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Backend     backends.Backend
	TCPHost     string
	ReplSetName string

	L             *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// test options
	DisablePushdown         bool
	EnableNestedPushdown    bool
	CappedCleanupInterval   time.Duration
	CappedCleanupPercentage uint8
	EnableNewAuth           bool
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	b := oplog.NewBackend(opts.Backend, opts.L.Named("oplog"))

	if opts.CappedCleanupPercentage >= 100 || opts.CappedCleanupPercentage <= 0 {
		return nil, fmt.Errorf(
			"percentage of documents to cleanup must be in range (0, 100), but %d given",
			opts.CappedCleanupPercentage,
		)
	}

	h := &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),

		cappedCleanupStop: make(chan struct{}),
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

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		h.runCappedCleanup()
	}()

	return h, nil
}

// runCapppedCleanup calls capped collections cleanup function according to the given interval.
func (h *Handler) runCappedCleanup() {
	if h.CappedCleanupInterval <= 0 {
		h.L.Info("Capped collections cleanup disabled.")
		return
	}

	h.L.Info("Capped collections cleanup enabled.", zap.Duration("interval", h.CappedCleanupInterval))

	ticker := time.NewTicker(h.CappedCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := h.cleanupAllCappedCollections(context.Background()); err != nil {
				h.L.Error("Failed to cleanup capped collections.", zap.Error(err))
			}

		case <-h.cappedCleanupStop:
			h.L.Info("Capped collections cleanup stopped.")
			return
		}
	}
}

// Close gracefully shutdowns handler.
// It should be called after listener closes all client connections and stops listening.
func (h *Handler) Close() {
	h.cursors.Close()
	close(h.cappedCleanupStop)
	h.wg.Wait()
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

// cleanupAllCappedCollections drops the given percent of documents from all capped collections.
func (h *Handler) cleanupAllCappedCollections(ctx context.Context) error {
	h.L.Debug("cleanupAllCappedCollections: started", zap.Uint8("percentage", h.CappedCleanupPercentage))

	start := time.Now()
	defer func() {
		h.L.Debug("cleanupAllCappedCollections: finished", zap.Duration("duration", time.Since(start)))
	}()

	connInfo := conninfo.New()
	connInfo.BypassBackendAuth = true
	ctx = conninfo.Ctx(ctx, connInfo)

	dbList, err := h.b.ListDatabases(ctx, nil)
	if err != nil {
		return lazyerrors.Error(err)
	}

	for _, dbInfo := range dbList.Databases {
		db, err := h.b.Database(dbInfo.Name)
		if err != nil {
			return lazyerrors.Error(err)
		}

		cList, err := db.ListCollections(ctx, nil)
		if err != nil {
			return lazyerrors.Error(err)
		}

		for _, cInfo := range cList.Collections {
			if !cInfo.Capped() {
				continue
			}

			deleted, bytesFreed, err := h.cleanupCappedCollection(ctx, db, &cInfo, false)
			if err != nil {
				if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) ||
					backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseDoesNotExist) {
					continue
				}

				return lazyerrors.Error(err)
			}

			if deleted > 0 || bytesFreed > 0 {
				h.L.Info("Capped collection cleaned up.",
					zap.String("db", dbInfo.Name), zap.String("collection", cInfo.Name),
					zap.Int32("deleted", deleted), zap.Int64("bytesFreed", bytesFreed),
				)
			}

			h.cleanupCappedCollectionsDocs.WithLabelValues(dbInfo.Name, cInfo.Name).Add(float64(deleted))
			h.cleanupCappedCollectionsBytes.WithLabelValues(dbInfo.Name, cInfo.Name).Add(float64(bytesFreed))
		}
	}

	return nil
}

// cleanupCappedCollection drops a percent of documents from the given capped collection and compacts it.
func (h *Handler) cleanupCappedCollection(ctx context.Context, db backends.Database, cInfo *backends.CollectionInfo, force bool) (int32, int64, error) { //nolint:lll // for readability
	must.BeTrue(cInfo.Capped())

	coll, err := db.Collection(cInfo.Name)
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	statsBefore, err := coll.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	h.L.Debug("cleanupCappedCollection: stats before", zap.Any("stats", statsBefore))

	if statsBefore.SizeCollection < cInfo.CappedSize && statsBefore.CountDocuments < cInfo.CappedDocuments {
		return 0, 0, nil
	}

	res, err := coll.Query(ctx, &backends.QueryParams{
		Sort:          must.NotFail(types.NewDocument("$natural", int64(1))),
		Limit:         int64(float64(statsBefore.CountDocuments) * float64(h.CappedCleanupPercentage) / 100),
		OnlyRecordIDs: true,
	})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	defer res.Iter.Close()

	var recordIDs []int64
	for {
		var doc *types.Document
		if _, doc, err = res.Iter.Next(); err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return 0, 0, lazyerrors.Error(err)
		}

		recordIDs = append(recordIDs, doc.RecordID())
	}

	if len(recordIDs) == 0 {
		h.L.Debug("cleanupCappedCollection: no documents to delete")
		return 0, 0, nil
	}

	deleteRes, err := coll.DeleteAll(ctx, &backends.DeleteAllParams{RecordIDs: recordIDs})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	if _, err = coll.Compact(ctx, &backends.CompactParams{Full: force}); err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	statsAfter, err := coll.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	h.L.Debug("cleanupCappedCollection: stats after", zap.Any("stats", statsAfter))

	bytesFreed := statsBefore.SizeTotal - statsAfter.SizeTotal

	// There's a possibility that the size of a collection might be greater at the
	// end of a compact operation if the collection is being actively written to at
	// the time of compaction.
	if bytesFreed < 0 {
		bytesFreed = 0
	}

	return deleteRes.Deleted, bytesFreed, nil
}

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
	"github.com/FerretDB/FerretDB/internal/handler/users"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "handler"

	// Maximum size of a batch for inserting data.
	maxWriteBatchSize = int32(100000)

	// Required by C# driver for `IsMaster` and `hello` op reply, without it `DPANIC` is thrown.
	connectionID = int32(42)

	// Default session timeout in minutes.
	logicalSessionTimeoutMinutes = int32(30)
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

	SetupDatabase string
	SetupUsername string
	SetupPassword password.Password
	SetupTimeout  time.Duration

	L             *zap.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// test options
	DisablePushdown         bool
	EnableNestedPushdown    bool
	CappedCleanupInterval   time.Duration
	CappedCleanupPercentage uint8
	EnableNewAuth           bool
	BatchSize               int
	MaxBsonObjectSizeBytes  int
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	if opts.CappedCleanupPercentage == 0 {
		opts.CappedCleanupPercentage = 10
	}

	if opts.CappedCleanupPercentage >= 100 || opts.CappedCleanupPercentage <= 0 {
		return nil, fmt.Errorf(
			"percentage of documents to cleanup must be in range (0, 100), but %d given",
			opts.CappedCleanupPercentage,
		)
	}

	if opts.MaxBsonObjectSizeBytes == 0 {
		opts.MaxBsonObjectSizeBytes = types.MaxDocumentLen
	}

	b := oplog.NewBackend(opts.Backend, opts.L.Named("oplog"))

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

	if err := h.setup(); err != nil {
		h.Close()
		return nil, err
	}

	h.initCommands()

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		h.runCappedCleanup()
	}()

	return h, nil
}

// Setup creates initial database and user if needed.
func (h *Handler) setup() error {
	if h.SetupDatabase == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.TODO(), h.SetupTimeout)
	defer cancel()

	info := conninfo.New()
	info.SetBypassBackendAuth()

	ctx = conninfo.Ctx(ctx, info)

	l := h.L.Named("setup")

	var retry int64

	for ctx.Err() == nil {
		_, err := h.b.Status(ctx, nil)
		if err == nil {
			break
		}

		l.Debug("Status failed", zap.Error(err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	res, err := h.b.ListDatabases(ctx, &backends.ListDatabasesParams{Name: h.SetupDatabase})
	if err != nil {
		return err
	}

	if len(res.Databases) > 0 {
		l.Debug("Database already exists")
		return nil
	}

	l.Info("Setting up database and user", zap.String("database", h.SetupDatabase), zap.String("username", h.SetupUsername))

	db, err := h.b.Database(h.SetupDatabase)
	if err != nil {
		return err
	}

	// that's the only way to create a database
	if err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: "setup"}); err != nil {
		return err
	}

	if err = db.DropCollection(ctx, &backends.DropCollectionParams{Name: "setup"}); err != nil {
		return err
	}

	return users.CreateUser(ctx, h.b, &users.CreateUserParams{
		Database: h.SetupDatabase,
		Username: h.SetupUsername,
		Password: h.SetupPassword,
	})
}

// runCappedCleanup calls capped collections cleanup function according to the given interval.
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
	connInfo.SetBypassBackendAuth()
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

	var docsDeleted int32
	var bytesFreed int64
	var statsBefore, statsAfter *backends.CollectionStatsResult

	coll, err := db.Collection(cInfo.Name)
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	statsBefore, err = coll.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}
	h.L.Debug("cleanupCappedCollection: stats before", zap.Any("stats", statsBefore))

	// In order to be more precise w.r.t number of documents getting dropped and to avoid
	// deleting too many documents unnecessarily,
	//
	// - First, drop the surplus documents, if document count exceeds capped configuration.
	// - Collect stats again.
	// - If collection size still exceeds the capped size, then drop the documents based on
	//   CappedCleanupPercentage.

	if count := getDocCleanupCount(cInfo, statsBefore); count > 0 {
		err = deleteFirstNDocuments(ctx, coll, count)
		if err != nil {
			return 0, 0, lazyerrors.Error(err)
		}

		statsAfter, err = coll.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
		if err != nil {
			return 0, 0, lazyerrors.Error(err)
		}

		h.L.Debug("cleanupCappedCollection: stats after document count reduction", zap.Any("stats", statsAfter))

		docsDeleted += int32(count)
		bytesFreed += (statsBefore.SizeTotal - statsAfter.SizeTotal)

		statsBefore = statsAfter
	}

	if count := getSizeCleanupCount(cInfo, statsBefore, h.CappedCleanupPercentage); count > 0 {
		err = deleteFirstNDocuments(ctx, coll, count)
		if err != nil {
			return 0, 0, lazyerrors.Error(err)
		}

		docsDeleted += int32(count)
	}

	if _, err = coll.Compact(ctx, &backends.CompactParams{Full: force}); err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	statsAfter, err = coll.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
	if err != nil {
		return 0, 0, lazyerrors.Error(err)
	}

	h.L.Debug("cleanupCappedCollection: stats after compact", zap.Any("stats", statsAfter))

	bytesFreed += (statsBefore.SizeTotal - statsAfter.SizeTotal)

	// There's a possibility that the size of a collection might be greater at the
	// end of a compact operation if the collection is being actively written to at
	// the time of compaction.
	if bytesFreed < 0 {
		bytesFreed = 0
	}

	return docsDeleted, bytesFreed, nil
}

// getDocCleanupCount returns the number of documents to be deleted during capped collection cleanup
// based on document count of the collection and capped configuration.
func getDocCleanupCount(cInfo *backends.CollectionInfo, cStats *backends.CollectionStatsResult) int64 {
	if cInfo.CappedDocuments == 0 || cInfo.CappedDocuments >= cStats.CountDocuments {
		return 0
	}

	return (cStats.CountDocuments - cInfo.CappedDocuments)
}

// getSizeCleanupCount returns the number of documents to be deleted during capped collection cleanup
// based collection size, capped configuration and cleanup percentage.
func getSizeCleanupCount(cInfo *backends.CollectionInfo, cStats *backends.CollectionStatsResult, cleanupPercent uint8) int64 {
	if cInfo.CappedSize >= cStats.SizeCollection {
		return 0
	}

	return int64(float64(cStats.CountDocuments) * float64(cleanupPercent) / 100)
}

// deleteFirstNDocuments drops first n documents (based on order of insertion) from the collection.
func deleteFirstNDocuments(ctx context.Context, coll backends.Collection, n int64) error {
	if n == 0 {
		return nil
	}

	res, err := coll.Query(ctx, &backends.QueryParams{
		Sort:          must.NotFail(types.NewDocument("$natural", int64(1))),
		Limit:         n,
		OnlyRecordIDs: true,
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	defer res.Iter.Close()

	var recordIDs []int64

	for {
		var doc *types.Document

		_, doc, err = res.Iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return lazyerrors.Error(err)
		}

		recordIDs = append(recordIDs, doc.RecordID())
	}

	if len(recordIDs) > 0 {
		_, err := coll.DeleteAll(ctx, &backends.DeleteAllParams{RecordIDs: recordIDs})
		if err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

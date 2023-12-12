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
	"time"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/decorators/oplog"
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
	DisablePushdown bool
	EnableOplog     bool
	EnableNewAuth   bool
}

// New returns a new handler.
func New(opts *NewOpts) (*Handler, error) {
	b := opts.Backend

	if opts.EnableOplog {
		b = oplog.NewBackend(b, opts.L.Named("oplog"))
	}

	h := &Handler{
		b:       b,
		NewOpts: opts,
		cursors: cursor.NewRegistry(opts.L.Named("cursors")),

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

	// FIXME: how to test
	// docker compose exec -T mongodb mongosh --port=27017 --eval="db.createCollection('coll', {capped: true, size: 1000, max: 100}" cleanup
	//
	// for i in {1..100}; do
	// docker compose exec -T mongodb mongosh --port=27017 --eval="db.coll.insert({data: 'document $i'})" cleanup
	// done
	go func() {
		select {
		case <-time.Tick(time.Minute):
			if err := h.CleanupCappedCollections(context.Background(), 10); err != nil {
				opts.L.Error("Failed to cleanup capped collections", zap.Error(err))
			}
			// fixme: send a signal top stop everything?
		}
	}()

	return h, nil
}

// Close gracefully shutdowns handler.
// It should be called after listener closes all client connections and stops listening.
func (h *Handler) Close() {
	h.cursors.Close()
}

// Describe implements prometheus.Collector interface.
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	h.b.Describe(ch)
	h.cursors.Describe(ch)
}

// Collect implements prometheus.Collector interface.
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.b.Collect(ch)
	h.cursors.Collect(ch)
}

// CleanupCappedCollections drops the given percent of documents from all capped collections.
func (h *Handler) CleanupCappedCollections(ctx context.Context, toDrop uint8) error {
	if toDrop == 0 || toDrop > 100 {
		return fmt.Errorf("invalid percent to drop: %d (must be in range [1, 100])", toDrop)
	}

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

		cols, err := database.ListCollections(ctx, nil)
		if err != nil {
			return lazyerrors.Error(err)
		}

		for _, col := range cols.Collections {
			if !col.Capped() {
				continue
			}

			collection, err := database.Collection(col.Name)
			if err != nil {
				return lazyerrors.Error(err)
			}

			if collection == nil {
				continue
			}

			stats, err := collection.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
			if err != nil {
				if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
					continue
				}

				return lazyerrors.Error(err)
			}

			if stats.SizeCollection < col.CappedSize && stats.CountDocuments < col.CappedDocuments {
				continue
			}

			params := backends.QueryParams{
				Limit:         int64(float64(stats.CountDocuments) * float64(toDrop) / 100),
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

			h.cleanupedCappedCollections.WithLabelValues(db.Name, col.Name).Add(float64(deleted.Deleted))

			if _, err := collection.Compact(ctx, &backends.CompactParams{Full: false}); err != nil {
				return lazyerrors.Error(err)
			}
		}
	}

	return nil
}

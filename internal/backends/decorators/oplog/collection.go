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

package oplog

import (
	"cmp"
	"context"
	"slices"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// fixed OpLog database and collection names.
const (
	oplogDatabase   = "local"
	oplogCollection = "oplog.rs"
)

// collection implements backends.Collection interface by adding OpLog functionality to the wrapped collection.
type collection struct {
	origC  backends.Collection
	name   string
	dbName string
	origB  backends.Backend
	l      *zap.Logger
}

// newCollection creates a new Collection that wraps the given collection.
func newCollection(origC backends.Collection, name, dbName string, origB backends.Backend, l *zap.Logger) backends.Collection {
	return &collection{
		origC:  origC,
		name:   name,
		dbName: dbName,
		origB:  origB,
		l:      l,
	}
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	return c.origC.Query(ctx, params)
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.origC.InsertAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if oplogC := c.oplogCollection(ctx); oplogC != nil {
		oplogDocs := make([]*types.Document, len(params.Docs))

		var oplogDoc *types.Document
		now := time.Now()

		for i, doc := range params.Docs {
			d := &document{
				o:  doc,
				ns: c.dbName + "." + c.name,
				op: "i",
			}
			if oplogDoc, err = d.marshal(now); err != nil {
				c.l.Error("Failed to create document", zap.Error(err))
				return res, nil
			}

			oplogDocs[i] = oplogDoc
		}

		_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
			Docs: oplogDocs,
		})
		if err != nil {
			c.l.Error("Failed to insert documents", zap.Error(err))
		}
	}

	return res, nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.origC.UpdateAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if oplogC := c.oplogCollection(ctx); oplogC != nil {
		oplogDocs := make([]*types.Document, len(params.Docs))

		var oplogDoc *types.Document
		now := time.Now()

		for i, doc := range params.Docs {
			d := &document{
				o:  doc,
				ns: c.dbName + "." + c.name,
				op: "u",
			}
			if oplogDoc, err = d.marshal(now); err != nil {
				c.l.Error("Failed to create document", zap.Error(err))
				return res, nil
			}

			oplogDocs[i] = oplogDoc
		}

		_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
			Docs: oplogDocs,
		})
		if err != nil {
			c.l.Error("Failed to insert documents", zap.Error(err))
		}
	}

	return res, nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.origC.DeleteAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if oplogC := c.oplogCollection(ctx); oplogC != nil {
		oplogDocs := make([]*types.Document, len(params.IDs))

		var oplogDoc *types.Document
		now := time.Now()

		for i, id := range params.IDs {
			if oplogDoc, err = types.NewDocument("_id", id); err != nil {
				c.l.Error("Failed to create _id document", zap.Error(err))
				return res, nil
			}

			d := &document{
				o:  oplogDoc,
				ns: c.dbName + "." + c.name,
				op: "d",
			}
			if oplogDoc, err = d.marshal(now); err != nil {
				c.l.Error("Failed to create document", zap.Error(err))
				return res, nil
			}

			oplogDocs[i] = oplogDoc
		}

		_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
			Docs: oplogDocs,
		})
		if err != nil {
			c.l.Error("Failed to insert documents", zap.Error(err))
		}
	}

	return res, nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	return c.origC.Explain(ctx, params)
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	return c.origC.Stats(ctx, params)
}

// Compact implements backends.Collection interface.
func (c *collection) Compact(ctx context.Context, params *backends.CompactParams) (*backends.CompactResult, error) {
	return c.origC.Compact(ctx, params)
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	return c.origC.ListIndexes(ctx, params)
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	return c.origC.CreateIndexes(ctx, params)
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	return c.origC.DropIndexes(ctx, params)
}

// oplogCollection returns the OpLog collection if it exist.
func (c *collection) oplogCollection(ctx context.Context) backends.Collection {
	db := must.NotFail(c.origB.Database(oplogDatabase))

	cList, err := db.ListCollections(ctx, nil)
	if err != nil {
		c.l.Error("Failed to list collections", zap.Error(err))
		return nil
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3601
	_, found := slices.BinarySearchFunc(cList.Collections, oplogCollection, func(e backends.CollectionInfo, t string) int {
		return cmp.Compare(e.Name, t)
	})
	if !found {
		c.l.Debug("Collection not found")
		return nil
	}

	return must.NotFail(db.Collection(oplogCollection))
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)

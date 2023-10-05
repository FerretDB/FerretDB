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
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
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

// newCollection creates a new collection that wraps the given collection.
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
			c.l.Error("Failed to create oplog document", zap.Error(err))
			return res, nil
		}

		oplogDocs[i] = oplogDoc
	}

	var oplogDB backends.Database
	var oplogC backends.Collection

	if oplogDB, err = c.origB.Database(oplogDatabase); err == nil {
		if oplogC, err = oplogDB.Collection(oplogCollection); err == nil {
			_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
				Docs: oplogDocs,
			})
		}
	}

	if err != nil {
		c.l.Error("Failed to insert oplog documents", zap.Error(err))
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

	// TODO
	_ = res

	return res, nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.origC.DeleteAll(ctx, params)
	if err != nil {
		return nil, err
	}

	oplogDocs := make([]*types.Document, len(params.IDs))

	var idDoc, oplogDoc *types.Document
	now := time.Now()

	for i, id := range params.IDs {
		if idDoc, err = types.NewDocument("_id", id); err != nil {
			c.l.Error("Failed to create oplog _id document", zap.Error(err))
			return res, nil
		}

		d := &document{
			o:  idDoc,
			ns: c.dbName + "." + c.name,
			op: "d",
		}

		if oplogDoc, err = d.marshal(now); err != nil {
			c.l.Error("Failed to create oplog document", zap.Error(err))
			return res, nil
		}

		oplogDocs[i] = oplogDoc
	}

	var oplogDB backends.Database
	var oplogC backends.Collection

	if oplogDB, err = c.origB.Database(oplogDatabase); err == nil {
		if oplogC, err = oplogDB.Collection(oplogCollection); err == nil {
			_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
				Docs: oplogDocs,
			})
		}
	}

	if err != nil {
		c.l.Error("Failed to insert oplog documents", zap.Error(err))
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

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)

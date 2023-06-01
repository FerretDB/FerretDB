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
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// collection implements backends.Collection interface.
type collection struct {
	db   *database
	name string
}

// newDatabase creates a new Collection.
func newCollection(db *database, name string) backends.Collection {
	return backends.CollectionContract(&collection{
		db:   db,
		name: name,
	})
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	conn, err := c.db.b.pool.DB(c.db.name)
	if err != nil {
		return nil, err
	}

	table, err := c.db.b.metadataStorage.tableName(ctx, c.db.name, c.name)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT sjson FROM "%s"`, table)

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &backends.QueryResult{
		DocsIterator: newQueryIterator(ctx, rows, &queryIteratorParams{unmarshal: sjson.Unmarshal}),
	}, nil
}

// Insert implements backends.Collection interface.
func (c *collection) Insert(ctx context.Context, params *backends.InsertParams) (*backends.InsertResult, error) {
	err := c.db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: c.name})
	if err != nil {
		return nil, err
	}

	conn, err := c.db.b.pool.DB(c.db.name)
	if err != nil {
		return nil, err
	}

	table, err := c.db.b.metadataStorage.tableName(ctx, c.db.name, c.name)
	if err != nil {
		return nil, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			// TODO: check error
			tx.Rollback()
		}
	}()

	var inserted int64

	iter := params.Docs.Iterator()
	defer iter.Close()

	for {
		var val any

		_, val, err = iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, err
		}

		doc, ok := val.(*types.Document)
		if !ok {
			return nil, lazyerrors.Errorf("expected document, got %T", val)
		}

		query := fmt.Sprintf(`INSERT INTO "%s" (sjson) VALUES (?)`, table)

		var bytes []byte

		bytes, err = sjson.Marshal(doc)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, query, bytes)
		if err != nil {
			return nil, err
		}

		inserted++
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &backends.InsertResult{
		InsertedCount: inserted,
		Errors:        []error{},
	}, nil
}

// Update implements backends.Collection interface.
func (c *collection) Update(ctx context.Context, params *backends.UpdateParams) (*backends.UpdateResult, error) {
	var err error

	conn, err := c.db.b.pool.DB(c.db.name)
	if err != nil {
		return nil, err
	}

	table, err := c.db.b.metadataStorage.tableName(ctx, c.db.name, c.name)
	if err != nil {
		return nil, err
	}

	query := `UPDATE "%s" SET sjson = '%s' WHERE json_extract(sjson, '$._id') = '%s'`

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var committed bool

	defer func() {
		if committed {
			return
		}

		if rerr := tx.Rollback(); rerr != nil {
			if err == nil {
				err = rerr
			}
		}
	}()

	iter := params.Docs.Iterator()
	defer iter.Close()

	var updated int64

	for {
		var val any

		_, val, err = iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, err
		}

		doc, ok := val.(*types.Document)
		if !ok {
			panic(fmt.Sprintf("expected document, got %T", val))
		}

		id := must.NotFail(doc.Get("_id"))
		idBytes := strings.ReplaceAll(string(must.NotFail(sjson.MarshalSingleValue(id))), `"`, "")
		docBytes := must.NotFail(sjson.Marshal(doc))

		var res sql.Result

		res, err = tx.ExecContext(ctx, fmt.Sprintf(query, table, docBytes, idBytes))
		if err != nil {
			return nil, err
		}

		var rowsUpdated int64

		rowsUpdated, err = res.RowsAffected()
		if err != nil {
			return nil, err
		}

		updated += rowsUpdated
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	committed = true

	return &backends.UpdateResult{
		Updated: updated,
	}, nil
}

// Delete implements backends.Collection interface.
func (c *collection) Delete(ctx context.Context, params *backends.DeleteParams) (*backends.DeleteResult, error) {
	conn, err := c.db.b.pool.DB(c.db.name)
	if err != nil {
		return nil, err
	}

	res, err := c.Query(ctx, nil)
	if err != nil {
		return nil, err
	}

	iter := res.DocsIterator

	resDocs := make([]*types.Document, 0, 16)

	for {
		var doc *types.Document

		if _, doc, err = iter.Next(); err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, err
		}

		var matches bool

		if matches, err = common.FilterDocument(doc, params.Filter); err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)

		// if limit is set, no need to fetch all the documents
		if params.Limited {
			break
		}
	}

	iter.Close()

	// if no documents matched, there is nothing to delete
	if len(resDocs) == 0 {
		return new(backends.DeleteResult), nil
	}

	rowsDeleted, err := c.deleteDocuments(ctx, conn, resDocs)
	if err != nil {
		return nil, err
	}

	return &backends.DeleteResult{
		Deleted: rowsDeleted,
	}, nil
}

func (c *collection) deleteDocuments(ctx context.Context, db *sql.DB, docs []*types.Document) (int64, error) {
	var deleted int64
	var ids []any

	for _, doc := range docs {
		id := must.NotFail(doc.Get("_id"))

		ids = append(ids, strings.ReplaceAll(string(must.NotFail(sjson.MarshalSingleValue(id))), `"`, ``))
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE json_extract(sjson, '$._id') IN (?%s)",
		c.name,
		strings.Repeat(", ?", len(ids)-1),
	)

	res, err := db.ExecContext(ctx, query, ids...)
	if err != nil {
		return 0, err
	}

	d, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	deleted += d

	return deleted, nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)

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

package tigrisdb

import (
	"context"
	"errors"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// InsertManyDocuments inserts many documents into FerretDB database and collection.
// If database or collection does not exist, it will be created, the schema of the first document will be used
// to create the collection.
// Insertion is done in a transaction, if any document is not valid, it returns *types.ValidationError.
func (tdb *TigrisDB) InsertManyDocuments(ctx context.Context, db, collection string, docs *types.Array) error {
	if docs.Len() == 0 {
		return nil
	}

	if ok, _ := tdb.collectionExists(ctx, db, collection); !ok {
		doc := must.NotFail(docs.Get(0)).(*types.Document)

		schema, err := tjson.DocumentSchema(doc)
		if err != nil {
			return lazyerrors.Error(err)
		}
		schema.Title = collection
		b := must.NotFail(schema.Marshal())

		if _, err := tdb.CreateCollectionIfNotExist(ctx, db, collection, b); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return tdb.InTransaction(ctx, db, func(tx driver.Tx) error {
		iter := docs.Iterator()

		insertDocs := make([]driver.Document, docs.Len())

		for {
			i, d, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return lazyerrors.Error(err)
			}

			doc := d.(*types.Document)

			if err = doc.ValidateData(); err != nil {
				return err
			}

			b, err := tjson.Marshal(doc)
			if err != nil {
				return lazyerrors.Error(err)
			}

			insertDocs[i] = b
		}

		if _, err := tx.Insert(ctx, collection, insertDocs); err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
}

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
// If the document is not valid, it returns *types.ValidationError.
func (tdb *TigrisDB) InsertDocument(ctx context.Context, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	schema, err := tjson.DocumentSchema(doc)
	if err != nil {
		return lazyerrors.Error(err)
	}
	schema.Title = collection
	b := must.NotFail(schema.Marshal())

	if _, err := tdb.CreateCollectionIfNotExist(ctx, db, collection, b); err != nil {
		return lazyerrors.Error(err)
	}

	b, err = tjson.Marshal(doc)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = tdb.Driver.UseDatabase(db).Insert(ctx, collection, []driver.Document{b})

	return err
}

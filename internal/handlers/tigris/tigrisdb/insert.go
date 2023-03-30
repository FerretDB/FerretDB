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
	"fmt"
	"strings"

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
// Insertion is done in a single request.
// Documents are validated before insertion, if any document is not valid, it returns *types.ValidationError.
func (tdb *TigrisDB) InsertManyDocuments(ctx context.Context, db, collection string, docs *types.Array) error {
	if docs.Len() == 0 {
		return nil
	}

	collection = EncodeCollName(collection)

	iter := docs.Iterator()
	defer iter.Close()

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

	if _, err := tdb.Driver.UseDatabase(db).Insert(ctx, collection, insertDocs); err == nil ||
		(!IsNotFound(err) && !IsInvalidArgument(err)) {
		return err
	}

	doc := must.NotFail(docs.Get(0)).(*types.Document)

	schema, err := tdb.RefreshCollectionSchema(ctx, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if err = tjson.MergeDocumentSchema(schema, doc); err != nil {
		return lazyerrors.Error(err)
	}

	if _, err = tdb.CreateOrUpdateCollection(ctx, db, collection, schema); err != nil && !IsAlreadyExists(err) {
		return lazyerrors.Error(err)
	}

	_, err = tdb.Driver.UseDatabase(db).Insert(ctx, collection, insertDocs)

	return err
}

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
// If the document is not valid, it returns *types.ValidationError.
func (tdb *TigrisDB) InsertDocument(ctx context.Context, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	collection = EncodeCollName(collection)

	b, err := tjson.Marshal(doc)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if _, err = tdb.Driver.UseDatabase(db).Insert(ctx, collection, []driver.Document{b}); err == nil ||
		(!IsNotFound(err) && !IsInvalidArgument(err)) {
		return err
	}

	schema, err := tdb.RefreshCollectionSchema(ctx, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if err = tjson.MergeDocumentSchema(schema, doc); err != nil {
		return lazyerrors.Error(err)
	}

	if _, err = tdb.CreateOrUpdateCollection(ctx, db, collection, schema); err != nil {
		if IsInvalidArgument(err) && strings.HasPrefix(err.Error(), "data type mismatch for field \"") {
			keyPath := strings.TrimPrefix(strings.TrimSuffix(err.Error(), `"`), `data type mismatch for field "`)
			if !strings.Contains(keyPath, ".") {
				return lazyerrors.Error(err)
			}

			keyParts := strings.Split(keyPath, ".")
			if err = convertToMap(keyParts, schema); err != nil {
				return lazyerrors.Error(err)
			}

			if _, err = tdb.CreateOrUpdateCollection(ctx, db, collection, schema); err != nil {
				return lazyerrors.Error(err)
			}
		} else {
			return lazyerrors.Error(err)
		}
	}

	_, err = tdb.Driver.UseDatabase(db).Insert(ctx, collection, []driver.Document{b})

	return err
}

func convertToMap(keyParts []string, schema *tjson.Schema) error {
	for i := 0; i < len(keyParts)-1; i++ {
		v := keyParts[i]

		p := schema.Properties[v]
		if p.Type != tjson.Object {
			return fmt.Errorf("expected object type in schema. field %v got %v", v, p.Type)
		}

		schema = p
	}

	b := true
	schema.AdditionalProperties = &b
	delete(schema.Properties, keyParts[len(keyParts)-1])

	return nil
}

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
	"strings"
	"sync"

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// CreateOrUpdateCollection ensures that given collection exist.
// If needed, it creates both database and collection.
// It returns true if the collection was created.
func (tdb *TigrisDB) CreateOrUpdateCollection(ctx context.Context, db, collection string,
	schema *tjson.Schema,
) (bool, error) {
	_, err := tdb.createDatabaseIfNotExists(ctx, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	encCollection := EncodeCollName(collection)

	schema.Title = encCollection
	b, err := schema.Marshal()
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	exists, err := tdb.CollectionExists(ctx, db, encCollection)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	err = tdb.Driver.UseDatabase(db).CreateOrUpdateCollection(ctx, encCollection, b)
	tdb.l.Debug(
		"CreateOrUpdateCollection",
		zap.String("db", db), zap.String("collection", encCollection), zap.ByteString("schema", b), zap.Error(err),
	)

	var driverErr *driver.Error

	switch {
	case err == nil:
		CacheCollectionSchema(db, DecodeCollName(encCollection), schema)
		return !exists, nil
	case errors.As(err, &driverErr):
		if IsAlreadyExists(err) {
			return false, nil
		}

		return false, err
	default:
		return false, lazyerrors.Error(err)
	}
}

// CollectionExists returns true if collection exists.
func (tdb *TigrisDB) CollectionExists(ctx context.Context, db, collection string) (bool, error) {
	collection = EncodeCollName(collection)
	_, err := tdb.Driver.UseDatabase(db).DescribeCollection(ctx, collection)
	switch err := err.(type) {
	case nil:
		return true, nil
	case *driver.Error:
		if IsNotFound(err) {
			return false, nil
		}

		return false, lazyerrors.Error(err)
	default:
		return false, lazyerrors.Error(err)
	}
}

// EncodeCollName allows to have collection with / and . In the name.
func EncodeCollName(name string) string {
	name = strings.ReplaceAll(name, "/", "__A__")
	return strings.ReplaceAll(name, ".", "__B__")
}

// DecodeCollName opposite of EncodeCollName.
func DecodeCollName(name string) string {
	name = strings.ReplaceAll(name, "__A__", "/")
	return strings.ReplaceAll(name, "__B__", ".")
}

var schemaCache sync.Map

// CacheCollectionSchema add collection schema to schema cache.
func CacheCollectionSchema(db, collection string, schema *tjson.Schema) {
	schemaCache.Store(db+"$$$$"+collection, schema)
}

// RefreshCollectionSchema reads and caches collection schema from the Tigris.
func (tdb *TigrisDB) RefreshCollectionSchema(ctx context.Context, db, collection string) (*tjson.Schema, error) {
	coll, err := tdb.Driver.UseDatabase(db).DescribeCollection(ctx, EncodeCollName(collection))
	if err != nil {
		return nil, err
	}

	var schema tjson.Schema
	if err = schema.Unmarshal(coll.Schema); err != nil {
		return nil, lazyerrors.Error(err)
	}

	schema.Title = DecodeCollName(collection)

	schemaCache.Store(db+"$$$$"+collection, &schema)

	return &schema, nil
}

// GetCollectionSchema get schema from cache or read and caches from the Tigris.
func (tdb *TigrisDB) GetCollectionSchema(ctx context.Context, db, collection string) (*tjson.Schema, error) {
	res, ok := schemaCache.Load(db + "$$$$" + collection)
	if ok {
		return res.(*tjson.Schema), nil
	}

	return tdb.RefreshCollectionSchema(ctx, db, collection)
}

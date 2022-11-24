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

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

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

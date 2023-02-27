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

package pgdb

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/jackc/pgx/v4"
)

const (
	// PostgreSQL max index name length.
	maxIndexNameLength = 59
)

// setMetadataIndex sets the index info in the metadata table.
// It returns a PostgreSQL table name and index name that can be used to create index.
//
// Indexes are stored in the `indexes` array of metadata entry.
//
// Index settings are stored as an object:
//   - the corresponding formatted PostgreSQL index name is stored in the pgindex field;
//   - the corresponding FerretDB index name is stored in the name field;
//   - the index specification (field-order pairs) is stored in the key field;
//   - the unique flag is stored in the unique field.
//
// It returns a possibly wrapped error:
//   - ErrTableNotExist - if the metadata table doesn't exist.
//   - ErrIndexAlreadyExist - if the given index already exists.
func setIndexMetadata(ctx context.Context, tx pgx.Tx, params *indexParams) (pgTable string, pgIndex string, err error) {
	// TODO Validate index key: https://github.com/FerretDB/FerretDB/issues/1509

	pgIndex = formatIndexName(params.index)

	newIndex := must.NotFail(types.NewDocument(
		"pgindex", pgIndex,
		"name", params.index,
		"key", params.key,
		"unique", params.unique,
	))

	metadata, err := getMetadata(ctx, tx, params.db, params.collection, true)
	if err != nil {
		return "", "", err
	}

	pgTable = must.NotFail(metadata.Get("table")).(string)

	var indexes *types.Array
	if metadata.Has("indexes") {
		indexes = must.NotFail(metadata.Get("indexes")).(*types.Array)

		it := indexes.Iterator()
		for {
			_, idx, err := it.Next()
			if err != nil {
			}

			idxData := idx.(*types.Document)
			idxName := must.NotFail(idxData.Get("name")).(string)

			if idxName == params.index {
				return "", "", ErrIndexAlreadyExist
			}
		}
	}

	indexes.Append(newIndex)
	metadata.Set("indexes", indexes)

	if err = setMetadata(ctx, tx, params.db, params.collection, metadata); err != nil {
		return "", "", lazyerrors.Error(err)
	}

	return
}

// formatIndexName returns index name in form <shortened_name>_<name_hash>.
// Changing this logic will break compatibility with existing databases.
func formatIndexName(name string) string {
	hash32 := fnv.New32a()
	_ = must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxIndexNameLength - hash32.Size()*2 - 1
	truncateTo := len(name)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{}))
}

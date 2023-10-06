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

package metadata

import (
	"errors"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Indexes represents information about all indexes in a collection.
type Indexes []IndexInfo

// IndexInfo represents information about a single index.
type IndexInfo struct {
	Name    string
	PgIndex string
	Key     []IndexKeyPair
	Unique  bool
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field      string
	Descending bool
}

// deepCopy returns a deep copy.
func (indexes Indexes) deepCopy() Indexes {
	res := make(Indexes, len(indexes))

	for i, index := range indexes {
		res[i] = IndexInfo{
			Name:    index.Name,
			PgIndex: index.PgIndex,
			Key:     slices.Clone(index.Key),
			Unique:  index.Unique,
		}
	}

	return res
}

// marshal returns [*types.Array] for indexes.
func (indexes Indexes) marshal() *types.Array {
	res := types.MakeArray(len(indexes))

	for _, index := range indexes {
		key := types.MakeDocument(len(index.Key))

		for _, pair := range index.Key {
			order := int32(1)
			if pair.Descending {
				order = int32(-1)
			}

			key.Set(pair.Field, order)
		}

		res.Append(must.NotFail(types.NewDocument(
			"pgindex", index.PgIndex,
			"name", index.Name,
			"key", key,
			"unique", index.Unique,
		)))
	}

	return res
}

// unmarshal sets indexes from [*types.Array].
func (s *Indexes) unmarshal(a *types.Array) error {
	res := make(Indexes, a.Len())

	iter := a.Iterator()
	defer iter.Close()

	for {
		i, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		index := v.(*types.Document)

		keyDoc := must.NotFail(index.Get("key")).(*types.Document)
		fields := keyDoc.Keys()
		orders := keyDoc.Values()
		key := make([]IndexKeyPair, keyDoc.Len())

		for j, f := range fields {
			descending := false
			if orders[j].(int32) == -1 {
				descending = true
			}

			key[j] = IndexKeyPair{
				Field:      f,
				Descending: descending,
			}
		}

		// it was possible for it to be null in pgdb
		v, _ = index.Get("unique")
		unique, _ := v.(bool)

		res[i] = IndexInfo{
			Name:    must.NotFail(index.Get("name")).(string),
			PgIndex: must.NotFail(index.Get("pgindex")).(string),
			Key:     key,
			Unique:  unique,
		}
	}

	*s = res

	return nil
}

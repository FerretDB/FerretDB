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
	"runtime"
	"runtime/pprof"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// queryIteratorProfiles keeps track on all query iterators.
var queryIteratorProfiles = pprof.NewProfile("github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb.queryIterator")

type queryIterator struct {
	iter   driver.Iterator
	schema *tjson.Schema

	stack []byte
	n     int
}

func newQueryIterator(tigrisIter driver.Iterator, schema *tjson.Schema) iterator.Interface[int, *types.Document] {
	iter := &queryIterator{
		iter:   tigrisIter,
		schema: schema,
		stack:  debugbuild.Stack(),
	}

	queryIteratorProfiles.Add(iter, 1)

	runtime.SetFinalizer(iter, func(iter *queryIterator) {
		msg := "queryIterator.Close() has not been called"
		if iter.stack != nil {
			msg += "\nqueryIterator created by " + string(iter.stack)
		}

		panic(msg)
	})

	return iter
}

// Next implements iterator.Interface.
func (iter *queryIterator) Next() (int, *types.Document, error) {
	var document driver.Document

	ok := iter.iter.Next(&document)
	if !ok {
		return 0, nil, iterator.ErrIteratorDone
	}

	doc, err := tjson.Unmarshal(document, iter.schema)
	if err != nil {
		return 0, nil, err
	}

	iter.n++

	return iter.n, doc.(*types.Document), nil
}

// Close implements iterator.Interface.
func (iter *queryIterator) Close() {
	iter.iter.Close()
}

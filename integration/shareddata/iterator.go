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
// See the License for the specific language governing permissions and limitations under the License.

package shareddata

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/resource"
	"go.mongodb.org/mongo-driver/bson"
)

func newBenchmarkIterator(generator func() bson.D) iterator.Interface[struct{}, bson.D] {
	iter := &benchmarkIterator{
		token: resource.NewToken(),
		gen:   generator,
		hash:  sha256.New(),
	}

	resource.Track(iter, iter.token)

	return iter
}

type benchmarkIterator struct {
	gen func() bson.D

	m sync.Mutex

	token *resource.Token

	hash hash.Hash
}

func (iter *benchmarkIterator) Next() (struct{}, bson.D, error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	var unused struct{}

	doc := iter.gen()
	if doc == nil {
		// to avoid context cancellation changing the next `Next()` error
		// from `iterator.ErrIteratorDone` to `context.Canceled`
		iter.close()

		return unused, nil, iterator.ErrIteratorDone
	}

	rawDoc := []byte(fmt.Sprintf("%x", doc))
	currHash := iter.hash.Sum(rawDoc)

	iter.hash.Reset()

	_, err := iter.hash.Write(currHash)
	if err != nil {
		panic("Unexpected error: " + err.Error())
	}

	return unused, doc, nil
}

func (iter *benchmarkIterator) Hash() string {
	sum := iter.hash.Sum([]byte{})

	return base64.StdEncoding.EncodeToString(sum)
}

func (iter *benchmarkIterator) Close() {
	iter.m.Lock()
	defer iter.m.Unlock()

	iter.close()
}

func (iter *benchmarkIterator) close() {
	//if iter.rows != nil {
	//	iter.rows.Close()
	//	iter.rows = nil
	//}

	resource.Untrack(iter, iter.token)
}

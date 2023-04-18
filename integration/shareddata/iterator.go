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

package shareddata

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"sync"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// newValuesIterator creates iterator that iterates through bson.D documents generated
// by generator function.
// Generator should return next deterministic bson.D document on every execution.
// To stop iterator, generator must return nil.
func newValuesIterator(generator func() bson.D) *valuesIterator {
	iter := &valuesIterator{
		token:     resource.NewToken(),
		generator: generator,
		hash:      sha256.New(),
	}

	resource.Track(iter, iter.token)

	return iter
}

// valuesIterator iterates through bson.D documents created by generator function.
// It also calculates the checksum of all documents on-fly.
type valuesIterator struct {
	generator func() bson.D
	token     *resource.Token
	hash      hash.Hash

	// m Mutex protects generator function, and hash from parallel calls.
	// It's not placed on top of struct fields because of field alignment.
	m sync.Mutex
}

// Next implements iterator.Interface.
func (iter *valuesIterator) Next() (struct{}, bson.D, error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	var unused struct{}

	if iter.generator == nil {
		return unused, nil, iterator.ErrIteratorDone
	}

	doc := iter.generator()
	if doc == nil {
		return unused, nil, iterator.ErrIteratorDone
	}

	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return unused, nil, err
	}

	// write json representation of document to calculate hash later
	if _, err := iter.hash.Write(jsonDoc); err != nil {
		return unused, doc, err
	}

	return unused, doc, nil
}

// Close implements iterator.Interface.
func (iter *valuesIterator) Close() {
	iter.m.Lock()
	defer iter.m.Unlock()

	resource.Untrack(iter, iter.token)
	iter.generator = nil
}

// Hash returns calculated checksum of all returned by valuesIterator documents.
// It must be called after closing iterator, otherwise it returns the error.
func (iter *valuesIterator) Hash() (string, error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	if iter.generator != nil {
		return "", errors.New("Hash needs to be called on closed iterator")
	}

	sum := iter.hash.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum), nil
}

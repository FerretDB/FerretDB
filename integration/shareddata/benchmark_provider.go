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
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// BenchmarkProvider is implemented by shared data sets that provide documents for benchmarks.
type BenchmarkProvider interface {
	// Name returns benchmark provider name.
	Name() string

	// NewIterator returns a new iterator for documents in the same order.
	NewIterator() iterator.Interface[struct{}, bson.D]
}

// hashBenchmarkProvider checks that BenchmarkProvider's NewIterator methods returns a new iterator
// for the same documents in the same order,
// and returns a hash of those documents that could be used as a part of benchmark name.
func hashBenchmarkProvider(bp BenchmarkProvider) string {
	iter1 := bp.NewIterator()
	defer iter1.Close()

	iter2 := bp.NewIterator()
	defer iter2.Close()

	h := sha256.New()

	for {
		_, v1, err := iter1.Next()
		switch {
		case err == nil:
			_, v2, err := iter2.Next()
			if err != nil {
				panic(err)
			}

			must.BeTrue(reflect.DeepEqual(v1, v2))

			b := must.NotFail(json.Marshal(v1))
			h.Write(b)

		case errors.Is(err, iterator.ErrIteratorDone):
			_, _, err = iter2.Next()
			if !errors.Is(err, iterator.ErrIteratorDone) {
				panic("iter2 should be done too")
			}

			return hex.EncodeToString(h.Sum(nil)[:8])

		default:
			panic(err)
		}
	}
}

// generatorFunc is a function that returns the next generated bson.D document, or nil.
//
// The order of documents returned must be deterministic.
type generatorFunc func() bson.D

// newGeneratorFunc returns a new generatorFunc.
//
// All returned functions should be independent from each other, but return the same documents in the same order.
type newGeneratorFunc func() generatorFunc

// generatorBenchmarkProvider uses generator functions to implement BenchmarkProvider.
type generatorBenchmarkProvider struct {
	baseName         string
	newGeneratorFunc newGeneratorFunc
	hash             string
}

// newGeneratorBenchmarkProvider returns BenchmarkProvider with a given base name and newGeneratorFunc.
func newGeneratorBenchmarkProvider(baseName string, newGeneratorFunc newGeneratorFunc) BenchmarkProvider {
	gbp := &generatorBenchmarkProvider{
		baseName:         baseName,
		newGeneratorFunc: newGeneratorFunc,
	}

	gbp.hash = hashBenchmarkProvider(gbp)

	return gbp
}

func (gbp *generatorBenchmarkProvider) Name() string {
	return gbp.baseName + "/" + gbp.hash
}

func (gbp *generatorBenchmarkProvider) NewIterator() iterator.Interface[struct{}, bson.D] {
	var unused struct{}
	next := gbp.newGeneratorFunc()

	f := func() (struct{}, bson.D, error) {
		v := next()
		if v == nil {
			return unused, nil, iterator.ErrIteratorDone
		}

		return unused, v, nil
	}

	return iterator.ForFunc(f)
}

// check interfaces
var (
	_ BenchmarkProvider = (*generatorBenchmarkProvider)(nil)
)

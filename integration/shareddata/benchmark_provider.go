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
	"fmt"
	"iter"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// BenchmarkProvider is implemented by shared data sets that provide documents for benchmarks.
type BenchmarkProvider interface {
	// Name returns full benchmark provider name.
	Name() string

	// BaseName returns a part of the full name that does not include a number of documents and their hash.
	BaseName() string

	// NewIter returns a new iterator for the same documents.
	NewIter() iter.Seq[bson.D]
}

// BenchmarkGenerator provides documents for benchmarks by generating them.
type BenchmarkGenerator interface {
	// Init sets the number of documents to generate.
	Init(docs int)

	BenchmarkProvider
}

// hashBenchmarkProvider checks that BenchmarkProvider's NewIter methods returns a new iterator
// for the same documents in the same order,
// and returns a hash of those documents that could be used as a part of benchmark name.
func hashBenchmarkProvider(bp BenchmarkProvider) string {
	next, stop := iter.Pull(bp.NewIter())
	defer stop()

	h := sha256.New()

	for v1 := range bp.NewIter() {
		v2, ok := next()
		must.BeTrue(ok)
		must.BeTrue(reflect.DeepEqual(v1, v2))

		b := must.NotFail(bson.MarshalExtJSON(v1, true, false))
		h.Write(b)
	}

	_, ok := next()
	must.BeZero(ok)

	return hex.EncodeToString(h.Sum(nil)[:2])
}

// newGen returns a function that generates the next bson.D document, or nil.
//
// Returned functions should be independent from each other, but return the same documents in the same order.
type newGen func(docs int) func() bson.D

// generatorBenchmarkProvider implements BenchmarkProvider.
type generatorBenchmarkProvider struct {
	baseName string
	newGen   newGen
	docs     int
}

// newGeneratorBenchmarkProvider returns BenchmarkProvider.
func newGeneratorBenchmarkProvider(baseName string, newGen newGen) BenchmarkProvider {
	return &generatorBenchmarkProvider{
		baseName: baseName,
		newGen:   newGen,
	}
}

// Init implements [BenchmarkGenerator].
func (gbp *generatorBenchmarkProvider) Init(docs int) {
	gbp.docs = docs
}

// Name implements [BenchmarkProvider].
func (gbp *generatorBenchmarkProvider) Name() string {
	hash := hashBenchmarkProvider(gbp)

	return fmt.Sprintf("%s/Docs%d/%s", gbp.baseName, gbp.docs, hash)
}

// BaseName implements [BenchmarkProvider].
func (gbp *generatorBenchmarkProvider) BaseName() string {
	return gbp.baseName
}

// NewIter implements [BenchmarkProvider].
func (gbp *generatorBenchmarkProvider) NewIter() iter.Seq[bson.D] {
	if gbp.docs == 0 {
		panic("A number of documents must be more than zero")
	}

	g := gbp.newGen(gbp.docs)

	return func(yield func(bson.D) bool) {
		for {
			v := g()
			if v == nil {
				return
			}

			if !yield(v) {
				return
			}
		}
	}
}

// check interfaces
var (
	_ BenchmarkProvider  = (*generatorBenchmarkProvider)(nil)
	_ BenchmarkGenerator = (*generatorBenchmarkProvider)(nil)
)

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
	"iter"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// BenchmarkProvider is implemented by shared data sets that provide documents for benchmarks.
type BenchmarkProvider interface {
	// baseName returns a part of the full name that does not include a number of documents and their hash.
	baseName() string

	// Name returns full benchmark provider name.
	Name() string

	// Docs returns a new iterator for the same documents.
	// All calls should return the same set of documents in the same order.
	// All sequences should be independent from each other.
	Docs() iter.Seq[any]
}

// hashBenchmarkProvider checks that BenchmarkProvider's Docs methods returns a new iterator
// for the same documents in the same order,
// and returns a hash of those documents that could be used as a part of benchmark name.
func hashBenchmarkProvider(bp BenchmarkProvider) string {
	next, stop := iter.Pull(bp.Docs())
	defer stop()

	h := sha256.New()

	for v1 := range bp.Docs() {
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

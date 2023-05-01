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
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"go.mongodb.org/mongo-driver/bson"
)

// BenchmarkProvider is implemented by shared data sets that provide documents for benchmarks.
// It also calculates checksum of all provided documents.
type BenchmarkProvider interface {
	// Name returns benchmark provider name.
	Name() string

	// Hash returns actual hash of all provider documents.
	// It should be called after closing iterator.
	Hash() string

	// Docs returns iterator that returns all documents from provider.
	// They should be always in deterministic order.
	// The iterator calculates the checksum of all documents on go.
	Docs() iterator.Interface[struct{}, bson.D]
}

// benchmarkValues returns shared data documents for benchmark in deterministic order.
type benchmarkValues struct {
	// iter returns all documents in deterministic order.
	iter *valuesIterator

	// name represents the name of the benchmark values set.
	name string
}

// Name implements BenchmarkProvider interface.
func (b *benchmarkValues) Name() string {
	return b.name
}

// Hash implements BenchmarkProvider interface.
// It returns actual hash of all documents produced by BenchmarkValues.
// It will panic if iterator was not closed.
func (b *benchmarkValues) Hash() string {
	return must.NotFail(b.iter.Hash())
}

// Docs implements BenchmarkProvider interface.
func (b *benchmarkValues) Docs() iterator.Interface[struct{}, bson.D] {
	return b.iter
}

// check interfaces
var (
	_ BenchmarkProvider = (*benchmarkValues)(nil)
)

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
	"fmt"
	"iter"
)

// Generator implements BenchmarkProvider by generating documents.
type Generator struct {
	newGen   func(n int) iter.Seq[any]
	name     string
	fullName string
	docs     int
}

// baseName implements [BenchmarkProvider].
func (g *Generator) baseName() string {
	return g.name
}

// Name implements [BenchmarkProvider].
func (g *Generator) Name() string {
	if g.fullName != "" {
		return g.fullName
	}

	hash := hashBenchmarkProvider(g)
	g.fullName = fmt.Sprintf("%s/Docs%d/%s", g.name, g.docs, hash)

	return g.fullName
}

// Docs implements [BenchmarkProvider].
func (g *Generator) Docs() iter.Seq[any] {
	if g.docs <= 0 {
		panic("A number of documents must be more than zero")
	}

	return g.newGen(g.docs)
}

// Init sets a number of documents to generate.
func (g *Generator) Init(docs int) {
	g.docs = docs
}

// check interfaces
var (
	_ BenchmarkProvider = (*Generator)(nil)
)

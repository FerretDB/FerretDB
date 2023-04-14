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

// Package shareddata provides data for tests and benchmarks.
package shareddata

import "go.mongodb.org/mongo-driver/bson"

// SimpleBenchmarkValues contain documents with various types of values.
var SimpleBenchmarkValues BenchmarkValues = NewBenchmarkValues(
	"Ha59TmC9nwTXqEob73HIwqWUcx2Lp7aengCtNsqjWK0=",
	generateMixedValues(),
)

// generateMixedValues returns generator that generates deterministic set
// of 400 documents that contains string, double, and object types.
func generateMixedValues() func() bson.D {
	values := []any{
		"foo", 42, "42", bson.D{{"42", "hello"}},
	}
	valuesLen := len(values)

	i := 0

	gen := func() bson.D {
		if i >= 400 {
			return nil
		}

		v := values[i%valuesLen]

		doc := bson.D{{"_id", i}, {"v", v}}
		i++

		return doc
	}

	return gen
}

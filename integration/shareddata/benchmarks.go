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
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/maps"
)

// BenchmarkSmallDocuments provides documents that look like:
//
//	{_id: int32(0), id: int32(0), v: "foo"}
//	{_id: int32(1), id: int32(1), v: int32(42)}
//	{_id: int32(2), id: int32(2), v: "42"}
//	{_id: int32(3), id: int32(3), v: {"foo": int32(42)}}
//	...
//
// `_id` is an int32 primary key that starts from 0.
// `id` has the same value as `_id`, but is not indexed by default.
// `v` has one of the four values shown above.
var BenchmarkSmallDocuments = newGeneratorBenchmarkProvider("SmallDocuments", func(docs int) generatorFunc {
	values := []any{
		"foo", int32(42), "42", bson.D{{"foo", int32(42)}},
	}
	l := len(values)

	var total int

	return func() bson.D {
		if total >= docs {
			return nil
		}

		doc := bson.D{
			{"_id", int32(total)},
			{"id", int32(total)},
			{"v", values[total%l]},
		}

		total++

		return doc
	}
})

// BenchmarkSettingsDocuments provides large documents with 100 fields of various types.
//
// It simulates a settings document like the one FastNetMon uses.
var BenchmarkSettingsDocuments = newGeneratorBenchmarkProvider("SettingsDocuments", func(docs int) generatorFunc {
	var total int
	f := newFaker()

	return func() bson.D {
		if total >= docs {
			return nil
		}

		doc := make(bson.D, 100)
		doc[0] = bson.E{"_id", f.ObjectID()}
		for i := 1; i < len(doc); i++ {
			doc[i] = bson.E{
				Key:   f.FieldName(),
				Value: f.ScalarValue(),
			}
		}

		total++

		return doc
	}
})

// AllBenchmarkProviders returns all benchmark providers in random order.
func AllBenchmarkProviders() []BenchmarkProvider {
	providers := []BenchmarkProvider{
		BenchmarkSmallDocuments,
		BenchmarkSettingsDocuments,
	}

	// check that bse names are unique and randomize order
	res := make(map[string]BenchmarkProvider, len(providers))

	for _, p := range providers {
		n := p.BaseName()
		if _, ok := res[n]; ok {
			panic("duplicate benchmark provider base name: " + n)
		}

		res[n] = p
	}

	return maps.Values(res)
}

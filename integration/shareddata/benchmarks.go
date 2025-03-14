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
	"iter"

	"go.mongodb.org/mongo-driver/bson"
)

// BenchSmall provides documents that look like:
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
var BenchSmall = &Generator{
	name: "Small",
	newGen: func(n int) iter.Seq[any] {
		values := []any{
			"foo", int32(42), "42", bson.D{{"foo", int32(42)}},
		}
		l := len(values)

		return func(yield func(any) bool) {
			for i := range n {
				doc := bson.D{
					{"_id", int32(i)},
					{"id", int32(i)},
					{"v", values[i%l]},
				}

				if !yield(doc) {
					return
				}
			}
		}
	},
}

// BenchSettings provides documents with 100 fields of various types.
//
// It simulates a settings document like the one FastNetMon uses.
// `_id` is an int32 primary key that starts from 0.
var BenchSettings = &Generator{
	name: "Settings",
	newGen: func(n int) iter.Seq[any] {
		f := newFaker()

		return func(yield func(any) bool) {
			for i := range n {
				doc := make(bson.D, 100)
				doc[0] = bson.E{"_id", int32(i)}
				for e := 1; e < len(doc); e++ {
					doc[e] = bson.E{
						Key:   f.FieldName(),
						Value: f.ScalarValue(),
					}
				}

				if !yield(doc) {
					return
				}
			}
		}
	},
}

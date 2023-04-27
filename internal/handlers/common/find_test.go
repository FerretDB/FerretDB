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

package common

import (
	"testing"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func BenchmarkGetFindParams(b *testing.B) {
	doc := must.NotFail(types.NewDocument(
		"$db", "test",
		"collection", "test",
		"filter", must.NotFail(types.NewDocument("a", "b")),
		"sort", must.NotFail(types.NewDocument("a", "b")),
		"projection", must.NotFail(types.NewDocument("a", "b")),
		"skip", int64(123),
		"limit", int64(123),
		"batchSize", int64(484),
		"singleBatch", false,
		"comment", "123",
		"maxTimeMS", int64(123),
	))

	l := zap.NewNop()

	var err error

	for i := 0; i < b.N; i++ {
		var params = &FindParams{}

		params, err = GetFindParams(doc, l)
		if err != nil {
			b.Fatal(err)
		}

		_ = params.DB
	}
}

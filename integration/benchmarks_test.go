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

package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func BenchmarkQuery(b *testing.B) {
	provider := shareddata.MixedBenchmarkValues

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: provider,
	})

	ctx, coll := s.Ctx, s.Collection

	for name, bm := range map[string]struct {
		filter bson.D
	}{
		"String": {
			filter: bson.D{{"v", "foo"}},
		},
		"DotNotation": {
			filter: bson.D{{"v.42", "hello"}},
		},
	} {
		b.Run(fmt.Sprint(name, "_", provider.Hash()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cur, err := coll.Find(ctx, bm.filter)
				require.NoError(b, err)

				var res []bson.D
				require.NoError(b, cur.All(ctx, &res))
			}
		})
	}
}

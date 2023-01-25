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
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestEmbedded(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	testcases := map[string]struct {
		postgreSQLURL string
		isTLS         bool
	}{
		"TLS": {
			postgreSQLURL: "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
			isTLS:         true,
		},
		"TCP": {
			postgreSQLURL: "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
			isTLS:         false,
		},
	}

	for name, tc := range testcases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			port := 0
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				Flags: setup.Flags{
					TargetTLS:     &tc.isTLS,
					TargetPort:    &port,
					PostgreSQLURL: &tc.postgreSQLURL,
				},
			})

			ctx, collection := s.Ctx, s.Collection

			_, err := collection.Database().ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)
		})
	}
}

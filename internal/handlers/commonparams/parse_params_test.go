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

package commonparams

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		doc        *types.Document
		wantParams any
		wantErr    error
	}{
		"simple": {
			doc: must.NotFail(types.NewDocument(
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
			)),
			wantParams: common.FindParams{
				DB:          "test",
				Collection:  "test",
				Filter:      must.NotFail(types.NewDocument("a", "b")),
				Sort:        must.NotFail(types.NewDocument("a", "b")),
				Projection:  must.NotFail(types.NewDocument("a", "b")),
				Skip:        123,
				Limit:       123,
				BatchSize:   484,
				SingleBatch: false,
				Comment:     "123",
				MaxTimeMS:   123,
			},
			wantErr: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			params := common.FindParams{}

			err := Unmarshal(tt.doc, &params, zap.NewNop())
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr, err)
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantParams, params)
		})
	}
}

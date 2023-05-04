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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestParse(t *testing.T) {
	type findParams struct { //nolint:vet // this is a test struct
		DB          string          `name:"$db"`
		Collection  string          `name:"collection"`
		Filter      *types.Document `name:"filter,opt" error:"badValue"`
		Sort        *types.Document `name:"sort,opt"`
		Projection  *types.Document `name:"projection,opt"`
		Skip        int64           `name:"skip"`
		Limit       int64           `name:"limit"`
		BatchSize   int64           `name:"batchSize"`
		SingleBatch bool            `name:"singleBatch"`
		Comment     string          `name:"comment"`
		MaxTimeMS   int64           `name:"maxTimeMS"`

		ReturnKey           bool `name:"returnKey,non-default"`
		ShowRecordId        bool `name:"showRecordId,non-default"`
		Tailable            bool `name:"tailable,non-default"`
		OplogReplay         bool `name:"oplogReplay,non-default"`
		NoCursorTimeout     bool `name:"noCursorTimeout,non-default"`
		AwaitData           bool `name:"awaitData,non-default"`
		AllowPartialResults bool `name:"allowPartialResults,non-default"`

		Collation any `name:"collation,unimplemented"`
		Let       any `name:"let,unimplemented"`

		AllowDiskUse any `name:"allowDiskUse,ignored"`
		ReadConcern  any `name:"readConcern,ignored"`
		Max          any `name:"max,ignored"`
		Min          any `name:"min,ignored"`
	}

	tests := map[string]struct {
		doc        *types.Document
		wantParams any
		wantErr    error
	}{
		"simple": {
			doc: must.NotFail(types.NewDocument(
				"$db", "test",
				"find", "test",
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
			wantParams: findParams{
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
			params := findParams{}

			err := Unmarshal(tt.doc, "find", &params, zap.NewNop())
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr, err)
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantParams, params)
		})
	}
}

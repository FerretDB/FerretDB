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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		doc        *types.Document
		params     any
		wantParams any
		wantErr    error
	}{
		"simple": {
			doc: must.NotFail(types.NewDocument("$db", "test")),
			params: FindParams{
				DB: "test",
			},
			wantParams: FindParams{
				DB: "test",
			},
			wantErr: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := Unmarshal(tt.doc, tt.params)
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr, err)
			}

			require.Equal(t, tt.wantParams, tt.params)
		})
	}
}

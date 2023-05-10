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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestParse(t *testing.T) {
	type allTagsThatPass struct {
		DB           string          `ferretdb:"$db"`
		Collection   string          `ferretdb:"collection"`
		Filter       *types.Document `ferretdb:"filter,opt"`
		AllowDiskUse any             `ferretdb:"allowDiskUse,ignored"`
	}

	type noTag struct {
		Find string
	}

	type unimplementedTag struct {
		Find string `ferretdb:"find,unimplemented"`
	}

	type nonDefaultTag struct {
		Find bool `ferretdb:"find,non-default"`
	}

	tests := map[string]struct {
		doc        *types.Document
		command    string
		params     any
		wantParams any
		wantErr    error
	}{
		"AllTagTypesThatPass": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"$db", "test",
				"find", "test",
				"filter", must.NotFail(types.NewDocument("a", "b")),
				"allowDiskUse", "123",
			)),
			params: &allTagsThatPass{},
			wantParams: &allTagsThatPass{
				DB:         "test",
				Collection: "test",
				Filter:     must.NotFail(types.NewDocument("a", "b")),
			},
			wantErr: nil,
		},
		"UnimplementedTag": {
			command: "command",
			doc: must.NotFail(types.NewDocument(
				"find", "test",
			)),
			params: &unimplementedTag{},
			wantParams: &unimplementedTag{
				Find: "test",
			},
			wantErr: errors.New("NotImplemented (238): find: support for field \"find\" with value test is not implemented yet"),
		},
		"NonDefaultTag": {
			command: "command",
			doc: must.NotFail(types.NewDocument(
				"find", true,
			)),
			params:     &nonDefaultTag{},
			wantParams: &nonDefaultTag{},
			wantErr:    errors.New("NotImplemented (238): find: support for field \"find\" with non-default value true is not implemented yet"),
		},
		"EmptyTag": {
			command:    "count",
			doc:        must.NotFail(types.NewDocument("find", "test")),
			params:     &noTag{},
			wantParams: &noTag{},
			wantErr:    errors.New("[extract_params.go:80 commonparams.ExtractParams] unexpected field 'find' encountered"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ExtractParams(tt.doc, tt.command, tt.params, zap.NewNop())
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantParams, tt.params)
		})
	}
}

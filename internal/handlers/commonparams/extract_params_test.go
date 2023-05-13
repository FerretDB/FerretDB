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
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestParse(t *testing.T) {
	type allTagsThatPass struct { //nolint:vet // it's a test struct
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

	type update struct {
		Filter *types.Document `ferretdb:"q,opt"`
	}

	type updates struct {
		Update []update `ferretdb:"updates"`
	}

	type updateAny struct {
		Update any `ferretdb:"u"`
	}

	type numericBool struct {
		Find bool `ferretdb:"f,numericBool"`
	}

	type strict struct {
		Find int64 `ferretdb:"f,strict"`
	}

	type positive struct {
		Find int64 `ferretdb:"f,positive"`
	}

	tests := map[string]struct { //nolint:vet // it's a test table
		doc        *types.Document
		command    string
		params     any
		wantParams any
		wantErr    string
	}{
		"AllTagTypesThatPass": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"$db", "test",
				"find", "test",
				"filter", must.NotFail(types.NewDocument("a", "b")),
				"allowDiskUse", "123",
			)),
			params: new(allTagsThatPass),
			wantParams: &allTagsThatPass{
				DB:         "test",
				Collection: "test",
				Filter:     must.NotFail(types.NewDocument("a", "b")),
			},
		},
		"UnimplementedTag": {
			command: "command",
			doc: must.NotFail(types.NewDocument(
				"find", "test",
			)),
			params:  new(unimplementedTag),
			wantErr: "support for field \"find\" with value test is not implemented yet",
		},
		"NonDefaultTag": {
			command: "command",
			doc: must.NotFail(types.NewDocument(
				"find", true,
			)),
			params: new(nonDefaultTag),
			wantErr: "support for field \"find\"" +
				" with non-default value true is not implemented yet",
		},
		"EmptyTag": {
			command: "count",
			doc:     must.NotFail(types.NewDocument("find", "test")),
			params:  new(noTag),
			wantErr: "unexpected field 'find' encountered",
		},
		"ExtraFieldPassed": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"$db", "test",
				"find", "test",
				"extra", "field",
			)),
			params:  new(allTagsThatPass),
			wantErr: "unexpected field 'extra' encountered",
		},
		"MissingRequiredField": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"$db", "test",
			)),
			params:  new(allTagsThatPass),
			wantErr: "required field is not populated",
		},
		"ArrayTag": {
			command: "update",
			doc: must.NotFail(types.NewDocument(
				"updates", must.NotFail(types.NewArray(
					must.NotFail(types.NewDocument(
						"q", must.NotFail(types.NewDocument("a", "b")),
					)),
				)),
			)),
			params: new(updates),
			wantParams: &updates{
				Update: []update{
					{
						Filter: must.NotFail(types.NewDocument("a", "b")),
					},
				},
			},
		},
		"ParseDocumentIntoAnyTag": {
			command: "update",
			doc: must.NotFail(types.NewDocument(
				"u", must.NotFail(types.NewDocument("a", "b")),
			)),
			params: new(updateAny),
			wantParams: &updateAny{
				Update: must.NotFail(types.NewDocument("a", "b")),
			},
		},
		"ParseArrayIntoAnyTag": {
			command: "update",
			doc: must.NotFail(types.NewDocument(
				"u", must.NotFail(types.NewArray("a", "b")),
			)),
			params: new(updateAny),
			wantParams: &updateAny{
				Update: must.NotFail(types.NewArray("a", "b")),
			},
		},
		"ParseInt32IntoBoolTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", int32(1),
			)),
			params: new(numericBool),
			wantParams: &numericBool{
				Find: true,
			},
		},
		"ParseInt64IntoBoolTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", int64(1),
			)),
			params: new(numericBool),
			wantParams: &numericBool{
				Find: true,
			},
		},
		"ParseFloat64IntoBoolTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", 3.14,
			)),
			params: new(numericBool),
			wantParams: &numericBool{
				Find: true,
			},
		},
		"ParseStringIntoBoolTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", "true",
			)),
			params:  new(numericBool),
			wantErr: "field 'f' is the wrong type 'string', expected types '\\[bool, long, int, decimal, double\\]'",
		},
		"ParseStrictTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", 12.23,
			)),
			params: new(strict),
			wantParams: &strict{
				Find: 12,
			},
		},
		"ParseStrictTagWithWrongType": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", "12.23",
			)),
			params:  new(strict),
			wantErr: "field 'find.f' is the wrong type 'string', expected types '\\[long, int, decimal, double\\]'",
		},
		"ParsePositiveTag": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", int32(12),
			)),
			params: new(positive),
			wantParams: &positive{
				Find: 12,
			},
		},
		"ParsePositiveTagWithNegativeFloatValue": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", -12.23,
			)),
			params:  new(positive),
			wantErr: "f has non-integral value",
		},
		"ParsePositiveTagWithNegativeIntValue": {
			command: "find",
			doc: must.NotFail(types.NewDocument(
				"f", int32(-1),
			)),
			params:  new(positive),
			wantErr: "-1 value for f is out of range",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ExtractParams(tt.doc, tt.command, tt.params, zap.NewNop())
			if tt.wantErr != "" {
				require.Regexp(t, regexp.MustCompile(".*"+tt.wantErr), err.Error())
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantParams, tt.params)
		})
	}
}

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

package tjson

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestDocumentSchema(t *testing.T) {
	t.Parallel()

	lastUpdate := time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

	for name, tc := range map[string]struct { //nolint:paralleltest // false positive
		doc    *types.Document
		schema map[string]any
		err    error
	}{
		"doc": {
			doc: must.NotFail(types.NewDocument(
				"readOnly", false,
				"ok", float64(1),
				"regex", types.Regex{Pattern: "/[1-9]+/", Options: "g"},
				"timestamp", types.Timestamp(1652360855),
				"binary", types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
				"doc", must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
					"actor_id", float64(1.1),
					"first_name", "PENELOPE",
					"last_update", lastUpdate,
				)),
			)),
			schema: map[string]any{
				"$k":   []string{"type", "properties"},
				"type": "object",
				"properties": map[string]any{
					"readOnly": map[string]any{"type": "boolean"},
					"ok":       map[string]any{"type": "number"},
					"regex": map[string]any{
						"$k":   []string{"type", "properties"},
						"type": "object",
						"properties": map[string]any{
							"$r": map[string]any{"type": "string"},
							"o":  map[string]any{"type": "string"},
						},
					},
					"timestamp": map[string]any{
						"$k":   []string{"type", "properties"},
						"type": "object",
						"properties": map[string]any{
							"$t": map[string]any{"type": "string"},
						},
					},
					"binary": map[string]any{
						"$k":   []string{"type", "properties"},
						"type": "object",
						"properties": map[string]any{
							"$b": map[string]any{"type": "string", "format": "byte"},   // binary data
							"s":  map[string]any{"type": "integer", "format": "int32"}, // binary subtype
						},
					},
					"doc": map[string]any{
						"$k":   []string{"type", "properties"},
						"type": "object",
						"properties": map[string]any{
							"_id": map[string]any{
								"$k":   []string{"type", "properties"},
								"type": "object",
								"properties": map[string]any{
									"$o": map[string]any{"type": "string"},
								},
							},
							"actor_id":   map[string]any{"type": "number"},
							"first_name": map[string]any{"type": "string"},
							"last_update": map[string]any{
								"type":   "string",
								"format": "date-time",
							},
							"$k": []string{"_id", "actor_id", "first_name", "last_update"},
						},
					},
					"$k": []string{"readOnly", "ok", "regex", "timestamp", "binary", "doc"},
				},
			},
		},
		"array": {
			doc: must.NotFail(types.NewDocument(
				"array", must.NotFail(types.NewArray()),
			)),
			err: errors.New("arrays not supported yet"),
		},
		"int32": {
			doc: must.NotFail(types.NewDocument(
				"int32", int32(0),
			)),
			err: errors.New("int32 not supported yet"),
		},
		"int64": {
			doc: must.NotFail(types.NewDocument(
				"int64", int64(0),
			)),
			err: errors.New("int64 not supported yet"),
		},
		"nuls": {
			doc: must.NotFail(types.NewDocument(
				"null", types.Null,
			)),
			err: errors.New("cannot determine type"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := DocumentSchema(tc.doc)
			if tc.err != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.schema, actual)
		})
	}
}

func TestParseSchema(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:paralleltest // false positive
		raw    string
		schema map[string]any
		err    error
	}{
		"doc": {
			raw: `{ "name": "user", "description": "Collection of documents with details of users",` +
				`"properties": { "id": { "description": "A unique identifier for the user", "type": "string" },` +
				`"name": { "description": "Name of the user", "type": "string", "maxLength": 100 },` +
				`"active": { "description": "User account active", "type": "boolean" } }, "primary_key": ["id"] }`,
			schema: map[string]any{
				"id":     map[string]any{"type": "string"},
				"name":   map[string]any{"type": "string"},
				"active": map[string]any{"type": "boolean"},
			},
		},
		"array": {
			raw: `{ "name": "user", properties": { "id": {"type": "array" }}}`,
			err: errors.New("arrays not supported yet"),
		},
		"int": {
			raw: `{ "name": "user", properties": { "id": {"type": "integer" }}}`,
			err: errors.New("int32 not supported yet"),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseSchema([]byte(tc.raw))
			if tc.err != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.schema, actual)
		})
	}
}

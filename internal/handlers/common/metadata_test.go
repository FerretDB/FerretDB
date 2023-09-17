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

// Package common provides common code for all handlers.
package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

var (
	clientMetadataDocument = must.NotFail(types.NewDocument(
		"client", must.NotFail(types.NewDocument(
			"driver", must.NotFail(types.NewDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			)),
			"os", must.NotFail(types.NewDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			)),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", must.NotFail(types.NewDocument(
				"name", "mongosh 1.0.1",
			)),
		)),
	))

	noClientMetadataDocument = must.NotFail(types.NewDocument())
)

func TestCheckClientMetadata(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string][]struct { //nolint:vet // used for test only
		document *types.Document
		errMsg   string
	}{
		"NoClientMetadata": {{
			document: noClientMetadataDocument,
		}},
		"ClientMetadataDocument": {{
			document: clientMetadataDocument,
		}},
		"ClientMetadataDocumentAndNoClientMetadata": {
			{
				document: clientMetadataDocument,
			},
			{
				document: noClientMetadataDocument,
			},
		},
		"NoClientMetadataAndClientMetadataDocument": {
			{
				document: noClientMetadataDocument,
			},
			{
				document: clientMetadataDocument,
			},
		},
		"2xNoClientMetadata": {
			{
				document: noClientMetadataDocument,
			},
			{
				document: noClientMetadataDocument,
			},
		},
		"2xClientMetadataDocument": {
			{
				document: clientMetadataDocument,
			},
			{
				document: clientMetadataDocument,
				errMsg:   "ClientMetadataCannotBeMutated (186): The client metadata document may only be sent in the first hello",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			connInfo := conninfo.NewConnInfo()
			ctx = conninfo.WithConnInfo(ctx, connInfo)

			for _, test := range tc {
				err := CheckClientMetadata(ctx, test.document)

				if test.errMsg != "" {
					assert.EqualError(t, err, test.errMsg)
					return
				}

				if test.document == clientMetadataDocument {
					assert.True(t, connInfo.ClientMetadataPresence())
				}

				if name == "2xNoClientMetadata" || name == "NoClientMetadata" {
					assert.False(t, connInfo.ClientMetadataPresence())
				}

				assert.NoError(t, err)
			}
		})
	}
}

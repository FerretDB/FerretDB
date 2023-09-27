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
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestCheckClientMetadata(t *testing.T) {
	t.Parallel()

	metadata := must.NotFail(types.NewDocument(
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

	empty := must.NotFail(types.NewDocument())

	for name, tc := range map[string][]struct { //nolint:vet // used for test only
		document *types.Document
		err      error
		recv     bool
	}{
		"NoClientMetadata": {
			{document: empty},
		},
		"ClientMetadataDocument": {
			{document: metadata, recv: true},
		},
		"ClientMetadataDocumentAndNoClientMetadata": {
			{document: metadata, recv: true},
			{document: empty, recv: true},
		},
		"NoClientMetadataAndClientMetadataDocument": {
			{document: empty},
			{document: metadata, recv: true},
		},
		"2xNoClientMetadata": {
			{document: empty},
			{document: empty},
		},
		"2xClientMetadataDocument": {
			{document: metadata, recv: true},
			{
				document: metadata,
				err: commonerrors.NewCommandErrorMsg(
					commonerrors.ErrClientMetadataCannotBeMutated,
					"The client metadata document may only be sent in the first hello",
				),
				recv: true,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			connInfo := conninfo.New()
			ctx := conninfo.Ctx(context.Background(), connInfo)

			for _, test := range tc {
				err := CheckClientMetadata(ctx, test.document)
				assert.Equal(t, test.err, err)
				assert.Equal(t, test.recv, connInfo.MetadataRecv())
			}
		})
	}
}

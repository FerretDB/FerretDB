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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// CheckClientMetadata checks if the message does not contain client metadata after it was received already.
func CheckClientMetadata(ctx context.Context, doc *types.Document) error {
	c, _ := doc.Get("client")
	if c == nil {
		return nil
	}

	connInfo := conninfo.Get(ctx)
	if connInfo.MetadataRecv() {
		return commonerrors.NewCommandErrorMsg(
			commonerrors.ErrClientMetadataCannotBeMutated,
			"The client metadata document may only be sent in the first hello",
		)
	}

	connInfo.SetMetadataRecv()
	return nil
}

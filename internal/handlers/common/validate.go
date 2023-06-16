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

package common

import (
	"context"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Validate is a part of a common implementation of the validate command.
func Validate(ctx context.Context, msg *wire.OpMsg, l *zap.Logger) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	Ignored(document, l, "full", "repair", "metadata")

	command := document.Command()

	db, err := GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ns", db+"."+collection,
			"nInvalidDocuments", int32(0),
			"nNonCompliantDocuments", int32(0),
			"nrecords", int32(-1), // TODO
			"nIndexes", int32(1),
			// "keysPerIndex", TODO
			// "indexDetails", TODO
			"valid", true,
			"repaired", false,
			"warnings", types.MakeArray(0),
			"errors", types.MakeArray(0),
			"extraIndexEntries", types.MakeArray(0),
			"missingIndexEntries", types.MakeArray(0),
			"corruptRecords", types.MakeArray(0),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

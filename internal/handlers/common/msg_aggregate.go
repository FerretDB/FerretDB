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
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/stages"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Aggregate is a part of common implementation of the aggregate command.
func Aggregate(ctx context.Context, msg *wire.OpMsg, l *zap.Logger) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = Unimplemented(document, "explain", "cursor", "bypassDocumentValidation", "hint"); err != nil {
		return nil, err
	}
	if err = Unimplemented(document, "readConcern", "writeConcern"); err != nil {
		return nil, err
	}

	Ignored(document, l, "allowDiskUse", "maxTimeMS", "collation", "comment", "let")

	pipeline, err := GetRequiredParam[*types.Array](document, "pipeline")
	if err != nil {
		return nil, err
	}

	stages := make([]stages.Stage, pipeline.Len())
	iter := pipeline.Iterator()
	defer iter.Close()

	if pipeline.Len() > 0 {
		d := must.NotFail(pipeline.Get(0)).(*types.Document)

		return nil, NewCommandErrorMsgWithArgument(
			ErrNotImplemented,
			fmt.Sprintf("`aggregate` %q is not implemented yet", d.Command()),
			d.Command(),
		)
	}

	return nil, NewCommandErrorMsg(ErrNotImplemented, "`aggregate` command is not implemented yet")
}

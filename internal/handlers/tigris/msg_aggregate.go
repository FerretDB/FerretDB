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

package tigris

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	pipeline, err := common.GetRequiredParam[*types.Array](document, "pipeline")
	if err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1891

	if pipeline.Len() > 0 {
		d := must.NotFail(pipeline.Get(0)).(*types.Document)

		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf("`aggregate` %q is not implemented yet", d.Command()),
			d.Command(),
		)
	}

	return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrNotImplemented, "`aggregate` command is not implemented yet")
}

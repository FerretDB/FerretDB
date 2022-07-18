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

package pg

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgExplain implements HandlerInterface.
func (h *Handler) MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	verbosity := "allPlansExecution"
	if verbosity, err = common.GetOptionalParam(document, "verbosity", verbosity); err != nil {
		return nil, err
	}
	if verbosity == "executionStats" || verbosity == "serverParameters" {
		err = fmt.Errorf("verbosity: support for value %q is not implemented yet", verbosity)
		return nil, common.NewError(common.ErrNotImplemented, err)
	}

	var explain *types.Document
	if explain, err = common.GetRequiredParam[*types.Document](document, "explain"); err != nil {
		return nil, err
	}

	return nil, nil
}

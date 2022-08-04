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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate is a common implementation of the aggregate command.
func MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	m := conninfo.GetConnInfo(ctx).AggregationStages

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	pipeline, err := GetRequiredParam[*types.Array](document, "pipeline")
	if err != nil {
		return nil, err
	}

	for i := 0; i < pipeline.Len(); i++ {
		stage := must.NotFail(pipeline.Get(i)).(*types.Document)
		m.WithLabelValues(document.Command(), stage.Command()).Inc()
	}

	return nil, NewErrorMsg(ErrNotImplemented, "`aggregate` command is not implemented yet")
}

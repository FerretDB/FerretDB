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

package handler

import (
	"context"
	"strconv"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations/stages"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgBuildInfo implements `buildInfo` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgBuildInfo(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	aggregationStages := types.MakeArray(len(stages.Stages))
	for stage := range stages.Stages {
		aggregationStages.Append(stage)
	}

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(
			"version", version.Get().MongoDBVersion,
			"gitVersion", version.Get().Commit,
			"modules", must.NotFail(types.NewArray()),
			"sysInfo", "deprecated",
			"versionArray", version.Get().MongoDBVersionArray,
			"bits", int32(strconv.IntSize),
			"debug", version.Get().DebugBuild,
			"maxBsonObjectSize", int32(h.MaxBsonObjectSizeBytes),
			"buildEnvironment", version.Get().BuildEnvironment,

			// our extensions
			"ferretdbVersion", version.Get().Version,
			"ferretdbFeatures", must.NotFail(types.NewDocument(
				"aggregationStages", aggregationStages,
			)),

			"ok", float64(1),
		)),
	)
}

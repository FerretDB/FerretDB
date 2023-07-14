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

package commoncommands

import (
	"context"
	"strconv"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/stages"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgBuildInfo is a common implementation of the buildInfo command.
func MsgBuildInfo(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
	aggregationStages := types.MakeArray(len(stages.Stages))
	for stage := range stages.Stages {
		aggregationStages.Append(stage)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"version", version.Get().MongoDBVersion,
			"gitVersion", version.Get().Commit,
			"modules", must.NotFail(types.NewArray()),
			"sysInfo", "deprecated",
			"versionArray", version.Get().MongoDBVersionArray,
			"bits", int32(strconv.IntSize),
			"debug", version.Get().DebugBuild,
			"maxBsonObjectSize", int32(types.MaxDocumentLen),
			"buildEnvironment", version.Get().BuildEnvironment,

			// our extensions
			"ferretdbVersion", version.Get().Version,
			"ferretdbFeatures", must.NotFail(types.NewDocument(
				"aggregationStages", aggregationStages,
			)),

			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

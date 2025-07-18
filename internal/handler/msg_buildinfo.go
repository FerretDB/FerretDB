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
	"maps"
	"slices"
	"strconv"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgBuildInfo implements `buildInfo` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgBuildInfo(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	info := version.Get()

	buildEnvironment := wirebson.MakeDocument(len(info.BuildEnvironment))
	for _, k := range slices.Sorted(maps.Keys(info.BuildEnvironment)) {
		must.NoError(buildEnvironment.Add(k, info.BuildEnvironment[k]))
	}

	versionArray := wirebson.MakeArray(len(info.MongoDBVersionArray))
	for _, v := range info.MongoDBVersionArray {
		must.NoError(versionArray.Add(v))
	}

	return middleware.ResponseDoc(req, wirebson.MustDocument(
		"version", info.MongoDBVersion,
		"gitVersion", info.Commit,
		"modules", wirebson.MustArray(),
		"sysInfo", "deprecated",
		"versionArray", versionArray,
		"bits", int32(strconv.IntSize),
		"debug", info.DevBuild,
		"maxBsonObjectSize", maxBsonObjectSize,
		"buildEnvironment", buildEnvironment,

		// our extensions
		"ferretdb", wirebson.MustDocument(
			"version", info.Version,
			"package", info.Package,
		),

		"ok", float64(1),
	))
}

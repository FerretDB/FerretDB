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
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgHostInfo is a common implementation of the hostInfo command.
func MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	now := time.Now().UTC()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var osName, osVersion string

	if runtime.GOOS == "linux" {
		file, err := os.Open("/etc/os-release")
		if err != nil {
			file, err = os.Open("/etc/arch-release")
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		defer file.Close()

		osName, osVersion, err = parseOSRelease(file)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	os := "unknown"
	switch runtime.GOOS {
	case "linux":
		os = "Linux"
	case "darwin":
		os = "macOS"
	case "windows":
		os = "Windows"
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"system", must.NotFail(types.NewDocument(
				"currentTime", now,
				"hostname", hostname,
				"cpuAddrSize", int32(strconv.IntSize),
				"numCores", int32(runtime.NumCPU()),
				"cpuArch", runtime.GOARCH,
			)),
			"os", must.NotFail(types.NewDocument(
				"type", os,
				"name", osName,
				"version", osVersion,
			)),
			"extra", must.NotFail(types.NewDocument()),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

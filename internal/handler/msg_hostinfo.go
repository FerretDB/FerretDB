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
	"bufio"
	"context"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgHostInfo implements `hostInfo` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgHostInfo(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	now := time.Now().UTC()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var osName, osVersion string

	// try to parse Linux distro name and version, but do not fail if they are not present
	if runtime.GOOS == "linux" {
		file, err := os.Open("/etc/os-release")
		if err != nil {
			file, err = os.Open("/usr/lib/os-release")
		}

		if err == nil {
			defer file.Close() //nolint:errcheck // we are only reading it
			osName, osVersion, _ = parseOSRelease(file)
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

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(
			"system", must.NotFail(types.NewDocument(
				"currentTime", now,
				"hostname", hostname,
				"cpuAddrSize", int32(strconv.IntSize),
				"numCores", int32(runtime.GOMAXPROCS(-1)),
				"cpuArch", runtime.GOARCH,
			)),
			"os", must.NotFail(types.NewDocument(
				"type", os,
				"name", osName,
				"version", osVersion,
			)),
			"extra", must.NotFail(types.NewDocument()),
			"ok", float64(1),
		)),
	)
}

// parseOSRelease parses the /etc/os-release or /usr/lib/os-release file content,
// returning the OS name and version.
func parseOSRelease(r io.Reader) (string, string, error) {
	scanner := bufio.NewScanner(r)

	configParams := map[string]string{}

	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}

		if v, err := strconv.Unquote(value); err == nil {
			value = v
		}

		configParams[key] = value
	}

	if err := scanner.Err(); err != nil {
		return "", "", lazyerrors.Error(err)
	}

	return configParams["NAME"], configParams["VERSION"], nil
}

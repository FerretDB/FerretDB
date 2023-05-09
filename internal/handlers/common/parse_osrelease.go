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
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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

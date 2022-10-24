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

// Package hex provides helpers for working with hex dumps.
package hex

import (
	"bufio"
	"encoding/hex"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Dump makes a hex dump of byte array.
func Dump(b []byte) string {
	return hex.Dump(b)
}

// ParseDump decodes from hex dump to the byte array.
func ParseDump(s string) ([]byte, error) {
	var res []byte

	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(s)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line[len(line)-1] == '|' {
			// go dump
			line = strings.TrimSpace(line[8:60])
			line = strings.Join(strings.Split(line, " "), "")
		} else {
			// wireshark dump
			line = strings.TrimSpace(line[7:54])
			line = strings.Join(strings.Split(line, " "), "")
		}

		b, err := hex.DecodeString(line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		res = append(res, b...)
	}

	if err := scanner.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

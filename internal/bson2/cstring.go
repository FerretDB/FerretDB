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

package bson2

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func sizeCString(v string) int {
	return len(v) + 1
}

func encodeCString(b []byte, s string) {
	// ensure b length early
	b[len(s)] = 0

	copy(b, s)
}

func decodeCString(b []byte) (string, error) {
	if len(b) < 1 {
		return "", fmt.Errorf("decodeCString: expected at least 1 byte, got %d: %w", len(b), ErrDecodeShortInput)
	}

	var i int
	var v byte
	for i, v = range b {
		if v == 0 {
			break
		}
	}

	if v != 0 {
		return "", lazyerrors.Error(ErrDecodeInvalidInput)
	}

	return string(b[:i]), nil
}

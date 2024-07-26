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

package bson

import (
	"encoding/binary"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FindRaw finds the first raw BSON document or array in b and returns its length l.
// It should start from the first byte of b.
// RawDocument(b[:l] / RawArray(b[:l]) might not be valid. It is the caller's responsibility to check it.
//
// Use RawDocument(b) / RawArray(b) conversion instead if b contains exactly one document/array and no extra bytes.
func FindRaw(b []byte) (int, error) {
	bl := len(b)
	if bl < 5 {
		return 0, lazyerrors.Errorf("len(b) = %d: %w", bl, ErrDecodeShortInput)
	}

	dl := int(binary.LittleEndian.Uint32(b))
	if dl < 5 {
		return 0, lazyerrors.Errorf("dl = %d: %w", dl, ErrDecodeInvalidInput)
	}

	if bl < dl {
		return 0, lazyerrors.Errorf("len(b) = %d, dl = %d: %w", bl, dl, ErrDecodeShortInput)
	}

	if b[dl-1] != 0 {
		return 0, lazyerrors.Errorf("invalid last byte: %w", ErrDecodeInvalidInput)
	}

	return dl, nil
}

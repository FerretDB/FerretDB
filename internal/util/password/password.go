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

// Package password provides utilities for password hashing and verification.
package password

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// params represent Argon2id parameters.
type params struct {
	m uint32 // m / memory usage in KiB
	t uint32 // t / number of iterations
	p uint8  // p / parallelism / threads
}

// encodedRe is the regular expression used to parse encoded hashes.
var encodedRe = regexp.MustCompile(`^\$argon2id\$v=19\$m=(\d+),t=(\d+),p=(\d+)\$(\S+)\$(\S+)$`)

// encode converts given password hash, salt and parameters into a string
// compatible with a reference implementation and libsodium.
func encode(hash, salt []byte, params params) (encoded string) {
	encoded = fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.m, params.t, params.p,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return
}

// decode parses given encoded hash and returns the hash, salt and parameters.
//
// Returned error is safe for logging, but should not be returned to the client.
func decode(encoded string) (hash, salt []byte, params params, err error) {
	if len(encoded) > 100 {
		err = lazyerrors.Errorf("encoded hash is too long: %d", len(encoded))
		return
	}

	matches := encodedRe.FindStringSubmatch(encoded)
	if len(matches) != 6 {
		err = lazyerrors.Errorf("expected 6 matches, got %d", len(matches))
		return
	}

	v, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		err = lazyerrors.Errorf("failed to parse %q: %s", matches[1], err)
		return
	}
	params.m = uint32(v)

	v, err = strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		err = lazyerrors.Errorf("failed to parse %q: %s", matches[2], err)
		return
	}
	params.t = uint32(v)

	v, err = strconv.ParseUint(matches[3], 10, 8)
	if err != nil {
		err = lazyerrors.Errorf("failed to parse %q: %s", matches[3], err)
		return
	}
	params.p = uint8(v)

	if salt, err = base64.RawStdEncoding.DecodeString(matches[4]); err != nil {
		err = lazyerrors.Errorf("failed to decode salt %q: %s", matches[4], err)
		return
	}

	if hash, err = base64.RawStdEncoding.DecodeString(matches[5]); err != nil {
		err = lazyerrors.Errorf("failed to decode hash %q: %s", matches[5], err)
		return
	}

	return
}

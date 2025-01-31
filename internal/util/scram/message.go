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

package scram

import (
	"encoding/base64"
	"log/slog"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// message represents a parsed SCRAM message.
type message struct {
	gs2 string // gs2-cbind-flag,authzid
	n   string // username
	c   string // channel-binding
	r   string // nonce
	s   string // base64-encoded salt
	p   string // base64-encoded proof
	v   string // base64-encoded verifier
	i   int    // iteration-count
}

// String returns the string representation of the message.
func (m *message) String() string {
	parts := make([]string, 0, 2)

	if m.gs2 != "" {
		parts = append(parts, m.gs2)
	}

	if m.n != "" {
		n := strings.ReplaceAll(m.n, "=", "=3D")
		n = strings.ReplaceAll(n, ",", "=2C")
		parts = append(parts, "n="+n)
	}

	if m.c != "" {
		parts = append(parts, "c="+m.c)
	}

	if m.r != "" {
		parts = append(parts, "r="+m.r)
	}

	if m.s != "" {
		parts = append(parts, "s="+m.s)
	}

	if m.i != 0 {
		parts = append(parts, "i="+strconv.Itoa(m.i))
	}

	if m.p != "" {
		parts = append(parts, "p="+m.p)
	}

	if m.v != "" {
		parts = append(parts, "v="+m.v)
	}

	return strings.Join(parts, ",")
}

// isClientFirst returns true if the message is a client-first message.
func (m *message) isClientFirst() bool {
	ok := m.gs2 == "n,"
	ok = ok && m.n != ""
	ok = ok && m.c == ""
	ok = ok && m.r != ""
	ok = ok && m.s == ""
	ok = ok && m.i == 0
	ok = ok && m.p == ""
	ok = ok && m.v == ""

	return ok
}

// isServerFirst returns true if the message is a server-first message.
func (m *message) isServerFirst() bool {
	ok := m.gs2 == ""
	ok = ok && m.n == ""
	ok = ok && m.c == ""
	ok = ok && m.r != ""
	ok = ok && m.s != ""
	ok = ok && m.i != 0
	ok = ok && m.p == ""
	ok = ok && m.v == ""

	return ok
}

// isClientFinal returns true if the message is a client-final message.
func (m *message) isClientFinal() bool {
	ok := m.gs2 == ""
	ok = ok && m.n == ""
	ok = ok && m.c != ""
	ok = ok && m.r != ""
	ok = ok && m.s == ""
	ok = ok && m.i == 0
	ok = ok && m.p != ""
	ok = ok && m.v == ""

	return ok
}

// isServerFinal returns true if the message is a server-final message.
func (m *message) isServerFinal() bool {
	ok := m.gs2 == ""
	ok = ok && m.n == ""
	ok = ok && m.c == ""
	ok = ok && m.r == ""
	ok = ok && m.s == ""
	ok = ok && m.i == 0
	ok = ok && m.p == ""
	ok = ok && m.v != ""

	return ok
}

// base64Decode decodes a base64-encoded value, producing better messages on errors.
func base64Decode(field, value string) ([]byte, error) {
	if value == "" {
		return nil, lazyerrors.Errorf("empty SCRAM attribute: %q", field)
	}

	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, lazyerrors.Errorf("failed to decode base64 SCRAM attribute: %q: %w", field, err)
	}

	return b, nil
}

// parseMessage parses a SCRAM message.
func parseMessage(msg string, l *slog.Logger) (*message, error) {
	if !utf8.ValidString(msg) {
		return nil, lazyerrors.Errorf("invalid UTF-8: %q", msg)
	}

	var res message

	// better check for gs2 header
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/901
	if strings.HasPrefix(msg, "n,,") {
		msg = strings.TrimPrefix(msg, "n,,")
		res.gs2 = "n,"
	}

	fields := strings.Split(msg, ",")

	// https://datatracker.ietf.org/doc/html/rfc5802#section-5.1 says:
	// > Note that the order of attributes in client or server messages is fixed
	// We should enforce that.
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/901

	for _, field := range fields {
		name, value, ok := strings.Cut(field, "=")
		if !ok {
			return nil, lazyerrors.Errorf("malformed SCRAM attribute: %q", field)
		}

		// in order of https://datatracker.ietf.org/doc/html/rfc5802#section-5.1
		switch name {
		case "a":
			if value != "" {
				return nil, lazyerrors.Errorf("unsupported SCRAM attribute 'a': %q", field)
			}

		case "n":
			// SASLprep, check if = is not followed by either 2C or 3D
			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/901
			value = strings.ReplaceAll(value, "=2C", ",")
			value = strings.ReplaceAll(value, "=3D", "=")

			if value == "" {
				return nil, lazyerrors.Errorf("empty SCRAM attribute 'n': %q", field)
			}

			res.n = value

		case "m":
			return nil, lazyerrors.Errorf("unsupported SCRAM attribute 'm': %q", field)

		case "r":
			if len(value) < 16 {
				return nil, lazyerrors.Errorf("SCRAM attribute 'r' is too short: %q", field)
			}

			res.r = value

		case "c":
			if value != "biws" { // "n,," base64-encoded
				return nil, lazyerrors.Errorf("unsupported SCRAM attribute 'c': %q", field)
			}

			res.c = value

		case "s":
			s, err := base64Decode(field, value)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			sl := len(s)
			if sl < 12 {
				return nil, lazyerrors.Errorf("SCRAM attribute 's' has incorrect length %d: %q", sl, field)
			}

			if sl != 28 {
				// Legacy mongo shell and some older drivers require exactly 28 bytes,
				// but most current drivers are fine.
				// We still want users to use credentials created by DocumentDB.
				l.Warn(
					"SCRAM attribute 's' has unexpected length. It is recommended to use users created by `createUser` command.",
					slog.Int("l", sl),
				)
			}

			res.s = value

		case "i":
			i, err := strconv.Atoi(value)
			if err != nil {
				return nil, lazyerrors.Errorf("failed to parse SCRAM attribute 'i': %q: %w", field, err)
			}

			if i < 4096 {
				return nil, lazyerrors.Errorf("SCRAM attribute 'i' is too small: %q", field)
			}

			res.i = i

		case "p":
			if _, err := base64Decode(field, value); err != nil {
				return nil, lazyerrors.Error(err)
			}

			res.p = value

		case "v":
			if _, err := base64Decode(field, value); err != nil {
				return nil, lazyerrors.Error(err)
			}

			res.v = value

		// case "e":

		default:
			return nil, lazyerrors.Errorf("unsupported SCRAM attribute: %q", field)
		}
	}

	return &res, nil
}

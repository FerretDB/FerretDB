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

package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamed(t *testing.T) {
	var buf bytes.Buffer

	l := slog.New(NewHandler(&buf, &NewHandlerOpts{
		Base:         "console",
		RemoveTime:   true,
		RemoveLevel:  true,
		RemoveSource: true,
	}))

	for name, tc := range map[string]struct {
		name string
		msg  string

		expected string
	}{
		"Empty": {
			name:     "",
			msg:      "test",
			expected: `test	{"name":""}`,
		},
		"Named": {
			name:     "test-logger",
			msg:      "test",
			expected: `test	{"name":"test-logger"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			testLogger := Named(l, tc.name)
			testLogger.Info(tc.msg)

			assert.Equal(t, tc.expected+"\n", buf.String())
			buf.Reset()
		})
	}
}

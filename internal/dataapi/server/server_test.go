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

package server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wiretest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareRequest(t *testing.T) {
	for name, tc := range map[string]struct {
		pairs    []any
		expected *wirebson.Document
		err      error
	}{
		"Simple": {
			pairs:    []any{"foo", "bar"},
			expected: wirebson.MustDocument("foo", "bar"),
		},
		"RawMessage": {
			pairs:    []any{"foo", pointer.To(json.RawMessage(`{"foo":"bar"}`))},
			expected: wirebson.MustDocument("foo", wirebson.MustDocument("foo", "bar")),
		},
		"RawMessageArray": {
			pairs:    []any{"foo", pointer.To(json.RawMessage(`["foo","bar"]`))},
			expected: wirebson.MustDocument("foo", wirebson.MustArray("foo", "bar")),
		},
		"Float32": {
			pairs:    []any{"foo", pointer.To(float32(1))},
			expected: wirebson.MustDocument("foo", float64(1)),
		},
		"EmptyRawMessage": {
			pairs: []any{"foo", pointer.To(json.RawMessage{})},
			err:   fmt.Errorf("server.go:272 (server.prepareRequest): Invalid object: []"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := prepareRequest(tc.pairs...)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			wiretest.AssertEqual(t, tc.expected, actual.Document())
		})
	}
}

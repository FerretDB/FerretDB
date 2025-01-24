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
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareOpMsg(t *testing.T) {
	for name, tc := range map[string]struct {
		expectedOpMsg *wire.OpMsg
		expectedErr   error
		pairs         []any
	}{
		"Simple": {
			pairs:         []any{"foo", "bar"},
			expectedOpMsg: wire.MustOpMsg("foo", "bar"),
		},
		"RawMessage": {
			pairs:         []any{"foo", pointer.To(json.RawMessage(`{"foo":"bar"}`))},
			expectedOpMsg: wire.MustOpMsg("foo", wirebson.MustDocument("foo", "bar")),
		},
		"RawMessageArray": {
			pairs:         []any{"foo", pointer.To(json.RawMessage(`["foo","bar"]`))},
			expectedOpMsg: wire.MustOpMsg("foo", wirebson.MustArray("foo", "bar")),
		},
		"Float32": {
			pairs:         []any{"foo", pointer.To(float32(1))},
			expectedOpMsg: wire.MustOpMsg("foo", float64(1)),
		},
		"EmptyRawMessage": {
			pairs:       []any{"foo", pointer.To(json.RawMessage{})},
			expectedErr: fmt.Errorf("Invalid object: []"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			actualDoc, err := prepareOpMsg(tc.pairs...)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expectedOpMsg, actualDoc)
		})
	}
}

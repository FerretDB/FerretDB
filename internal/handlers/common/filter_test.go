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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

func Test_filterFieldExprBitsAllClearBinaryMask(t *testing.T) {
	tests := map[string]struct {
		fieldValue any
		maskValue  any
		want       bool
	}{
		"binary": {
			fieldValue: types.Binary{Subtype: 0x80, B: []byte{42, 0, 13}},
			maskValue:  types.Binary{B: []byte{2}},
			want:       false,
		},
		"binary-empty": {
			fieldValue: types.Binary{B: []byte{}},
			maskValue:  types.Binary{B: []byte{2}},
			want:       true,
		},
		"binary-big": {
			fieldValue: types.Binary{B: []byte{0, 0, 128}},
			maskValue:  types.Binary{B: []byte{2}},
			want:       true,
		},
		"binary-user-1": {
			fieldValue: types.Binary{Subtype: 0x80, B: []byte{0, 0, 30}},
			maskValue:  types.Binary{B: []byte{2}},
			want:       true,
		},
		"binary-user-2": {
			fieldValue: types.Binary{Subtype: 0x80, B: []byte{15, 0, 0, 0}},
			maskValue:  types.Binary{B: []byte{2}},
			want:       false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := filterFieldExprBitsAllClear(tt.fieldValue, tt.maskValue)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

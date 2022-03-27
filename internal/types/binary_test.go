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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryFromArray(t *testing.T) {
	t.Parallel()

	type args struct {
		values *Array
	}
	testCases := []struct {
		name    string
		args    args
		want    Binary
		wantErr error
	}{
		{
			name: "all bits",
			args: args{values: MustNewArray(
				int32(1), int32(2), int32(3), int32(4),
				int32(5), int32(6), int32(7),
			)},
			want: Binary{
				Subtype: BinaryGeneric,
				B:       []byte{0b11111110, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
		},
		{
			name: "bits 1,3,5",
			args: args{values: MustNewArray(int32(1), int32(3), int32(5))},
			want: Binary{
				Subtype: BinaryGeneric,
				B:       []byte{0b101010, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := BinaryFromArray(tc.args.values)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
			}
			assert.Equalf(t, tc.want, got, "BinaryFromArray(%v)", tc.args.values)
		})
	}
}

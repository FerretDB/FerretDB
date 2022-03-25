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
	type args struct {
		values *Array
	}
	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BinaryFromArray(tt.args.values)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			assert.Equalf(t, tt.want, got, "BinaryFromArray(%v)", tt.args.values)
		})
	}
}

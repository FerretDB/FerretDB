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

package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestNewSCRAMSHA256(t *testing.T) {
	tests := map[string]struct { //nolint:vet // used for testing only
		username string
		password string

		want *types.Document
	}{
		"empty": {}, // Should still work.
		"success": {
			username: "username",
			password: "password",
		},
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := NewSCRAMSHA256(tc.username, tc.password)
			require.NoError(t, err)
			assert.Equal(t, must.NotFail(got.Get("iterationCount")), int32(15000))
			assert.Len(t, must.NotFail(got.Get("salt")).(string), 60)
			assert.Len(t, must.NotFail(got.Get("serverKey")).(string), 44)
			assert.Len(t, must.NotFail(got.Get("storedKey")).(string), 44)
		})
	}
}

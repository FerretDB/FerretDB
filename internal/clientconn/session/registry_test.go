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

package session

import (
	"fmt"
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/password"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCleanExpired(t *testing.T) {
	t.Parallel()

	hash := password.UID("user1", "password")
	fmt.Println(hash)

	l := testutil.Logger(t)
	r := NewRegistry(l)

	r.NewSession("user1")
	r.NewSession("user1")

	id := uuid.New().String()
	require.NotNil(t, r.GetSession("user2", id))
	r.m["user2"][id].lastUsed = r.m["user2"][id].lastUsed.Add(-timeout)

	require.Nil(t, r.GetSession("user2", "invalid"))

	assert.Equal(t, 2, len(r.m["user1"]))
	assert.Equal(t, 1, len(r.m["user2"]))

	r.CleanExpired()
	assert.Equal(t, 2, len(r.m["user1"]))
	assert.Equal(t, 0, len(r.m["user2"]))

	var sessionID string
	for sessionID = range r.m["user1"] {
		break
	}
	r.EndSession("user1", sessionID)

	r.CleanExpired()
	assert.Equal(t, 1, len(r.m["user1"]))
	_, ok := r.m["user2"]
	assert.False(t, ok)
}

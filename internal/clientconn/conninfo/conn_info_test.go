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

package conninfo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnInfo(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		peerAddr string
	}{
		"EmptyPeerAddr": {
			peerAddr: "",
		},
		"NonEmptyPeerAddr": {
			peerAddr: "127.0.0.8:1234",
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			connInfo := &ConnInfo{
				PeerAddr: tc.peerAddr,
			}
			ctx = WithConnInfo(ctx, connInfo)
			actual := Get(ctx)
			assert.Equal(t, connInfo, actual)
		})
	}

	// special cases: if context is not set or something wrong is set in context, it panics.
	for name, tc := range map[string]struct {
		ctx context.Context
	}{
		"EmptyContext": {
			ctx: context.Background(),
		},
		"WrongValueType": {
			ctx: context.WithValue(context.Background(), connInfoKey, "wrong value type"),
		},
		"NilValue": {
			ctx: context.WithValue(context.Background(), connInfoKey, nil),
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Panics(t, func() {
				Get(tc.ctx)
			})
		})
	}
}

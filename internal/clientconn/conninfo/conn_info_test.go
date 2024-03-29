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
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		peer  netip.AddrPort
		local bool
	}{
		"Unix": {
			local: true,
		},
		"Local": {
			peer:  netip.MustParseAddrPort("127.42.7.1:1234"),
			local: true,
		},
		"NonLocal": {
			peer:  netip.MustParseAddrPort("192.168.0.1:1234"),
			local: false,
		},
		"LocalIPv6": {
			peer:  netip.MustParseAddrPort("[::1]:1234"),
			local: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			connInfo := &ConnInfo{
				Peer: tc.peer,
			}
			ctx = Ctx(ctx, connInfo)
			actual := Get(ctx)
			assert.Equal(t, connInfo, actual)
			assert.Equal(t, tc.local, actual.LocalPeer())
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

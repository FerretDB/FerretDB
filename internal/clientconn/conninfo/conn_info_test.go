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

func TestGet(t *testing.T) {
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

func TestSetClientMetadata(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		clientMetadata         any
		expectedClientMetadata ClientMetadata
		errMsg                 string
	}{
		"Full client metadata": {
			clientMetadata: `{
				"$k": [
				  "application",
				  "driver",
				  "platform",
				  "os"
				],
				"application": {
				  "$k": [
					"name"
				  ],
				  "name": "mongosh 1.10.5"
				},
				"driver": {
				  "$k": [
					"name",
					"version"
				  ],
				  "name": "nodejs|mongosh",
				  "version": "5.7.0|1.10.5"
				},
				"platform": "Node.js v16.20.2, LE",
				"os": {
				  "$k": [
					"name",
					"architecture",
					"version",
					"type"
				  ],
				  "name": "linux",
				  "architecture": "arm64",
				  "version": "5.15.49-linuxkit-pr",
				  "type": "Linux"
				}
			  }`,
			expectedClientMetadata: ClientMetadata{
				K: []string{"application", "driver", "platform", "os"},
				Application: Application{
					K:    []string{"name"},
					Name: "mongosh 1.10.5",
				},
				Driver: Driver{
					K:       []string{"name", "version"},
					Name:    "nodejs|mongosh",
					Version: "5.7.0|1.10.5",
				},
				Platform: "Node.js v16.20.2, LE",
				Os: Os{
					K:            []string{"name", "architecture", "version", "type"},
					Name:         "linux",
					Architecture: "arm64",
					Version:      "5.15.49-linuxkit-pr",
					Type:         "Linux",
				},
			},
		},
		"Partial client metadata": {
			clientMetadata: `{
				"$k": [
				  "application"
				],
				"application": {
				  "$k": [
					"name"
				  ],
				  "name": "mongosh 1.10.5"
				}
			  }`,
			expectedClientMetadata: ClientMetadata{
				K: []string{"application"},
				Application: Application{
					K:    []string{"name"},
					Name: "mongosh 1.10.5",
				},
			},
		},
		"Typo client metadata": {
			clientMetadata: `{
				"SK": [
				  "application"
				],
				"application": {
				  "$k": [
					"name"
				  ],
				  "name": "mongosh 1.10.5"
				}
			  }`,
			expectedClientMetadata: ClientMetadata{
				Application: Application{
					K:    []string{"name"},
					Name: "mongosh 1.10.5",
				},
			},
		},
		"Empty client metadata": {
			clientMetadata:         nil,
			expectedClientMetadata: ClientMetadata{},
			errMsg:                 "failed converting the client's metadata",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			connInfo := NewConnInfo()
			err := connInfo.SetClientMetadata(tc.clientMetadata)

			if tc.errMsg != "" {
				assert.EqualValues(t, tc.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tc.expectedClientMetadata, connInfo.clientMetadata)
		})
	}
}

func TestIsClientMetadataSet(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		Platform string
		isSet    bool
	}{
		"With client metadata": {
			Platform: "Linux",
			isSet:    true,
		},
		"No client metadata": {
			isSet: false,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			connInfo := NewConnInfo()
			connInfo.clientMetadata.Platform = tc.Platform

			isSet := connInfo.IsClientMetadataSet()
			assert.EqualValues(t, tc.isSet, isSet)
		})
	}
}

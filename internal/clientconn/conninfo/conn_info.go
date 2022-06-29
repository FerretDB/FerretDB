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
	"net"
)

// contextKey is a special type to represent context.WithValue keys a bit more safely.
type contextKey string

// connInfoKey stores the key for withConnInfo context value.
const connInfoKey = contextKey("connInfo")

// ConnInfo represents connection info.
type ConnInfo struct {
	PeerAddr net.Addr
}

// WithConnInfo returns a new context with the given ConnInfo.
func WithConnInfo(ctx context.Context, connInfo *ConnInfo) context.Context {
	return context.WithValue(ctx, connInfoKey, connInfo)
}

// GetConnInfo returns the ConnInfo value stored in ctx, or empty connInfo if there is nothing stored there.
func GetConnInfo(ctx context.Context) *ConnInfo {
	value := ctx.Value(connInfoKey)
	if value == nil {
		return &ConnInfo{}
	}

	connInfo, ok := value.(*ConnInfo)
	if !ok {
		panic("connInfo stored in context with wrong value type")
	}
	return connInfo
}

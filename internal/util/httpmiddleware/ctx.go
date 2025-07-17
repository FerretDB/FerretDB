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

package httpmiddleware

import (
	"context"

	"github.com/FerretDB/wire/wirebson"
)

// contextKey is a named unexported type for the safe use of [context.WithValue].
type contextKey struct{}

// lsidKey is context key for setting and getting `lsid`.
var lsidKey = contextKey{}

// CtxWithLSID returns a derived context which sets `lsid`.
func CtxWithLSID(ctx context.Context, lsid *wirebson.Document) context.Context {
	return context.WithValue(ctx, lsidKey, lsid)
}

// GetLSID returns the `lsid` value stored in ctx.
func GetLSID(ctx context.Context) *wirebson.Document {
	value := ctx.Value(lsidKey)
	if value == nil {
		panic("lsid is not set in context")
	}

	lsid, ok := value.(*wirebson.Document)
	if !ok {
		panic("lsid is set in context with wrong value type")
	}

	if lsid == nil {
		panic("lsid is set in context with nil value")
	}

	return lsid
}

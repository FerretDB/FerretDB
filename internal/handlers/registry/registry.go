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

package registry

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
)

// NewHandler represents a function that constructs a new handler.
type NewHandler func(opts NewHandlerOpts) (handlers.Interface, error)

// NewHandlerOpts represents configuration for constructing handlers.
type NewHandlerOpts struct {
	PostgresURL string
	TigrisURL   string
	Ctx         context.Context
	Logger      *zap.Logger
}

// Handlers maps handler names to constructors.
// The values for `Handlers` must be set through the `init()` functions of the corresponding handlers
// so that we can control which handlers will be included in the build with build tags.
var Handlers = map[string]NewHandler{}

func New(handler string, opts NewHandlerOpts) handlers.Interface {
	newHandler := Handlers[handler]
	if newHandler == nil {
		panic(fmt.Sprintf("Unknown backend handler %q.", handler))
	}
	h, err := newHandler(opts)
	if err != nil {
		panic(err.Error())
	}
	return h
}

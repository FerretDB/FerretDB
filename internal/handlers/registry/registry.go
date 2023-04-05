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

// Package registry provides a registry of handlers.
package registry

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/dummy"
	"github.com/FerretDB/FerretDB/internal/handlers/pg"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// newHandlerFunc represents a function that constructs a new handler.
type newHandlerFunc func(opts *NewHandlerOpts) (handlers.Interface, error)

// registry maps handler names to constructors.
//
// Map values must be added through the `init()` functions in separate files
// so that we can control which handlers will be included in the build with build tags.
var registry = map[string]newHandlerFunc{}

// NewHandlerOpts represents configuration for constructing handlers.
type NewHandlerOpts struct {
	// for all handlers
	Logger        *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider

	// for `pg` handler
	PostgreSQLURL string

	// for `tigris` handler
	TigrisURL          string
	TigrisClientID     string
	TigrisClientSecret string

	// for `hana` handler
	HANAURL string

	TestOpts
}

// TestOpts represents experimental configuration options.
type TestOpts struct {
	DisablePushdown bool
	EnableCursors   bool
}

// NewHandler constructs a new handler.
func NewHandler(name string, opts *NewHandlerOpts) (handlers.Interface, error) {
	if opts == nil {
		return nil, fmt.Errorf("opts is nil")
	}

	newHandler := registry[name]
	if newHandler == nil {
		return nil, fmt.Errorf("unknown handler %q", name)
	}

	return newHandler(opts)
}

// Handlers returns a list of all handlers registered at compile-time.
func Handlers() []string {
	handlers := maps.Keys(registry)
	slices.Sort(handlers)
	return handlers
}

// init registers handlers that are always enabled.
func init() {
	registry["dummy"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		return dummy.New(opts.Logger)
	}

	registry["pg"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		handlerOpts := &pg.NewOpts{
			PostgreSQLURL: opts.PostgreSQLURL,

			L:             opts.Logger,
			Metrics:       opts.Metrics,
			StateProvider: opts.StateProvider,

			DisablePushdown: opts.DisablePushdown,
			EnableCursors:   opts.EnableCursors,
		}
		return pg.New(handlerOpts)
	}
}

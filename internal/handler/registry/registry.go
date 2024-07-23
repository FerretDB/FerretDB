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
	"log/slog"
	"time"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handler"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// newHandlerFunc represents a function that constructs a new handler.
type newHandlerFunc func(opts *NewHandlerOpts) (*handler.Handler, CloseBackendFunc, error)

// CloseBackendFunc represents a function that closes a backend.
type CloseBackendFunc func()

// registry maps handler names to constructors.
//
// Map values must be added through the `init()` functions in separate files
// so that we can control which handlers will be included in the build with build tags.
var registry = map[string]newHandlerFunc{}

// NewHandlerOpts represents configuration for constructing handlers.
type NewHandlerOpts struct {
	// for all backends
	Logger        *slog.Logger
	ConnMetrics   *connmetrics.ConnMetrics
	StateProvider *state.Provider
	TCPHost       string
	ReplSetName   string
	SetupDatabase string
	SetupUsername string
	SetupPassword password.Password
	SetupTimeout  time.Duration

	// for `postgresql` handler
	PostgreSQLURL string

	// for `sqlite` handler
	SQLiteURL string

	// for `hana` handler
	HANAURL string

	// for `mysql` handler
	MySQLURL string

	TestOpts

	_ struct{} // prevent unkeyed literals
}

// TestOpts represents experimental configuration options.
type TestOpts struct {
	DisablePushdown         bool
	EnableNestedPushdown    bool
	CappedCleanupInterval   time.Duration
	CappedCleanupPercentage uint8
	EnableNewAuth           bool
	BatchSize               int
	MaxBsonObjectSizeBytes  int
	_                       struct{} // prevent unkeyed literals
}

// NewHandler constructs a new handler.
//
// The caller is responsible to call CloseBackendFunc when the handler is no longer needed.
func NewHandler(name string, opts *NewHandlerOpts) (*handler.Handler, CloseBackendFunc, error) {
	if opts == nil {
		return nil, nil, fmt.Errorf("opts is nil")
	}

	// handle deprecated variant
	if name == "pg" {
		name = "postgresql"
	}

	newHandler := registry[name]
	if newHandler == nil {
		return nil, nil, fmt.Errorf("unknown handler %q", name)
	}

	return newHandler(opts)
}

// Handlers returns a list of all handlers registered at compile-time.
func Handlers() []string {
	res := make([]string, 0, len(registry))

	// double check registered names and return them in the right order
	for _, h := range []string{"postgresql", "sqlite", "hana", "mysql"} {
		if _, ok := registry[h]; !ok {
			continue
		}

		res = append(res, h)
	}

	if len(res) != len(registry) {
		panic("registry is not in sync")
	}

	return res
}

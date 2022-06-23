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

//go:build tigris
// +build tigris

package main

import (
	"flag"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris"
)

// `tigris` handler flags.
var (
	tigrisURLF = flag.String("tigris-url", "127.0.0.1:8081", "Tigris URL")
)

// init registers `tigris` handler for Tigris that is enabled only when compiled with `tigris` build tag.
func init() {
	registeredHandlers["tigris"] = func(opts *newHandlerOpts) (handlers.Interface, error) {
		handlerOpts := &tigris.NewOpts{
			TigrisURL: *tigrisURLF,
			L:         opts.logger,
		}
		return tigris.New(handlerOpts)
	}
}

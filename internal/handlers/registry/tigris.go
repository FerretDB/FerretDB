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

//go:build ferretdb_tigris

package registry

import (
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris"
)

// init registers "tigris" handler for Tigris when "ferretdb_tigris" build tag is provided.
func init() {
	registry["tigris"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		handlerOpts := &tigris.NewOpts{
			ClientID:     opts.TigrisClientID,
			ClientSecret: opts.TigrisClientSecret,
			Token:        opts.TigrisToken,
			URL:          opts.TigrisURL,

			L:               opts.Logger,
			Metrics:         opts.Metrics,
			StateProvider:   opts.StateProvider,
			DisablePushdown: opts.DisablePushdown,
		}
		return tigris.New(handlerOpts)
	}
}

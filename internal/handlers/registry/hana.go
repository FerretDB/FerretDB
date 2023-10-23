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

//go:build ferretdb_hana

package registry

import (
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/hana"
	"github.com/FerretDB/FerretDB/internal/handlers/sqlite"
)

// init registers "hana" handler for Hana when "ferretdb_hana" build tag is provided.
func init() {
	registry["hana"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		opts.Logger.Warn("HANA handler is in alpha. It is not supported yet.")

		if opts.UseNewHana {
			handlerOpts := &sqlite.NewOpts{
				Backend: "hana",
				URI:     opts.HANAURL,

				L:             opts.Logger.Named("hana"),
				ConnMetrics:   opts.ConnMetrics,
				StateProvider: opts.StateProvider,

				DisableFilterPushdown: opts.DisableFilterPushdown,
				EnableSortPushdown:    opts.EnableSortPushdown,
				EnableOplog:           opts.EnableOplog,
			}

			return sqlite.New(handlerOpts)
		}

		handlerOpts := &hana.NewOpts{
			HANAURL:       opts.HANAURL,
			L:             opts.Logger,
			ConnMetrics:   opts.ConnMetrics,
			StateProvider: opts.StateProvider,
		}

		return hana.New(handlerOpts)
	}
}

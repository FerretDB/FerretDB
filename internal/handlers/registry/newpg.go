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

//go:build ferretdb_newpg

package registry

import (
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/sqlite"
)

// init registers new "pg" handler.
func init() {
	registry["pg"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		opts.Logger.Warn("New PostgreSQL backend is in alpha.")

		handlerOpts := &sqlite.NewOpts{
			Backend: "postgresql",
			URI:     opts.PostgreSQLURL,

			L:             opts.Logger.Named("postgresql"),
			ConnMetrics:   opts.ConnMetrics,
			StateProvider: opts.StateProvider,

			DisableFilterPushdown: opts.DisableFilterPushdown,
			EnableSortPushdown:    opts.EnableSortPushdown,
			EnableOplog:           opts.EnableOplog,
		}

		return sqlite.New(handlerOpts)
	}
}

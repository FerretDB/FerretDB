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

// This file contains the implementation of `task env-data` command
// that creates collections in the `test` database for manual experiments.
// It is not a real test file but wrapped into one
// because all needed functionality is already available in setup helpers.
// This file is skipped by default; we pass the build tag only in the `task env-data` command.

// go:build ferretdb_testenvdata

package integration

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestEnvData(t *testing.T) {
	t.Parallel()

	// Setups one collection for each data set for all handlers and MongoDB.
	t.Run("All", func(t *testing.T) {
		t.Parallel()

		setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
			DatabaseName: "test",
			Providers:    shareddata.AllProviders(),
		})
	})

	// Setups old `values` collection with mixed types for `pg` handler and MongoDB.
	t.Run("Values", func(t *testing.T) {
		setup.SkipForTigris(t)

		t.Parallel()

		setup.SetupWithOpts(t, &setup.SetupOpts{
			DatabaseName:   "test",
			CollectionName: "values",
			Providers:      []shareddata.Provider{shareddata.Scalars, shareddata.Composites},
		})
	})
}

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

package integration

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestEnvData(t *testing.T) {
	t.Parallel()

	// see `env-data` Taskfile target

	t.Run("All", func(t *testing.T) {
		t.Parallel()

		setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
			DatabaseName: "test",
			Providers:    shareddata.AllProviders(),
		})
	})

	// setup old `values` collection with mixed types for PostgreSQL and MongoDB
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

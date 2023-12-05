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
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

// tagPackages returns packages that are imported by FerretDB only when the given Go build tag is provided.
func tagPackages(t *testing.T, tag string) []string {
	t.Helper()

	type list struct {
		Deps []string `json:"Deps"`
	}

	var withTag list
	b, err := exec.Command("go", "list", "-json", "-tags", tag, "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &withTag))

	var withoutTag list
	b, err = exec.Command("go", "list", "-json", "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &withoutTag))

	packages := make(map[string]struct{}, len(withTag.Deps))

	for _, p := range withTag.Deps {
		packages[p] = struct{}{}
	}

	for _, p := range withoutTag.Deps {
		delete(packages, p)
	}

	return maps.Keys(packages)
}

// negTagPackages returns packages that are NOT imported by FerretDB when the
// given Go build negative tag is provided. Eg of negative tags
// 'ferretdb_no_postgresql', 'ferretdb_no_sqlite'.
func negTagPackages(t *testing.T, negTag string) []string {
	t.Helper()

	type list struct {
		Deps []string `json:"Deps"`
	}

	// Few packages will not be imported if a negative tag is given
	var withNegTag list
	b, err := exec.Command("go", "list", "-json", "-tags", negTag, "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &withNegTag))

	// Packages which are removed using negative tag will be prensent in this
	// list
	var withoutNegTag list
	b, err = exec.Command("go", "list", "-json", "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &withoutNegTag))

	packages := make(map[string]struct{}, len(withoutNegTag.Deps))

	for _, p := range withoutNegTag.Deps {
		packages[p] = struct{}{}
	}

	for _, p := range withNegTag.Deps {
		delete(packages, p)
	}

	return maps.Keys(packages)
}

// TestDeps ensures that some packages are imported
// only when the corresponding backend handler is enabled via Go build tag.
func TestDeps(t *testing.T) {
	t.Parallel()

	t.Run("Hana", func(t *testing.T) {
		t.Parallel()

		diff := tagPackages(t, "ferretdb_hana")
		assert.Contains(t, diff, "github.com/SAP/go-hdb/driver")
	})

	t.Run("NoPostgres", func(t *testing.T) {
		t.Parallel()

		diff := negTagPackages(t, "ferretdb_no_postgresql")
		assert.Contains(t, diff, "github.com/FerretDB/FerretDB/internal/backends/postgresql")
		assert.Contains(t, diff, "github.com/jackc/pgx/v5")
	})

	t.Run("NoSqlite", func(t *testing.T) {
		t.Parallel()

		diff := negTagPackages(t, "ferretdb_no_sqlite")
		assert.Contains(t, diff, "github.com/FerretDB/FerretDB/internal/backends/sqlite")
		assert.Contains(t, diff, "modernc.org/sqlite")
	})
}

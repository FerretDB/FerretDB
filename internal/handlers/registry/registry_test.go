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

// extraDeps returns extra FerretDB dependencies (packages) when the given tag is enabled.
func extraDeps(t *testing.T, tag string) []string {
	t.Helper()

	type goListResult struct {
		Deps []string `json:"Deps"`
	}

	b, err := exec.Command("go", "list", "-json", "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	var withoutTag goListResult
	require.NoError(t, json.Unmarshal(b, &withoutTag))

	b, err = exec.Command("go", "list", "-json", "-tags", tag, "../../../cmd/ferretdb").Output()
	require.NoError(t, err)
	var withTag goListResult
	require.NoError(t, json.Unmarshal(b, &withTag))

	diff := make(map[string]struct{}, len(withTag.Deps))
	for _, dep := range withTag.Deps {
		diff[dep] = struct{}{}
	}

	for _, dep := range withoutTag.Deps {
		delete(diff, dep)
	}

	return maps.Keys(diff)
}

// TestDeps ensures that extra dependencies are not used when the given Go build tag / backend is not enabled.
func TestDeps(t *testing.T) {
	t.Parallel()

	t.Run("Tigris", func(t *testing.T) {
		t.Parallel()

		diff := extraDeps(t, "ferretdb_tigris")
		assert.Contains(t, diff, "github.com/tigrisdata/tigris-client-go/driver")
		assert.Contains(t, diff, "google.golang.org/grpc")
	})

	t.Run("Hana", func(t *testing.T) {
		t.Parallel()

		diff := extraDeps(t, "ferretdb_hana")
		assert.Contains(t, diff, "github.com/SAP/go-hdb/driver")
	})
}

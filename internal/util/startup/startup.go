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

// Package startup provides initialization code shared by main and ferretdb packages.
package startup

import (
	"fmt"
	"path/filepath"

	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// State setups state provider for the given directory.
func State(dir string) (*state.Provider, error) {
	if dir == "" {
		return nil, fmt.Errorf("state directory is not set")
	}

	f, err := filepath.Abs(filepath.Join(dir, "state.json"))
	if err != nil {
		return nil, err
	}

	sp, err := state.NewProvider(f)
	if err != nil {
		return nil, stateFileErr(f, err)
	}

	return sp, nil
}

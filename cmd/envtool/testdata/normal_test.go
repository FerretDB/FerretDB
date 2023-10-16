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

package testdata

import "testing"

func TestNormal1(t *testing.T) {
	t.Run("Parallel", func(t *testing.T) {
		t.Parallel()
	})

	t.Run("NotParallel", func(t *testing.T) {
	})
}

func TestNormal2(t *testing.T) {
	t.Parallel()

	t.Run("NotParallel", func(t *testing.T) {
	})

	t.Run("Parallel", func(t *testing.T) {
		t.Parallel()
	})
}

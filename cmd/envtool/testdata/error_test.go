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

func TestError1(t *testing.T) {
	t.Log("not hidden 1")

	t.Error("Error 1")

	t.Log("not hidden 2")
}

func TestError2(t *testing.T) {
	t.Log("not hidden 3")

	t.Parallel()

	t.Log("not hidden 4")

	t.Run("Parallel", func(t *testing.T) {
		t.Log("not hidden 5")

		t.Parallel()

		t.Log("not hidden 6")

		t.Error("Error 2")

		t.Log("not hidden 7")
	})

	t.Run("NotParallel", func(t *testing.T) {
		t.Log("hidden")
	})
}

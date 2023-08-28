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

package testfail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormal(t *testing.T) {
	// comment out to test manually
	t.Skip("test fails (as expected)")

	t.Parallel()

	t.Run("Fatal", func(t *testing.T) {
		t.Fatal("Fatal")

		panic("not reached")
	})

	assert.True(t, t.Failed())

	t.Run("ErrorAndSkip", func(t *testing.T) {
		t.Error("Error")
		t.SkipNow()

		panic("not reached")
	})

	assert.True(t, t.Failed())

	t.Skip("skipping failing test does not mark it as not failed")
}

func TestExpected(t *testing.T) {
	t.Parallel()

	t.Run("Fatal", func(tt *testing.T) {
		t := Expected(tt, "expected failure")
		t.Fatal("Fatal")

		panic("not reached")
	})

	assert.False(t, t.Failed())

	t.Run("ErrorAndSkip", func(tt *testing.T) {
		t := Expected(tt, "expected failure")
		t.Error("Error")
		t.SkipNow()

		panic("not reached")
	})

	assert.False(t, t.Failed())
}

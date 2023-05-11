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

package contract

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

var testError = fmt.Errorf("expected error value")

func TestEnsureError(t *testing.T) {
	t.Parallel()

	require.True(t, debugbuild.Enabled)

	assert.NotPanics(t, func() {
		EnsureError(nil)
	})

	t.Run("Value", func(t *testing.T) {
		t.Parallel()

		t.Run("OK", func(t *testing.T) {
			t.Parallel()

			assert.NotPanics(t, func() {
				var err error
				EnsureError(err, testError)
			})

			assert.NotPanics(t, func() {
				err := testError
				EnsureError(err, testError)
			})

			assert.NotPanics(t, func() {
				err := fmt.Errorf("wrapped: %w", testError)
				EnsureError(err, testError)
			})
		})

		t.Run("Fail", func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithValue(t, `EnsureError: &errors.errorString{s:"some other error"}`, func() {
				err := fmt.Errorf("some other error")
				EnsureError(err, testError)
			})
		})

		t.Run("Invalid", func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithValue(t, `EnsureError: invalid actual value (*fs.PathError)(nil)`, func() {
				var err *fs.PathError
				EnsureError(err, testError)
			})

			assert.PanicsWithValue(t, `EnsureError: invalid expected value <nil>`, func() {
				err := testError
				EnsureError(err, nil)
			})

			assert.PanicsWithValue(t, `EnsureError: invalid expected value (*fs.PathError)(nil)`, func() {
				err := testError
				EnsureError(err, (*fs.PathError)(nil))
			})
		})
	})
}

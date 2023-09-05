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

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // we modify the global timestampCounter
func TestNextTimestamp(t *testing.T) {
	t.Run("UnixZero", func(t *testing.T) {
		d := time.Unix(0, 0).UTC()

		timestampCounter.Store(0)
		assert.Equal(t, Timestamp(1), NextTimestamp(d))
		assert.Equal(t, Timestamp(2), NextTimestamp(d))

		assert.Equal(t, d, NextTimestamp(d).Time())
	})

	t.Run("Normal", func(t *testing.T) {
		d := time.Date(2023, time.September, 12, 59, 44, 42, 0, time.UTC)

		timestampCounter.Store(0)
		assert.Equal(t, Timestamp(7278646209986691073), NextTimestamp(d))
		assert.Equal(t, Timestamp(7278646209986691074), NextTimestamp(d))

		assert.Equal(t, d, NextTimestamp(d).Time())
	})
}

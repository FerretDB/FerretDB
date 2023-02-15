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

//nolint:paralleltest // we modify the global objectIDProcess
func TestNewObjectID(t *testing.T) {
	objectIDProcess = [5]byte{0x0b, 0xad, 0xc0, 0xff, 0xee}
	ts := time.Date(2022, time.April, 13, 12, 44, 42, 0, time.UTC)

	objectIDCounter.Store(0)
	assert.Equal(
		t,
		ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x01},
		newObjectIDTime(ts),
	)
	assert.Equal(
		t,
		ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x02},
		newObjectIDTime(ts),
	)

	// test wraparound
	objectIDCounter.Store(1<<24 - 2)
	assert.Equal(
		t,
		ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0xff, 0xff, 0xff},
		newObjectIDTime(ts),
	)
	assert.Equal(
		t,
		ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x00},
		newObjectIDTime(ts),
	)
	assert.Equal(
		t,
		ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x01},
		newObjectIDTime(ts),
	)
}

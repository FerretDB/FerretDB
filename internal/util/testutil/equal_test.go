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

package testutil

import (
	"math"
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestEqual(t *testing.T) {
	t.Parallel()

	AssertEqual(
		t,
		must.NotFail(types.NewDocument("foo", "bar", "baz", int32(42))),
		must.NotFail(types.NewDocument("foo", "bar", "baz", int32(42))),
	)
	AssertNotEqual(
		t,
		must.NotFail(types.NewDocument("foo", "bar", "baz", int32(42))),
		must.NotFail(types.NewDocument("baz", int32(42), "foo", "bar")),
	)

	AssertEqual(t, math.Inf(+1), math.Inf(+1))
	AssertNotEqual(t, math.Inf(-1), math.Inf(+1))

	AssertEqual(t, 0.0, math.Copysign(0, +1))
	AssertNotEqual(t, math.Copysign(0, +1), math.Copysign(0, -1))

	AssertEqual(
		t,
		time.Date(2022, time.March, 11, 8, 8, 42, 123456789, time.FixedZone("Test", int(3*time.Hour.Seconds()))),
		time.Date(2022, time.March, 11, 5, 8, 42, 123456789, time.UTC),
	)
	AssertNotEqual(
		t,
		time.Date(2022, time.March, 11, 8, 8, 42, 123456789, time.FixedZone("Test", int(3*time.Hour.Seconds()))),
		time.Date(2022, time.March, 11, 8, 8, 42, 123456789, time.UTC),
	)
}

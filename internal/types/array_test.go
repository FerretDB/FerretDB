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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

var scalarTypesArray = must.NotFail(NewArray(
	int32(42),
	int64(42),
	float64(42),
	"foo",
	Binary{},
	ObjectID{},
	true,
	time.Time{},
	NullType{},
	Regex{},
	Timestamp(42),
))

func TestArray(t *testing.T) {
	t.Parallel()

	t.Run("MethodsOnNil", func(t *testing.T) {
		t.Parallel()

		var a *Array
		assert.Zero(t, a.Len())
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			NewArray(42)
		})
	})

	t.Run("ZeroValues", func(t *testing.T) {
		t.Parallel()

		// to avoid {} != nil in tests
		assert.Nil(t, must.NotFail(NewArray()).s)
		assert.Nil(t, MakeArray(0).s)

		var a Array
		assert.Equal(t, 0, a.Len())
		assert.Nil(t, a.s)

		a.Append(Null)
		value, err := a.Get(0)
		assert.NoError(t, err)
		assert.Equal(t, Null, value)
	})

	t.Run("DeepCopy", func(t *testing.T) {
		t.Parallel()

		a := must.NotFail(NewArray(int32(42)))
		b := a.DeepCopy()
		assert.Equal(t, a, b)
		assert.NotSame(t, a, b)

		a.s[0] = "foo"
		assert.NotEqual(t, a, b)
		assert.Equal(t, int32(42), b.s[0])
	})
}

func TestArrayMinMax(t *testing.T) {
	int32Array := must.NotFail(NewArray(int32(42), int32(50), int32(2), int32(30)))
	int64Array := must.NotFail(NewArray(int64(42), int64(50), int64(2), int64(30)))
	intArray := must.NotFail(NewArray(int64(42), int64(50), int32(50), int64(2), int32(2), int64(30)))
	floatArray := must.NotFail(NewArray(42.0, 50.1, 2.2, 30.3))
	numericArray := must.NotFail(NewArray(int64(2), int32(2), 2.0))
	zeroArray := must.NotFail(NewArray(int64(0), int32(0), math.Copysign(0, +1), math.Copysign(0, -1)))
	stingArray := must.NotFail(NewArray("foo", "zoo", "", "bar"))

	const (
		min = "min"
		max = "max"
	)

	for name, tc := range map[string]struct {
		arr           *Array
		minOrMax      string
		expectedValue any
	}{
		"Int32Min": {
			arr:           int32Array,
			minOrMax:      min,
			expectedValue: int32(2),
		},
		"Int64Min": {
			arr:           int64Array,
			minOrMax:      min,
			expectedValue: int64(2),
		},
		"IntMin": {
			arr:           intArray,
			minOrMax:      min,
			expectedValue: int64(2),
		},
		"FloatMin": {
			arr:           floatArray,
			minOrMax:      min,
			expectedValue: 2.2,
		},
		"NumericMin": {
			arr:           numericArray,
			minOrMax:      min,
			expectedValue: int64(2),
		},
		"ZeroMin": {
			arr:           zeroArray,
			minOrMax:      min,
			expectedValue: int64(0),
		},
		"StringMin": {
			arr:           stingArray,
			minOrMax:      min,
			expectedValue: "",
		},
		"AllTypeMin": {
			arr:           scalarTypesArray,
			minOrMax:      min,
			expectedValue: NullType{},
		},
		"Int32Max": {
			arr:           int32Array,
			minOrMax:      max,
			expectedValue: int32(50),
		},
		"Int64Max": {
			arr:           int64Array,
			minOrMax:      max,
			expectedValue: int64(50),
		},
		"IntMax": {
			arr:           intArray,
			minOrMax:      max,
			expectedValue: int64(50),
		},
		"FloatMax": {
			arr:           floatArray,
			minOrMax:      max,
			expectedValue: 50.1,
		},
		"NumericMax": {
			arr:           numericArray,
			minOrMax:      max,
			expectedValue: int64(2),
		},
		"ZeroMax": {
			arr:           zeroArray,
			minOrMax:      max,
			expectedValue: int64(0),
		},
		"StringMax": {
			arr:           stingArray,
			minOrMax:      max,
			expectedValue: "zoo",
		},
		"AllTypeMax": {
			arr:           scalarTypesArray,
			minOrMax:      max,
			expectedValue: Regex{},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			var res any
			if tc.minOrMax == min {
				res = tc.arr.Min()
			} else {
				res = tc.arr.Max()
			}

			assert.Equal(t, res, tc.expectedValue)
		})
	}
}

func TestFilterArrayByType(t *testing.T) {
	for name, tc := range map[string]struct {
		arr           *Array
		refValue      any
		expectedArray any
	}{
		"Int32": {
			arr:      scalarTypesArray,
			refValue: int32(42),
			expectedArray: must.NotFail(NewArray(
				int32(42),
				int64(42),
				float64(42),
			)),
		},
		"Int64": {
			arr:      scalarTypesArray,
			refValue: int64(42),
			expectedArray: must.NotFail(NewArray(
				int32(42),
				int64(42),
				float64(42),
			)),
		},
		"Float64": {
			arr:      scalarTypesArray,
			refValue: 240.2,
			expectedArray: must.NotFail(NewArray(
				int32(42),
				int64(42),
				float64(42),
			)),
		},
		"Zero": {
			arr:      scalarTypesArray,
			refValue: math.Copysign(0, +1),
			expectedArray: must.NotFail(NewArray(
				int32(42),
				int64(42),
				float64(42),
			)),
		},
		"String": {
			arr:      scalarTypesArray,
			refValue: "a",
			expectedArray: must.NotFail(NewArray(
				"foo",
			)),
		},
		"Binary": {
			arr:      scalarTypesArray,
			refValue: Binary{},
			expectedArray: must.NotFail(NewArray(
				Binary{},
			)),
		},
		"ObjectID": {
			arr:      scalarTypesArray,
			refValue: ObjectID{},
			expectedArray: must.NotFail(NewArray(
				ObjectID{},
			)),
		},
		"Boolean": {
			arr:      scalarTypesArray,
			refValue: true,
			expectedArray: must.NotFail(NewArray(
				true,
			)),
		},
		"Time": {
			arr:      scalarTypesArray,
			refValue: time.Time{},
			expectedArray: must.NotFail(NewArray(
				time.Time{},
			)),
		},
		"NullType": {
			arr:      scalarTypesArray,
			refValue: NullType{},
			expectedArray: must.NotFail(NewArray(
				NullType{},
			)),
		},
		"Regex": {
			arr:      scalarTypesArray,
			refValue: Regex{},
			expectedArray: must.NotFail(NewArray(
				Regex{},
			)),
		},
		"Timestamp": {
			arr:      scalarTypesArray,
			refValue: Timestamp(42),
			expectedArray: must.NotFail(NewArray(
				Timestamp(42),
			)),
		},
		"EmptyArray": {
			arr:           must.NotFail(NewArray()),
			refValue:      Timestamp(42),
			expectedArray: must.NotFail(NewArray()),
		},
		"Nil": {
			arr:           nil,
			refValue:      Timestamp(42),
			expectedArray: must.NotFail(NewArray()),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res := tc.arr.FilterArrayByType(tc.refValue)

			assert.Equal(t, tc.expectedArray, res)
		})
	}
}

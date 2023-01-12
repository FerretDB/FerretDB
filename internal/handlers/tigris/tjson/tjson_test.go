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

package tjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// testCase represents a typical test case for TJSON to be used with the testing functions.
type testCase struct {
	name   string    // test case name
	v      tjsontype // tjson value
	schema *Schema   // tjson schema
	j      string    // json data to unmarshal
	canonJ string    // canonical form without extra object fields, zero values, etc.
	jErr   string    // unwrapped
	sErr   string
}

func TestMarshalUnmarshal(t *testing.T) {
	expected, err := types.NewDocument(
		"_id", types.ObjectID{0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		"string", "foo",
		"int32", int32(42),
		"int64", int64(123),
		"binary", types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
	)
	require.NoError(t, err)

	actualSchema, err := DocumentSchema(expected)
	require.NoError(t, err)

	expectedSchema := &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$k":     {Type: Array, Items: stringSchema},
			"_id":    objectIDSchema,
			"string": stringSchema,
			"int32":  int32Schema,
			"int64":  int64Schema,
			"binary": binarySchema,
		},
		PrimaryKey: []string{"_id"},
	}
	assert.Equal(t, actualSchema, expectedSchema)

	actualB, err := Marshal(expected)
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"$k": ["_id", "string", "int32", "int64", "binary"],
		"_id": "AAECBAUGBwgJCgsM",
		"string": "foo",
		"int32": 42,
		"int64": 123,
		"binary": {"$b": "Qg==", "s": 128}
	}`))
	assert.Equal(t, string(expectedB), string(actualB))

	actual, err := Unmarshal(expectedB, expectedSchema)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)

	assert.Equal(
		t,
		types.ObjectID{0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		must.NotFail(expected.Get("_id")).(types.ObjectID),
	)
}

// assertEqual is assert.Equal that also can compare NaNs and ±0.
func assertEqual(tb testing.TB, expected, actual any, msgAndArgs ...any) bool {
	tb.Helper()

	switch expected := expected.(type) {
	// should not be possible, check just in case
	case doubleType, float64:
		tb.Fatalf("unexpected type %[1]T: %[1]v", expected)

	case *doubleType:
		require.IsType(tb, expected, actual, msgAndArgs...)
		e := float64(*expected)
		a := float64(*actual.(*doubleType))
		if math.IsNaN(e) || math.IsNaN(a) {
			return assert.Equal(tb, math.IsNaN(e), math.IsNaN(a), msgAndArgs...)
		}
		if e == 0 && a == 0 {
			return assert.Equal(tb, math.Signbit(e), math.Signbit(a), msgAndArgs...)
		}
		// fallthrough to regular assert.Equal below
	}

	return assert.Equal(tb, expected, actual, msgAndArgs...)
}

// lastErr returns the last error in error chain.
func lastErr(err error) error {
	for {
		e := errors.Unwrap(err)
		if e == nil {
			return err
		}
		err = e
	}
}

func testJSON(t *testing.T, testCases []testCase, newFunc func() tjsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.j, "j should not be empty")

			t.Parallel()

			if tc.jErr == "" {
				var dst bytes.Buffer
				require.NoError(t, json.Compact(&dst, []byte(tc.j)))
				require.Equal(t, tc.j, dst.String(), "j should be compacted")
				if tc.canonJ != "" {
					dst.Reset()
					require.NoError(t, json.Compact(&dst, []byte(tc.canonJ)))
					require.Equal(t, tc.canonJ, dst.String(), "canonJ should be compacted")
				}
			}

			t.Run("UnmarshalJSON", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				ok, err := unmarshalJSON(v, tc.j)

				// UnmarshalJSON is not supported for some types, nothing to assert.
				if !ok {
					return
				}

				if tc.jErr == "" {
					require.NoError(t, err)
					assertEqual(t, tc.v, v)
					return
				}

				require.Error(t, err)
				require.Equal(t, tc.jErr, lastErr(err).Error())
			})

			t.Run("UnmarshalWithSchema", func(t *testing.T) {
				t.Parallel()

				v, err := Unmarshal([]byte(tc.j), tc.schema)

				if tc.sErr != "" {
					require.Error(t, err)
					require.Equal(t, tc.sErr, lastErr(err).Error())
					return
				}

				if tc.jErr != "" {
					require.Error(t, err)
					require.Equal(t, tc.jErr, lastErr(err).Error())
					return
				}

				require.NoError(t, err)
				assertEqual(t, tc.v, toTJSON(v))
			})

			t.Run("MarshalJSON", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})

			t.Run("Marshal", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				actualJ, err := Marshal(fromTJSON(tc.v))
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})
		})
	}
}

func fuzzJSON(f *testing.F, testCases []testCase) {
	for _, tc := range testCases {
		schema, err := tc.schema.Marshal()
		require.NoError(f, err)

		f.Add(tc.j, string(schema))

		if tc.canonJ != "" {
			f.Add(tc.canonJ, string(schema))
		}
	}

	f.Fuzz(func(t *testing.T, j, schema string) {
		t.Parallel()

		// raw "null" should never reach UnmarshalJSON due to the way encoding/json works
		if j == "null" {
			t.Skip()
		}

		var s Schema

		err := s.Unmarshal([]byte(schema))
		if err != nil {
			t.Skip()
		}

		// j may not be a canonical form.
		// We can't compare it with MarshalJSON() result directly.
		// Instead, we compare with round-trip result.

		val, err := Unmarshal([]byte(j), &s)
		if err != nil {
			t.Skip()
		}
		v := toTJSON(val)

		// test MarshalJSON
		{
			b, err := v.MarshalJSON()
			require.NoError(t, err)
			j = string(b)
		}

		// test Unmarshal
		{
			actualV, err := Unmarshal([]byte(j), &s)
			require.NoError(t, err)
			assertEqual(t, v, toTJSON(actualV))
		}
	})
}

func benchmark(b *testing.B, testCases []testCase) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Unmarshal", func(b *testing.B) {
				data := []byte(tc.j)
				var v any
				var err error

				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v, err = Unmarshal([]byte(tc.j), tc.schema)
				}

				b.StopTimer()

				if tc.jErr != "" {
					require.Error(b, err)
					require.Equal(b, tc.jErr, lastErr(err).Error())

					return
				}

				if tc.sErr != "" {
					require.Error(b, err)
					require.Equal(b, tc.sErr, lastErr(err).Error())

					return
				}

				require.NoError(b, err)
				assertEqual(b, tc.v, toTJSON(v))
			})
		})
	}
}

// unmarshalJSON encapsulates type switch and calls UnmarshalJSON on the given value.
// It is called this way as tjsontype itself doesn't implement json.Unmarshaler interface.
// This function returns true if UnmarshalJSON is implemented and called and false if not.
func unmarshalJSON(v tjsontype, j string) (bool, error) {
	var err error
	switch v := v.(type) {
	case *documentType:
		// UnmarshalJSON is not supported for documents.
		return false, nil
	case *arrayType:
		// UnmarshalJSON is not supported for arrays.
		return false, nil
	case *doubleType:
		err = v.UnmarshalJSON([]byte(j))
	case *stringType:
		err = v.UnmarshalJSON([]byte(j))
	case *binaryType:
		err = v.UnmarshalJSON([]byte(j))
	case *objectIDType:
		err = v.UnmarshalJSON([]byte(j))
	case *boolType:
		err = v.UnmarshalJSON([]byte(j))
	case *dateTimeType:
		err = v.UnmarshalJSON([]byte(j))
	case *regexType:
		err = v.UnmarshalJSON([]byte(j))
	case *int32Type:
		err = v.UnmarshalJSON([]byte(j))
	case *timestampType:
		err = v.UnmarshalJSON([]byte(j))
	case *int64Type:
		err = v.UnmarshalJSON([]byte(j))
	default:
		panic(fmt.Sprintf("testing is not implemented for the type %T", v))
	}

	return true, err
}

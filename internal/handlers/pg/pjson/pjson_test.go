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

package pjson

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
)

type testCase struct {
	name   string
	v      pjsontype
	sch    *elem
	j      string
	canonJ string // canonical form without extra object fields, zero values, etc.
	jErr   string // unwrapped
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

func testJSON(t *testing.T, testCases []testCase, newFunc func() pjsontype) {
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
				err := unmarshalJSON(v, &tc)

				if tc.jErr == "" {
					require.NoError(t, err)
					assertEqual(t, tc.v, v)
					return
				}

				require.Error(t, err)
				require.Equal(t, tc.jErr, lastErr(err).Error())
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

				actualJ, err := MarshalSingleValue(fromPJSON(tc.v))
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

func fuzzJSON(f *testing.F, testCases []testCase, newFunc func() pjsontype) {
	for _, tc := range testCases {
		sch := must.NotFail(json.Marshal(tc.sch))
		f.Add(tc.j, string(sch))

		if tc.canonJ != "" {
			f.Add(tc.canonJ, string(sch))
		}
	}

	f.Fuzz(func(t *testing.T, j, jsch string) {
		t.Parallel()

		// raw "null" should never reach UnmarshalJSON due to the way encoding/json works
		if j == "null" {
			t.Skip()
		}

		// no reason to fuzz all the possible input for schemas,
		// so if sch is not a valid schema, skip the test
		var sch elem
		if err := json.Unmarshal([]byte(jsch), &sch); err != nil {
			t.Skip()
		}

		// j may not be a canonical form.
		// We can't compare it with MarshalJSON() result directly.
		// Instead, we compare with round-trip result.

		v := newFunc()
		tc := testCase{
			sch: &sch,
			j:   j,
		}
		if err := unmarshalJSON(v, &tc); err != nil {
			t.Skip()
		}

		// Temporary hack, should be removed once we improve our validation.
		// TODO https://github.com/FerretDB/FerretDB/issues/1273
		{
			d, ok := fromPJSON(v).(*types.Document)
			if !ok {
				t.Skip()
			}
			if err := d.ValidateData(false); err != nil {
				t.Skip()
			}
		}

		// test MarshalJSON
		{
			b, err := v.MarshalJSON()
			require.NoError(t, err)
			j = string(b)
		}

		// test UnmarshalJSON
		{
			actualV := newFunc()
			tc := testCase{
				sch: &sch,
				j:   j,
			}
			err := unmarshalJSON(v, &tc)
			require.NoError(t, err)
			assertEqual(t, v, actualV)
		}
	})
}

func benchmark(b *testing.B, testCases []testCase, newFunc func() pjsontype) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("UnmarshalJSON", func(b *testing.B) {
				data := []byte(tc.j)
				var v pjsontype
				var err error

				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v = newFunc()
					err = unmarshalJSON(v, &tc)
				}

				b.StopTimer()

				if tc.jErr == "" {
					require.NoError(b, err)
					assertEqual(b, tc.v, v)
					return
				}

				require.Error(b, err)
				require.Equal(b, tc.jErr, lastErr(err).Error())
			})
		})
	}
}

// unmarshalJSON encapsulates type switch and calls UnmarshalJSON or UnmarshalJSONWithSchema on the given value.
func unmarshalJSON(v pjsontype, tc *testCase) error {
	var err error
	switch v := v.(type) {
	case *documentType:
		err = v.UnmarshalJSONWithSchema([]byte(tc.j), tc.sch.Schema)
	case *arrayType:
		err = v.UnmarshalJSONWithSchema([]byte(tc.j), tc.sch.Items)
	case *doubleType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *stringType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *binaryType:
		err = v.UnmarshalJSONWithSchema([]byte(tc.j), tc.sch)
	case *objectIDType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *boolType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *dateTimeType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *regexType:
		err = v.UnmarshalJSONWithSchema([]byte(tc.j), tc.sch)
	case *int32Type:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *timestampType:
		err = v.UnmarshalJSON([]byte(tc.j))
	case *int64Type:
		err = v.UnmarshalJSON([]byte(tc.j))
	default:
		panic(fmt.Sprintf("testing is not implemented for the type %T", v))
	}

	return err
}

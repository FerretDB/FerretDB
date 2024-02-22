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

package bson

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types/fjson"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

type testCase struct {
	name string
	v    bsontype
	b    []byte
	bErr string // unwrapped
}

// assertEqual is assert.Equal that also can compare NaNs and ±0.
func assertEqual(tb testtb.TB, expected, actual any, msgAndArgs ...any) bool {
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

func testBinary(t *testing.T, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.b, "b should not be empty")

			t.Parallel()

			t.Run("ReadFrom", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				br := bytes.NewReader(tc.b)
				bufr := bufio.NewReader(br)
				err := v.ReadFrom(bufr)
				if tc.bErr == "" {
					assert.NoError(t, err)
					assertEqual(t, tc.v, v)
					assert.Zero(t, br.Len(), "not all br bytes were consumed")
					assert.Zero(t, bufr.Buffered(), "not all bufr bytes were consumed")
					return
				}

				require.Error(t, err)
				require.Equal(t, tc.bErr, lastErr(err).Error())
			})

			t.Run("MarshalBinary", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				actualB, err := tc.v.MarshalBinary()
				require.NoError(t, err)
				if !assertEqual(t, tc.b, actualB, "actual:\n%s", hex.Dump(actualB)) {
					// unmarshal again to compare values
					v := newFunc()
					br := bytes.NewReader(actualB)
					bufr := bufio.NewReader(br)
					err := v.ReadFrom(bufr)
					assert.NoError(t, err)
					if assertEqual(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v) {
						t.Log("values are equal after unmarshaling")
					}
					assert.Zero(t, br.Len(), "not all br bytes were consumed")
					assert.Zero(t, bufr.Buffered(), "not all bufr bytes were consumed")
				}
			})

			t.Run("WriteTo", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				var buf bytes.Buffer
				bufw := bufio.NewWriter(&buf)
				err := tc.v.WriteTo(bufw)
				require.NoError(t, err)
				err = bufw.Flush()
				require.NoError(t, err)
				assert.Equal(t, tc.b, buf.Bytes(), "actual:\n%s", hex.Dump(buf.Bytes()))
			})
		})
	}
}

func fuzzBinary(f *testing.F, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		f.Add(tc.b)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		var v bsontype
		var expectedB []byte

		// test ReadFrom
		{
			v = newFunc()
			br := bytes.NewReader(b)
			bufr := bufio.NewReader(br)
			if err := v.ReadFrom(bufr); err != nil {
				t.Skip()
			}

			// remove random tail
			expectedB = b[:len(b)-bufr.Buffered()-br.Len()]
		}

		// test MarshalBinary
		{
			actualB, err := v.MarshalBinary()
			require.NoError(t, err)
			if !assert.Equal(t, expectedB, actualB, "MarshalBinary results differ") {
				// unmarshal again to compare values
				v2 := newFunc()
				br2 := bytes.NewReader(actualB)
				bufr2 := bufio.NewReader(br2)
				err = v2.ReadFrom(bufr2)
				assert.NoError(t, err)
				if assertEqual(t, v, v2, "expected: %s\nactual  : %s", v, v2) {
					t.Log("values are equal after unmarshaling")
				}
				assert.Zero(t, br2.Len(), "not all br bytes were consumed")
				assert.Zero(t, bufr2.Buffered(), "not all bufr bytes were consumed")
			}
		}

		// test WriteTo
		{
			var bw bytes.Buffer
			bufw := bufio.NewWriter(&bw)
			err := v.WriteTo(bufw)
			require.NoError(t, err)
			err = bufw.Flush()
			require.NoError(t, err)
			assert.Equal(t, expectedB, bw.Bytes(), "WriteTo results differ")
		}

		// Test that generated value can be marshaled for logging.
		// Currently, that seems to be the best place to check it since generating values from BSON bytes is very easy.
		{
			// not a "real" type
			if _, ok := v.(*CString); ok {
				t.Skip()
			}

			mB, err := fjson.Marshal(fromBSON(v))
			require.NoError(t, err)
			assert.NotEmpty(t, mB)
		}
	})
}

func benchmark(b *testing.B, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("ReadFrom", func(b *testing.B) {
				br := bytes.NewReader(tc.b)
				var bufr *bufio.Reader
				var v bsontype
				var readErr, seekErr error

				b.ReportAllocs()
				b.SetBytes(br.Size())
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_, seekErr = br.Seek(0, io.SeekStart)

					v = newFunc()
					bufr = bufio.NewReader(br)
					readErr = v.ReadFrom(bufr)
				}

				b.StopTimer()

				require.NoError(b, seekErr)

				if tc.bErr == "" {
					assert.NoError(b, readErr)
					assertEqual(b, tc.v, v)
					assert.Zero(b, br.Len(), "not all br bytes were consumed")
					assert.Zero(b, bufr.Buffered(), "not all bufr bytes were consumed")
					return
				}

				require.Error(b, readErr)
				require.Equal(b, tc.bErr, lastErr(readErr).Error())
			})
		})
	}
}

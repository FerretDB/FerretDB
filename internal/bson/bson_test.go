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
)

type testCase struct {
	name string
	v    bsontype
	b    []byte
	bErr string // unwrapped
}

// assertEqualWithNaN is assert.Equal that also can compare NaNs.
func assertEqualWithNaN(t testing.TB, expected, actual any) {
	t.Helper()

	if expectedD, ok := expected.(*doubleType); ok {
		require.IsType(t, expected, actual)
		actualD := actual.(*doubleType)
		if math.IsNaN(float64(*expectedD)) {
			assert.True(t, math.IsNaN(float64(*actualD)))
			return
		}
	}

	assert.Equal(t, expected, actual, "expected: %s\nactual  : %s", expected, actual)
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
					assertEqualWithNaN(t, tc.v, v)
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
				if !assert.Equal(t, tc.b, actualB, "actual:\n%s", hex.Dump(actualB)) {
					// unmarshal again to compare BSON values
					v := newFunc()
					br := bytes.NewReader(actualB)
					bufr := bufio.NewReader(br)
					err := v.ReadFrom(bufr)
					assert.NoError(t, err)
					if assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v) {
						t.Log("BSON values are equal after unmarshalling")
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
				t.Skip(err)
			}

			// remove random tail
			expectedB = b[:len(b)-bufr.Buffered()-br.Len()]
		}

		// test MarshalBinary
		{
			actualB, err := v.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, expectedB, actualB, "MarshalBinary results differ")
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
					assertEqualWithNaN(b, tc.v, v)
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

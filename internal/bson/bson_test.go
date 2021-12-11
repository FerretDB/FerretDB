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
	"encoding/json"
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

	j      string
	canonJ string // canonical form without extra object fields, zero values, etc.
	jErr   string // unwrapped
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
					if d, ok := v.(*Double); ok && math.IsNaN(float64(*d)) {
						// NaN != NaN, do special handling
						d, ok = tc.v.(*Double)
						assert.True(t, ok)
						assert.True(t, math.IsNaN(float64(*d)))
					} else {
						assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
					}
					assert.Zero(t, br.Len(), "not all br bytes were consumed")
					assert.Zero(t, bufr.Buffered(), "not all bufr bytes were consumed")
					return
				}

				require.Error(t, err)
				for {
					e := errors.Unwrap(err)
					if e == nil {
						break
					}
					err = e
				}
				require.Equal(t, tc.bErr, err.Error())
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

func testJSON(t *testing.T, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			if tc.j == "" {
				t.Skip("j is empty")
			}

			t.Parallel()

			var dst bytes.Buffer
			require.NoError(t, json.Compact(&dst, []byte(tc.j)))
			require.Equal(t, tc.j, dst.String(), "j should be compacted")
			if tc.canonJ != "" {
				dst.Reset()
				require.NoError(t, json.Compact(&dst, []byte(tc.canonJ)))
				require.Equal(t, tc.canonJ, dst.String(), "canonJ should be compacted")
			}

			t.Run("UnmarshalJSON", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				err := v.UnmarshalJSON([]byte(tc.j))
				if tc.jErr == "" {
					require.NoError(t, err)
					if d, ok := v.(*Double); ok && math.IsNaN(float64(*d)) {
						// NaN != NaN, do special handling
						d, ok = tc.v.(*Double)
						assert.True(t, ok)
						assert.True(t, math.IsNaN(float64(*d)))
					} else {
						assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
					}
					return
				}

				require.Error(t, err)
				for {
					e := errors.Unwrap(err)
					if e == nil {
						break
					}
					err = e
				}
				require.Equal(t, tc.jErr, err.Error())
			})

			t.Run("MarshalJSON", func(t *testing.T) {
				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
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

func fuzzJSON(f *testing.F, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		f.Add(tc.j)
		if tc.canonJ != "" {
			f.Add(tc.canonJ)
		}
	}

	f.Fuzz(func(t *testing.T, j string) {
		t.Parallel()

		// raw "null" should never reach UnmarshalJSON due to the way encoding/json works
		if j == "null" {
			t.Skip(j)
		}

		// j may not be a canonical form.
		// We can't compare it with MarshalJSON() result directly.
		// Instead, we compare second results.

		v := newFunc()
		if err := v.UnmarshalJSON([]byte(j)); err != nil {
			t.Skip(err)
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
			err := actualV.UnmarshalJSON([]byte(j))
			require.NoError(t, err)
			if d, ok := v.(*Double); ok && math.IsNaN(float64(*d)) {
				// NaN != NaN, do special handling
				d, ok = actualV.(*Double)
				assert.True(t, ok)
				assert.True(t, math.IsNaN(float64(*d)))
			} else {
				assert.Equal(t, v, actualV, "expected: %s\nactual  : %s", v, actualV)
			}
		}
	})
}

func benchmark(b *testing.B, testCases []testCase, newFunc func() bsontype) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("ReadFrom", func(b *testing.B) {
				br := bytes.NewReader(tc.b)
				var v bsontype
				var readErr, seekErr error

				b.ReportAllocs()
				b.SetBytes(br.Size())
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v = newFunc()
					readErr = v.ReadFrom(bufio.NewReader(br))
					_, seekErr = br.Seek(io.SeekStart, 0)
				}

				b.StopTimer()

				assert.NoError(b, readErr)
				assert.NoError(b, seekErr)
				assert.Equal(b, tc.v, v)
			})

			b.Run("UnmarshalJSON", func(b *testing.B) {
				data := []byte(tc.j)
				var v bsontype
				var err error

				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v = newFunc()
					err = v.UnmarshalJSON(data)
				}

				b.StopTimer()

				assert.NoError(b, err)
				assert.Equal(b, tc.v, v)
			})
		})
	}
}

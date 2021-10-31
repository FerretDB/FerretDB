// Copyright 2021 Baltoro OÜ.
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
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fuzzTestCase struct {
	name string
	v    bsontype
	b    []byte
	j    string
}

func testBinary(t *testing.T, testcases []fuzzTestCase, newFunc func() bsontype) {
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.b, "b should not be empty")

			t.Run("ReadFrom", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				br := bytes.NewReader(tc.b)
				bufr := bufio.NewReader(br)
				err := v.ReadFrom(bufr)
				assert.NoError(t, err)
				assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
				assert.Zero(t, br.Len(), "not all br bytes were consumed")
				assert.Zero(t, bufr.Buffered(), "not all bufr bytes were consumed")
			})

			t.Run("MarshalBinary", func(t *testing.T) {
				t.Parallel()

				actualB, err := tc.v.MarshalBinary()
				require.NoError(t, err)
				assert.Equal(t, tc.b, actualB)
			})

			t.Run("WriteTo", func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				bufw := bufio.NewWriter(&buf)
				err := tc.v.WriteTo(bufw)
				require.NoError(t, err)
				err = bufw.Flush()
				require.NoError(t, err)
				assert.Equal(t, tc.b, buf.Bytes())
			})
		})
	}
}

func fuzzBinary(f *testing.F, testcases []fuzzTestCase, newFunc func() bsontype) {
	for _, tc := range testcases {
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

func testJSON(t *testing.T, testcases []fuzzTestCase, newFunc func() bsontype) {
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.j, "j should not be empty")

			var dst bytes.Buffer
			require.NoError(t, json.Compact(&dst, []byte(tc.j)))
			require.Equal(t, tc.j, dst.String(), "testcase should be compacted")

			t.Run("UnmarshalJSON", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				err := v.UnmarshalJSON([]byte(tc.j))
				require.NoError(t, err)
				assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
			})

			t.Run("MarshalJSON", func(t *testing.T) {
				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, tc.j, string(actualJ))
			})
		})
	}
}

func fuzzJSON(f *testing.F, testcases []fuzzTestCase, newFunc func() bsontype) {
	for _, tc := range testcases {
		f.Add(tc.j)
	}

	f.Fuzz(func(t *testing.T, j string) {
		t.Parallel()

		// compact generated input
		var dst bytes.Buffer
		require.NoError(t, json.Compact(&dst, []byte(j)))
		j = dst.String()

		var v bsontype

		// test UnmarshalJSON
		{
			v = newFunc()
			if err := v.UnmarshalJSON([]byte(j)); err != nil {
				t.Skip(err)
			}
		}

		// test MarshalJSON
		{
			actualJ, err := v.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, j, string(actualJ))
		}
	})
}

func benchmark(b *testing.B, testcases []fuzzTestCase, newFunc func() bsontype) {
	for _, tc := range testcases {
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

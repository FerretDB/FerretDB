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

package wire

import (
	"bufio"
	"bytes"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

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

var lastUpdate = time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

type testCase struct {
	name      string
	headerB   []byte
	bodyB     []byte
	expectedB []byte
	msgHeader *MsgHeader
	msgBody   MsgBody
	command   string // only for OpMsg
	err       string // unwrapped
}

// setExpectedB checks and sets expectedB fields from headerB and bodyB.
func (tc *testCase) setExpectedB(tb testutil.TB) {
	tb.Helper()

	if (len(tc.headerB) == 0) != (len(tc.bodyB) == 0) {
		tb.Fatalf("header dump and body dump are not in sync")
	}

	if (len(tc.headerB) == 0) == (len(tc.expectedB) == 0) {
		tb.Fatalf("header/body dumps and expectedB are not in sync")
	}

	if len(tc.expectedB) == 0 {
		tc.expectedB = make([]byte, 0, len(tc.headerB)+len(tc.bodyB))
		tc.expectedB = append(tc.expectedB, tc.headerB...)
		tc.expectedB = append(tc.expectedB, tc.bodyB...)
		tc.headerB = nil
		tc.bodyB = nil
	}
}

func testMessages(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "name should not be empty")

			tc.setExpectedB(t)

			t.Run("ReadMessage", func(t *testing.T) {
				t.Parallel()

				br := bytes.NewReader(tc.expectedB)
				bufr := bufio.NewReader(br)
				msgHeader, msgBody, err := ReadMessage(bufr)
				if tc.err != "" {
					require.Equal(t, tc.err, lastErr(err).Error())
					return
				}

				assert.NoError(t, err)
				assert.Equal(t, tc.msgHeader, msgHeader)
				assert.Equal(t, tc.msgBody, msgBody)
				assert.Zero(t, br.Len(), "not all br bytes were consumed")
				assert.Zero(t, bufr.Buffered(), "not all bufr bytes were consumed")

				require.NotNil(t, msgHeader)
				require.NotNil(t, msgBody)
				assert.NotPanics(t, func() { _ = msgHeader.String() })
				assert.NotPanics(t, func() { _ = msgBody.String() })

				if msg, ok := tc.msgBody.(*OpMsg); ok {
					d, err := msg.Document()
					require.NoError(t, err)
					assert.Equal(t, tc.command, d.Command())
				}
			})

			t.Run("WriteMessage", func(t *testing.T) {
				if tc.msgHeader == nil {
					t.Skip("msgHeader is nil")
				}

				t.Parallel()

				var buf bytes.Buffer
				bufw := bufio.NewWriter(&buf)
				err := WriteMessage(bufw, tc.msgHeader, tc.msgBody)
				if err != nil {
					require.Equal(t, tc.err, lastErr(err).Error())
					return
				}

				require.NoError(t, err)
				err = bufw.Flush()
				require.NoError(t, err)
				actualB := buf.Bytes()
				require.Equal(t, tc.expectedB, actualB)
			})
		})
	}
}

func fuzzMessages(f *testing.F, testCases []testCase) {
	for _, tc := range testCases {
		tc.setExpectedB(f)
		f.Add(tc.expectedB)
	}

	if !testing.Short() {
		records, err := LoadRecords(filepath.Join("..", "..", "tmp", "records"), 1000)
		require.NoError(f, err)

		for _, rec := range records {
			if rec.HeaderB == nil || rec.BodyB == nil {
				continue
			}

			b := make([]byte, 0, len(rec.HeaderB)+len(rec.BodyB))
			b = append(b, rec.HeaderB...)
			b = append(b, rec.BodyB...)
			f.Add(b)
		}

		f.Logf("%d recorded messages were added to the seed corpus", len(records))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		var msgHeader *MsgHeader
		var msgBody MsgBody
		var expectedB []byte

		// test ReadMessage
		{
			br := bytes.NewReader(b)
			bufr := bufio.NewReader(br)
			var err error
			msgHeader, msgBody, err = ReadMessage(bufr)
			if err != nil {
				t.Skip()
			}

			assert.NotPanics(t, func() { _ = msgHeader.String() })
			assert.NotPanics(t, func() { _ = msgBody.String() })

			if msg, ok := msgBody.(*OpMsg); ok {
				assert.NotPanics(t, func() { msg.Document() })
			}

			// remove random tail
			expectedB = b[:len(b)-bufr.Buffered()-br.Len()]
		}

		// test WriteMessage
		{
			var bw bytes.Buffer
			bufw := bufio.NewWriter(&bw)
			err := WriteMessage(bufw, msgHeader, msgBody)
			require.NoError(t, err)
			err = bufw.Flush()
			require.NoError(t, err)
			assert.Equal(t, expectedB, bw.Bytes())
		}
	})
}

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
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var replyTestCases = []testCase{
	{
		name:    "handshake2",
		headerB: testutil.MustParseDumpFile("testdata", "handshake2_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake2_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 319,
			RequestID:     290,
			ResponseTo:    1,
			OpCode:        OpCodeReply,
		},
		msgBody: &OpReply{
			ResponseFlags: OpReplyFlags(OpReplyAwaitCapable),
			CursorID:      0,
			StartingFrom:  0,
			document: convertDocument(must.NotFail(types.NewDocument(
				"ismaster", true,
				"topologyVersion", must.NotFail(types.NewDocument(
					"processId", types.ObjectID{0x60, 0xfb, 0xed, 0x53, 0x71, 0xfe, 0x1b, 0xae, 0x70, 0x33, 0x95, 0x05},
					"counter", int64(0),
				)),
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(48000000),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Date(2021, time.July, 24, 12, 54, 41, 571000000, time.UTC).Local(),
				"logicalSessionTimeoutMinutes", int32(30),
				"connectionId", int32(28),
				"minWireVersion", int32(0),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			))),
		},
	},
	{
		name:    "handshake4",
		headerB: testutil.MustParseDumpFile("testdata", "handshake4_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake4_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 319,
			RequestID:     291,
			ResponseTo:    2,
			OpCode:        OpCodeReply,
		},
		msgBody: &OpReply{
			ResponseFlags: OpReplyFlags(OpReplyAwaitCapable),
			CursorID:      0,
			StartingFrom:  0,
			document: convertDocument(must.NotFail(types.NewDocument(
				"ismaster", true,
				"topologyVersion", must.NotFail(types.NewDocument(
					"processId", types.ObjectID{0x60, 0xfb, 0xed, 0x53, 0x71, 0xfe, 0x1b, 0xae, 0x70, 0x33, 0x95, 0x05},
					"counter", int64(0),
				)),
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(48000000),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Date(2021, time.July, 24, 12, 54, 41, 592000000, time.UTC).Local(),
				"logicalSessionTimeoutMinutes", int32(30),
				"connectionId", int32(29),
				"minWireVersion", int32(0),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			))),
		},
	},
}

func TestReply(t *testing.T) {
	t.Parallel()
	testMessages(t, replyTestCases)
}

func FuzzReply(f *testing.F) {
	fuzzMessages(f, replyTestCases)
}

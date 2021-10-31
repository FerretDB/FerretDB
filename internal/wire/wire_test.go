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

package wire

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
)

type fuzzTestCase struct {
	name      string
	headerB   []byte
	bodyB     []byte
	expectedB []byte
	msgHeader *MsgHeader
	msgBody   MsgBody
}

func testMessage(t testing.TB, b []byte) {
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
			t.Skip(err)
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
}

func fuzzMessage(f *testing.F, testcases []fuzzTestCase) {
	for _, tc := range testcases {
		f.Log(tc.name)

		if (len(tc.headerB) == 0) != (len(tc.bodyB) == 0) {
			f.Fatalf("header dump and body dump are not in sync")
		}
		if (len(tc.headerB) == 0) == (len(tc.expectedB) == 0) {
			f.Fatalf("header/body dumps and expectedB are not in sync")
		}

		if len(tc.expectedB) == 0 {
			expectedB := make([]byte, 0, len(tc.headerB)+len(tc.bodyB))
			expectedB = append(expectedB, tc.headerB...)
			expectedB = append(expectedB, tc.bodyB...)
			tc.expectedB = expectedB
		}

		f.Add(tc.expectedB)

		// test ReadMessage
		{
			br := bytes.NewReader(tc.expectedB)
			bufr := bufio.NewReader(br)
			msgHeader, msgBody, err := ReadMessage(bufr)
			require.NoError(f, err, "case %s", tc.name)
			assert.Equal(f, tc.msgHeader, msgHeader, "case %s", tc.name)
			assert.Equal(f, tc.msgBody, msgBody, "case %s", tc.name)
			assert.Zero(f, br.Len(), "not all br bytes were consumed")
			assert.Zero(f, bufr.Buffered(), "not all bufr bytes were consumed")
		}

		// test WriteMessage
		{
			var buf bytes.Buffer
			bufw := bufio.NewWriter(&buf)
			err := WriteMessage(bufw, tc.msgHeader, tc.msgBody)
			require.NoError(f, err, "case %s", tc.name)
			err = bufw.Flush()
			require.NoError(f, err, "case %s", tc.name)
			actualB := buf.Bytes()
			f.Add(actualB)
			require.Equal(f, tc.expectedB, actualB, "case %s", tc.name)
		}
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()
		testMessage(t, b)
	})
}

var testcases = []fuzzTestCase{{
	name:    "handshake1",
	headerB: testutil.MustParseDumpFile("testdata", "handshake1_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake1_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 372,
		RequestID:     1,
		ResponseTo:    0,
		OpCode:        OP_QUERY,
	},
	msgBody: &OpQuery{
		Flags:              0,
		FullCollectionName: "admin.$cmd",
		NumberToSkip:       0,
		NumberToReturn:     -1,
		Query: types.MakeDocument(
			"ismaster", true,
			"client", types.MakeDocument(
				"driver", types.MakeDocument(
					"name", "nodejs",
					"version", "4.0.0-beta.6",
				),
				"os", types.MakeDocument(
					"type", "Darwin",
					"name", "darwin",
					"architecture", "x64",
					"version", "20.6.0",
				),
				"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
				"application", types.MakeDocument(
					"name", "mongosh 1.0.1",
				),
			),
			"compression", types.Array{"none"},
			"loadBalanced", false,
		),
		ReturnFieldsSelector: nil,
	},
}, {
	name:    "handshake2",
	headerB: testutil.MustParseDumpFile("testdata", "handshake2_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake2_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 319,
		RequestID:     290,
		ResponseTo:    1,
		OpCode:        OP_REPLY,
	},
	msgBody: &OpReply{
		ResponseFlags:  OpReplyFlags(OpReplyAwaitCapable),
		CursorID:       0,
		StartingFrom:   0,
		NumberReturned: 1,
		Documents: []types.Document{types.MakeDocument(
			"ismaster", true,
			"topologyVersion", types.MakeDocument(
				"processId", types.ObjectID{0x60, 0xfb, 0xed, 0x53, 0x71, 0xfe, 0x1b, 0xae, 0x70, 0x33, 0x95, 0x05},
				"counter", int64(0),
			),
			"maxBsonObjectSize", int32(16777216),
			"maxMessageSizeBytes", int32(48000000),
			"maxWriteBatchSize", int32(100000),
			"localTime", time.Date(2021, time.July, 24, 12, 54, 41, 571000000, time.UTC),
			"logicalSessionTimeoutMinutes", int32(30),
			"connectionId", int32(28),
			"minWireVersion", int32(0),
			"maxWireVersion", int32(13),
			"readOnly", false,
			"ok", float64(1),
		)},
	},
}, {
	name:    "handshake3",
	headerB: testutil.MustParseDumpFile("testdata", "handshake3_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake3_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 372,
		RequestID:     2,
		ResponseTo:    0,
		OpCode:        OP_QUERY,
	},
	msgBody: &OpQuery{
		Flags:              0,
		FullCollectionName: "admin.$cmd",
		NumberToSkip:       0,
		NumberToReturn:     -1,
		Query: types.MakeDocument(
			"ismaster", true,
			"client", types.MakeDocument(
				"driver", types.MakeDocument(
					"name", "nodejs",
					"version", "4.0.0-beta.6",
				),
				"os", types.MakeDocument(
					"type", "Darwin",
					"name", "darwin",
					"architecture", "x64",
					"version", "20.6.0",
				),
				"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
				"application", types.MakeDocument(
					"name", "mongosh 1.0.1",
				),
			),
			"compression", types.Array{"none"},
			"loadBalanced", false,
		),
		ReturnFieldsSelector: nil,
	},
}, {
	name:    "handshake4",
	headerB: testutil.MustParseDumpFile("testdata", "handshake4_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake4_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 319,
		RequestID:     291,
		ResponseTo:    2,
		OpCode:        OP_REPLY,
	},
	msgBody: &OpReply{
		ResponseFlags:  OpReplyFlags(OpReplyAwaitCapable),
		CursorID:       0,
		StartingFrom:   0,
		NumberReturned: 1,
		Documents: []types.Document{types.MakeDocument(
			"ismaster", true,
			"topologyVersion", types.MakeDocument(
				"processId", types.ObjectID{0x60, 0xfb, 0xed, 0x53, 0x71, 0xfe, 0x1b, 0xae, 0x70, 0x33, 0x95, 0x05},
				"counter", int64(0),
			),
			"maxBsonObjectSize", int32(16777216),
			"maxMessageSizeBytes", int32(48000000),
			"maxWriteBatchSize", int32(100000),
			"localTime", time.Date(2021, time.July, 24, 12, 54, 41, 592000000, time.UTC),
			"logicalSessionTimeoutMinutes", int32(30),
			"connectionId", int32(29),
			"minWireVersion", int32(0),
			"maxWireVersion", int32(13),
			"readOnly", false,
			"ok", float64(1),
		)},
	},
}, {
	name:    "handshake5",
	headerB: testutil.MustParseDumpFile("testdata", "handshake5_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake5_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 92,
		RequestID:     3,
		ResponseTo:    0,
		OpCode:        OP_MSG,
	},
	msgBody: &OpMsg{
		FlagBits: 0,
		Documents: []types.Document{types.MakeDocument(
			"buildInfo", int32(1),
			"lsid", types.MakeDocument(
				"id", types.Binary{
					Subtype: types.BinaryUUID,
					B:       []byte{0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7, 0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe},
				},
			),
			"$db", "admin",
		)},
	},
}, {
	name:    "handshake6",
	headerB: testutil.MustParseDumpFile("testdata", "handshake6_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake6_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 1931,
		RequestID:     292,
		ResponseTo:    3,
		OpCode:        OP_MSG,
	},
	msgBody: &OpMsg{
		FlagBits: 0,
		Documents: []types.Document{types.MakeDocument(
			"version", "5.0.0",
			"gitVersion", "1184f004a99660de6f5e745573419bda8a28c0e9",
			"modules", types.Array{},
			"allocator", "tcmalloc",
			"javascriptEngine", "mozjs",
			"sysInfo", "deprecated",
			"versionArray", types.Array{int32(5), int32(0), int32(0), int32(0)},
			"openssl", types.MakeDocument(
				"running", "OpenSSL 1.1.1f  31 Mar 2020",
				"compiled", "OpenSSL 1.1.1f  31 Mar 2020",
			),
			"buildEnvironment", types.MakeDocument(
				"distmod", "ubuntu2004",
				"distarch", "x86_64",
				"cc", "/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",
				"ccflags", "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables "+
					"-ggdb -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer "+
					"-fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 "+
					"-Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations "+
					"-Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces "+
					"-fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp",
				"cxx", "/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",
				"cxxflags", "-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",
				"linkflags", "-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong "+
					"-Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack "+
					"-Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",
				"target_arch", "x86_64",
				"target_os", "linux",
				"cppdefines", "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 "+
					"_GNU_SOURCE _REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 "+
					"BOOST_THREAD_USES_DATETIME BOOST_SYSTEM_NO_DEPRECATED "+
					"BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER "+
					"BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG "+
					"BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
			),
			"bits", int32(64),
			"debug", false,
			"maxBsonObjectSize", int32(16777216),
			"storageEngines", types.Array{"devnull", "ephemeralForTest", "wiredTiger"},
			"ok", float64(1),
		)},
	},
}}

func FuzzMessage(f *testing.F) {
	fuzzMessage(f, testcases)
}

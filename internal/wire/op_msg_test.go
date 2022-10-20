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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var msgTestCases = []testCase{{
	name:    "handshake5",
	headerB: testutil.MustParseDumpFile("testdata", "handshake5_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake5_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 92,
		RequestID:     3,
		OpCode:        OpCodeMsg,
	},
	msgBody: &OpMsg{
		sections: []OpMsgSection{{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"buildInfo", int32(1),
				"lsid", must.NotFail(types.NewDocument(
					"id", types.Binary{
						Subtype: types.BinaryUUID,
						B: []byte{
							0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7,
							0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe,
						},
					},
				)),
				"$db", "admin",
			))},
		}},
	},
}, {
	name:    "handshake6",
	headerB: testutil.MustParseDumpFile("testdata", "handshake6_header.hex"),
	bodyB:   testutil.MustParseDumpFile("testdata", "handshake6_body.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 1931,
		RequestID:     292,
		ResponseTo:    3,
		OpCode:        OpCodeMsg,
	},
	msgBody: &OpMsg{
		sections: []OpMsgSection{{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"version", "5.0.0",
				"gitVersion", "1184f004a99660de6f5e745573419bda8a28c0e9",
				"modules", must.NotFail(types.NewArray()),
				"allocator", "tcmalloc",
				"javascriptEngine", "mozjs",
				"sysInfo", "deprecated",
				"versionArray", must.NotFail(types.NewArray(int32(5), int32(0), int32(0), int32(0))),
				"openssl", must.NotFail(types.NewDocument(
					"running", "OpenSSL 1.1.1f  31 Mar 2020",
					"compiled", "OpenSSL 1.1.1f  31 Mar 2020",
				)),
				"buildEnvironment", must.NotFail(types.NewDocument(
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
				)),
				"bits", int32(64),
				"debug", false,
				"maxBsonObjectSize", int32(16777216),
				"storageEngines", must.NotFail(types.NewArray("devnull", "ephemeralForTest", "wiredTiger")),
				"ok", float64(1),
			))},
		}},
	},
}, {
	name:      "import",
	expectedB: testutil.MustParseDumpFile("testdata", "import.hex"),
	msgHeader: &MsgHeader{
		MessageLength: 327,
		RequestID:     7,
		OpCode:        OpCodeMsg,
	},
	msgBody: &OpMsg{
		sections: []OpMsgSection{{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"insert", "actor",
				"ordered", true,
				"writeConcern", must.NotFail(types.NewDocument(
					"w", "majority",
				)),
				"$db", "monila",
			))},
		}, {
			Kind:       1,
			Identifier: "documents",
			Documents: []*types.Document{
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
					"actor_id", int32(1),
					"first_name", "PENELOPE",
					"last_name", "GUINESS",
					"last_update", lastUpdate,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				)),
			},
		}},
	},
}, {
	name:      "msg_fuzz1",
	expectedB: testutil.MustParseDumpFile("testdata", "msg_fuzz1.hex"),
	err:       `wire.OpMsg.readFrom: invalid kind 1 section length -13619152`,
}}

func TestMsg(t *testing.T) {
	t.Parallel()
	testMessages(t, msgTestCases)
}

func FuzzMsg(f *testing.F) {
	fuzzMessages(f, msgTestCases)
}

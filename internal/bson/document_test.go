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
	"testing"
	"time"

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
)

var (
	handshake1 = fuzzTestCase{
		name: "handshake1",
		v: NewDocument(types.MakeDocument(
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
		)),
		b: testutil.MustParseDump(`
			00000000  4d 01 00 00 08 69 73 6d  61 73 74 65 72 00 01 03  |M....ismaster...|
			00000010  63 6c 69 65 6e 74 00 08  01 00 00 03 64 72 69 76  |client......driv|
			00000020  65 72 00 30 00 00 00 02  6e 61 6d 65 00 07 00 00  |er.0....name....|
			00000030  00 6e 6f 64 65 6a 73 00  02 76 65 72 73 69 6f 6e  |.nodejs..version|
			00000040  00 0d 00 00 00 34 2e 30  2e 30 2d 62 65 74 61 2e  |.....4.0.0-beta.|
			00000050  36 00 00 03 6f 73 00 51  00 00 00 02 74 79 70 65  |6...os.Q....type|
			00000060  00 07 00 00 00 44 61 72  77 69 6e 00 02 6e 61 6d  |.....Darwin..nam|
			00000070  65 00 07 00 00 00 64 61  72 77 69 6e 00 02 61 72  |e.....darwin..ar|
			00000080  63 68 69 74 65 63 74 75  72 65 00 04 00 00 00 78  |chitecture.....x|
			00000090  36 34 00 02 76 65 72 73  69 6f 6e 00 07 00 00 00  |64..version.....|
			000000a0  32 30 2e 36 2e 30 00 00  02 70 6c 61 74 66 6f 72  |20.6.0...platfor|
			000000b0  6d 00 3e 00 00 00 4e 6f  64 65 2e 6a 73 20 76 31  |m.>...Node.js v1|
			000000c0  34 2e 31 37 2e 33 2c 20  4c 45 20 28 75 6e 69 66  |4.17.3, LE (unif|
			000000d0  69 65 64 29 7c 4e 6f 64  65 2e 6a 73 20 76 31 34  |ied)|Node.js v14|
			000000e0  2e 31 37 2e 33 2c 20 4c  45 20 28 75 6e 69 66 69  |.17.3, LE (unifi|
			000000f0  65 64 29 00 03 61 70 70  6c 69 63 61 74 69 6f 6e  |ed)..application|
			00000100  00 1d 00 00 00 02 6e 61  6d 65 00 0e 00 00 00 6d  |......name.....m|
			00000110  6f 6e 67 6f 73 68 20 31  2e 30 2e 31 00 00 00 04  |ongosh 1.0.1....|
			00000120  63 6f 6d 70 72 65 73 73  69 6f 6e 00 11 00 00 00  |compression.....|
			00000130  02 30 00 05 00 00 00 6e  6f 6e 65 00 00 08 6c 6f  |.0.....none...lo|
			00000140  61 64 42 61 6c 61 6e 63  65 64 00 00 00           |adBalanced...|`,
		),
		j: `{"$k":["ismaster","client","compression","loadBalanced"],"ismaster":true,"client":{"$k":["driver","os","platform","application"],"driver":{"$k":["name","version"],"name":"nodejs","version":"4.0.0-beta.6"},"os":{"$k":["type","name","architecture","version"],"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)","application":{"$k":["name"],"name":"mongosh 1.0.1"}},"compression":["none"],"loadBalanced":false}`,
	}

	handshake2 = fuzzTestCase{
		name: "handshake2",
		v: NewDocument(types.MakeDocument(
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
		)),
		b: testutil.MustParseDump(`
			00000000  4d 01 00 00 08 69 73 6d  61 73 74 65 72 00 01 03  |M....ismaster...|
			00000010  63 6c 69 65 6e 74 00 08  01 00 00 03 64 72 69 76  |client......driv|
			00000020  65 72 00 30 00 00 00 02  6e 61 6d 65 00 07 00 00  |er.0....name....|
			00000030  00 6e 6f 64 65 6a 73 00  02 76 65 72 73 69 6f 6e  |.nodejs..version|
			00000040  00 0d 00 00 00 34 2e 30  2e 30 2d 62 65 74 61 2e  |.....4.0.0-beta.|
			00000050  36 00 00 03 6f 73 00 51  00 00 00 02 74 79 70 65  |6...os.Q....type|
			00000060  00 07 00 00 00 44 61 72  77 69 6e 00 02 6e 61 6d  |.....Darwin..nam|
			00000070  65 00 07 00 00 00 64 61  72 77 69 6e 00 02 61 72  |e.....darwin..ar|
			00000080  63 68 69 74 65 63 74 75  72 65 00 04 00 00 00 78  |chitecture.....x|
			00000090  36 34 00 02 76 65 72 73  69 6f 6e 00 07 00 00 00  |64..version.....|
			000000a0  32 30 2e 36 2e 30 00 00  02 70 6c 61 74 66 6f 72  |20.6.0...platfor|
			000000b0  6d 00 3e 00 00 00 4e 6f  64 65 2e 6a 73 20 76 31  |m.>...Node.js v1|
			000000c0  34 2e 31 37 2e 33 2c 20  4c 45 20 28 75 6e 69 66  |4.17.3, LE (unif|
			000000d0  69 65 64 29 7c 4e 6f 64  65 2e 6a 73 20 76 31 34  |ied)|Node.js v14|
			000000e0  2e 31 37 2e 33 2c 20 4c  45 20 28 75 6e 69 66 69  |.17.3, LE (unifi|
			000000f0  65 64 29 00 03 61 70 70  6c 69 63 61 74 69 6f 6e  |ed)..application|
			00000100  00 1d 00 00 00 02 6e 61  6d 65 00 0e 00 00 00 6d  |......name.....m|
			00000110  6f 6e 67 6f 73 68 20 31  2e 30 2e 31 00 00 00 04  |ongosh 1.0.1....|
			00000120  63 6f 6d 70 72 65 73 73  69 6f 6e 00 11 00 00 00  |compression.....|
			00000130  02 30 00 05 00 00 00 6e  6f 6e 65 00 00 08 6c 6f  |.0.....none...lo|
			00000140  61 64 42 61 6c 61 6e 63  65 64 00 00 00           |adBalanced...|`,
		),
		j: `{"$k":["ismaster","client","compression","loadBalanced"],"ismaster":true,"client":{"$k":["driver","os","platform","application"],"driver":{"$k":["name","version"],"name":"nodejs","version":"4.0.0-beta.6"},"os":{"$k":["type","name","architecture","version"],"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)","application":{"$k":["name"],"name":"mongosh 1.0.1"}},"compression":["none"],"loadBalanced":false}`,
	}

	handshake3 = fuzzTestCase{
		name: "handshake3",
		v: NewDocument(types.MakeDocument(
			"buildInfo", int32(1),
			"lsid", types.MakeDocument(
				"id", types.Binary{
					Subtype: types.BinaryUUID,
					B:       []byte{0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7, 0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe},
				},
			),
			"$db", "admin",
		)),
		b: testutil.MustParseDump(`
			00000000  47 00 00 00 10 62 75 69  6c 64 49 6e 66 6f 00 01  |G....buildInfo..|
			00000010  00 00 00 03 6c 73 69 64  00 1e 00 00 00 05 69 64  |....lsid......id|
			00000020  00 10 00 00 00 04 a3 19  f2 b4 a1 75 40 c7 b8 e7  |...........u@...|
			00000030  a3 a3 2e c2 56 be 00 02  24 64 62 00 06 00 00 00  |....V...$db.....|
			00000040  61 64 6d 69 6e 00 00                              |admin..|`,
		),
		j: `{"$k":["buildInfo","lsid","$db"],"buildInfo":1,"lsid":{"$k":["id"],"id":{"$b":"oxnytKF1QMe456OjLsJWvg==","s":4}},"$db":"admin"}`,
	}

	handshake4 = fuzzTestCase{
		name: "handshake4",
		v: NewDocument(types.MakeDocument(
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
				"ccflags", "-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb "+
					"-Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer "+
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
				"cppdefines", "SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE "+
					"_REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME "+
					"BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS "+
					"BOOST_ENABLE_ASSERT_DEBUG_HANDLER BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG "+
					"BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS",
			),
			"bits", int32(64),
			"debug", false,
			"maxBsonObjectSize", int32(16777216),
			"storageEngines", types.Array{"devnull", "ephemeralForTest", "wiredTiger"},
			"ok", float64(1),
		)),
		b: testutil.MustParseDump(`
			00000000  76 07 00 00 02 76 65 72  73 69 6f 6e 00 06 00 00  |v....version....|
			00000010  00 35 2e 30 2e 30 00 02  67 69 74 56 65 72 73 69  |.5.0.0..gitVersi|
			00000020  6f 6e 00 29 00 00 00 31  31 38 34 66 30 30 34 61  |on.)...1184f004a|
			00000030  39 39 36 36 30 64 65 36  66 35 65 37 34 35 35 37  |99660de6f5e74557|
			00000040  33 34 31 39 62 64 61 38  61 32 38 63 30 65 39 00  |3419bda8a28c0e9.|
			00000050  04 6d 6f 64 75 6c 65 73  00 05 00 00 00 00 02 61  |.modules.......a|
			00000060  6c 6c 6f 63 61 74 6f 72  00 09 00 00 00 74 63 6d  |llocator.....tcm|
			00000070  61 6c 6c 6f 63 00 02 6a  61 76 61 73 63 72 69 70  |alloc..javascrip|
			00000080  74 45 6e 67 69 6e 65 00  06 00 00 00 6d 6f 7a 6a  |tEngine.....mozj|
			00000090  73 00 02 73 79 73 49 6e  66 6f 00 0b 00 00 00 64  |s..sysInfo.....d|
			000000a0  65 70 72 65 63 61 74 65  64 00 04 76 65 72 73 69  |eprecated..versi|
			000000b0  6f 6e 41 72 72 61 79 00  21 00 00 00 10 30 00 05  |onArray.!....0..|
			000000c0  00 00 00 10 31 00 00 00  00 00 10 32 00 00 00 00  |....1......2....|
			000000d0  00 10 33 00 00 00 00 00  00 03 6f 70 65 6e 73 73  |..3.......openss|
			000000e0  6c 00 58 00 00 00 02 72  75 6e 6e 69 6e 67 00 1c  |l.X....running..|
			000000f0  00 00 00 4f 70 65 6e 53  53 4c 20 31 2e 31 2e 31  |...OpenSSL 1.1.1|
			00000100  66 20 20 33 31 20 4d 61  72 20 32 30 32 30 00 02  |f  31 Mar 2020..|
			00000110  63 6f 6d 70 69 6c 65 64  00 1c 00 00 00 4f 70 65  |compiled.....Ope|
			00000120  6e 53 53 4c 20 31 2e 31  2e 31 66 20 20 33 31 20  |nSSL 1.1.1f  31 |
			00000130  4d 61 72 20 32 30 32 30  00 00 03 62 75 69 6c 64  |Mar 2020...build|
			00000140  45 6e 76 69 72 6f 6e 6d  65 6e 74 00 a6 05 00 00  |Environment.....|
			00000150  02 64 69 73 74 6d 6f 64  00 0b 00 00 00 75 62 75  |.distmod.....ubu|
			00000160  6e 74 75 32 30 30 34 00  02 64 69 73 74 61 72 63  |ntu2004..distarc|
			00000170  68 00 07 00 00 00 78 38  36 5f 36 34 00 02 63 63  |h.....x86_64..cc|
			00000180  00 32 00 00 00 2f 6f 70  74 2f 6d 6f 6e 67 6f 64  |.2.../opt/mongod|
			00000190  62 74 6f 6f 6c 63 68 61  69 6e 2f 76 33 2f 62 69  |btoolchain/v3/bi|
			000001a0  6e 2f 67 63 63 3a 20 67  63 63 20 28 47 43 43 29  |n/gcc: gcc (GCC)|
			000001b0  20 38 2e 35 2e 30 00 02  63 63 66 6c 61 67 73 00  | 8.5.0..ccflags.|
			000001c0  d6 01 00 00 2d 57 65 72  72 6f 72 20 2d 69 6e 63  |....-Werror -inc|
			000001d0  6c 75 64 65 20 6d 6f 6e  67 6f 2f 70 6c 61 74 66  |lude mongo/platf|
			000001e0  6f 72 6d 2f 62 61 73 69  63 2e 68 20 2d 66 61 73  |orm/basic.h -fas|
			000001f0  79 6e 63 68 72 6f 6e 6f  75 73 2d 75 6e 77 69 6e  |ynchronous-unwin|
			00000200  64 2d 74 61 62 6c 65 73  20 2d 67 67 64 62 20 2d  |d-tables -ggdb -|
			00000210  57 61 6c 6c 20 2d 57 73  69 67 6e 2d 63 6f 6d 70  |Wall -Wsign-comp|
			00000220  61 72 65 20 2d 57 6e 6f  2d 75 6e 6b 6e 6f 77 6e  |are -Wno-unknown|
			00000230  2d 70 72 61 67 6d 61 73  20 2d 57 69 6e 76 61 6c  |-pragmas -Winval|
			00000240  69 64 2d 70 63 68 20 2d  66 6e 6f 2d 6f 6d 69 74  |id-pch -fno-omit|
			00000250  2d 66 72 61 6d 65 2d 70  6f 69 6e 74 65 72 20 2d  |-frame-pointer -|
			00000260  66 6e 6f 2d 73 74 72 69  63 74 2d 61 6c 69 61 73  |fno-strict-alias|
			00000270  69 6e 67 20 2d 4f 32 20  2d 6d 61 72 63 68 3d 73  |ing -O2 -march=s|
			00000280  61 6e 64 79 62 72 69 64  67 65 20 2d 6d 74 75 6e  |andybridge -mtun|
			00000290  65 3d 67 65 6e 65 72 69  63 20 2d 6d 70 72 65 66  |e=generic -mpref|
			000002a0  65 72 2d 76 65 63 74 6f  72 2d 77 69 64 74 68 3d  |er-vector-width=|
			000002b0  31 32 38 20 2d 57 6e 6f  2d 75 6e 75 73 65 64 2d  |128 -Wno-unused-|
			000002c0  6c 6f 63 61 6c 2d 74 79  70 65 64 65 66 73 20 2d  |local-typedefs -|
			000002d0  57 6e 6f 2d 75 6e 75 73  65 64 2d 66 75 6e 63 74  |Wno-unused-funct|
			000002e0  69 6f 6e 20 2d 57 6e 6f  2d 64 65 70 72 65 63 61  |ion -Wno-depreca|
			000002f0  74 65 64 2d 64 65 63 6c  61 72 61 74 69 6f 6e 73  |ted-declarations|
			00000300  20 2d 57 6e 6f 2d 75 6e  75 73 65 64 2d 63 6f 6e  | -Wno-unused-con|
			00000310  73 74 2d 76 61 72 69 61  62 6c 65 20 2d 57 6e 6f  |st-variable -Wno|
			00000320  2d 75 6e 75 73 65 64 2d  62 75 74 2d 73 65 74 2d  |-unused-but-set-|
			00000330  76 61 72 69 61 62 6c 65  20 2d 57 6e 6f 2d 6d 69  |variable -Wno-mi|
			00000340  73 73 69 6e 67 2d 62 72  61 63 65 73 20 2d 66 73  |ssing-braces -fs|
			00000350  74 61 63 6b 2d 70 72 6f  74 65 63 74 6f 72 2d 73  |tack-protector-s|
			00000360  74 72 6f 6e 67 20 2d 57  61 2c 2d 2d 6e 6f 63 6f  |trong -Wa,--noco|
			00000370  6d 70 72 65 73 73 2d 64  65 62 75 67 2d 73 65 63  |mpress-debug-sec|
			00000380  74 69 6f 6e 73 20 2d 66  6e 6f 2d 62 75 69 6c 74  |tions -fno-built|
			00000390  69 6e 2d 6d 65 6d 63 6d  70 00 02 63 78 78 00 32  |in-memcmp..cxx.2|
			000003a0  00 00 00 2f 6f 70 74 2f  6d 6f 6e 67 6f 64 62 74  |.../opt/mongodbt|
			000003b0  6f 6f 6c 63 68 61 69 6e  2f 76 33 2f 62 69 6e 2f  |oolchain/v3/bin/|
			000003c0  67 2b 2b 3a 20 67 2b 2b  20 28 47 43 43 29 20 38  |g++: g++ (GCC) 8|
			000003d0  2e 35 2e 30 00 02 63 78  78 66 6c 61 67 73 00 4e  |.5.0..cxxflags.N|
			000003e0  00 00 00 2d 57 6f 76 65  72 6c 6f 61 64 65 64 2d  |...-Woverloaded-|
			000003f0  76 69 72 74 75 61 6c 20  2d 57 6e 6f 2d 6d 61 79  |virtual -Wno-may|
			00000400  62 65 2d 75 6e 69 6e 69  74 69 61 6c 69 7a 65 64  |be-uninitialized|
			00000410  20 2d 66 73 69 7a 65 64  2d 64 65 61 6c 6c 6f 63  | -fsized-dealloc|
			00000420  61 74 69 6f 6e 20 2d 73  74 64 3d 63 2b 2b 31 37  |ation -std=c++17|
			00000430  00 02 6c 69 6e 6b 66 6c  61 67 73 00 02 01 00 00  |..linkflags.....|
			00000440  2d 57 6c 2c 2d 2d 66 61  74 61 6c 2d 77 61 72 6e  |-Wl,--fatal-warn|
			00000450  69 6e 67 73 20 2d 70 74  68 72 65 61 64 20 2d 57  |ings -pthread -W|
			00000460  6c 2c 2d 7a 2c 6e 6f 77  20 2d 66 75 73 65 2d 6c  |l,-z,now -fuse-l|
			00000470  64 3d 67 6f 6c 64 20 2d  66 73 74 61 63 6b 2d 70  |d=gold -fstack-p|
			00000480  72 6f 74 65 63 74 6f 72  2d 73 74 72 6f 6e 67 20  |rotector-strong |
			00000490  2d 57 6c 2c 2d 2d 6e 6f  2d 74 68 72 65 61 64 73  |-Wl,--no-threads|
			000004a0  20 2d 57 6c 2c 2d 2d 62  75 69 6c 64 2d 69 64 20  | -Wl,--build-id |
			000004b0  2d 57 6c 2c 2d 2d 68 61  73 68 2d 73 74 79 6c 65  |-Wl,--hash-style|
			000004c0  3d 67 6e 75 20 2d 57 6c  2c 2d 7a 2c 6e 6f 65 78  |=gnu -Wl,-z,noex|
			000004d0  65 63 73 74 61 63 6b 20  2d 57 6c 2c 2d 2d 77 61  |ecstack -Wl,--wa|
			000004e0  72 6e 2d 65 78 65 63 73  74 61 63 6b 20 2d 57 6c  |rn-execstack -Wl|
			000004f0  2c 2d 7a 2c 72 65 6c 72  6f 20 2d 57 6c 2c 2d 2d  |,-z,relro -Wl,--|
			00000500  63 6f 6d 70 72 65 73 73  2d 64 65 62 75 67 2d 73  |compress-debug-s|
			00000510  65 63 74 69 6f 6e 73 3d  6e 6f 6e 65 20 2d 57 6c  |ections=none -Wl|
			00000520  2c 2d 7a 2c 6f 72 69 67  69 6e 20 2d 57 6c 2c 2d  |,-z,origin -Wl,-|
			00000530  2d 65 6e 61 62 6c 65 2d  6e 65 77 2d 64 74 61 67  |-enable-new-dtag|
			00000540  73 00 02 74 61 72 67 65  74 5f 61 72 63 68 00 07  |s..target_arch..|
			00000550  00 00 00 78 38 36 5f 36  34 00 02 74 61 72 67 65  |...x86_64..targe|
			00000560  74 5f 6f 73 00 06 00 00  00 6c 69 6e 75 78 00 02  |t_os.....linux..|
			00000570  63 70 70 64 65 66 69 6e  65 73 00 72 01 00 00 53  |cppdefines.r...S|
			00000580  41 46 45 49 4e 54 5f 55  53 45 5f 49 4e 54 52 49  |AFEINT_USE_INTRI|
			00000590  4e 53 49 43 53 20 30 20  50 43 52 45 5f 53 54 41  |NSICS 0 PCRE_STA|
			000005a0  54 49 43 20 4e 44 45 42  55 47 20 5f 58 4f 50 45  |TIC NDEBUG _XOPE|
			000005b0  4e 5f 53 4f 55 52 43 45  20 37 30 30 20 5f 47 4e  |N_SOURCE 700 _GN|
			000005c0  55 5f 53 4f 55 52 43 45  20 5f 52 45 45 4e 54 52  |U_SOURCE _REENTR|
			000005d0  41 4e 54 20 31 20 5f 46  4f 52 54 49 46 59 5f 53  |ANT 1 _FORTIFY_S|
			000005e0  4f 55 52 43 45 20 32 20  42 4f 4f 53 54 5f 54 48  |OURCE 2 BOOST_TH|
			000005f0  52 45 41 44 5f 56 45 52  53 49 4f 4e 20 35 20 42  |READ_VERSION 5 B|
			00000600  4f 4f 53 54 5f 54 48 52  45 41 44 5f 55 53 45 53  |OOST_THREAD_USES|
			00000610  5f 44 41 54 45 54 49 4d  45 20 42 4f 4f 53 54 5f  |_DATETIME BOOST_|
			00000620  53 59 53 54 45 4d 5f 4e  4f 5f 44 45 50 52 45 43  |SYSTEM_NO_DEPREC|
			00000630  41 54 45 44 20 42 4f 4f  53 54 5f 4d 41 54 48 5f  |ATED BOOST_MATH_|
			00000640  4e 4f 5f 4c 4f 4e 47 5f  44 4f 55 42 4c 45 5f 4d  |NO_LONG_DOUBLE_M|
			00000650  41 54 48 5f 46 55 4e 43  54 49 4f 4e 53 20 42 4f  |ATH_FUNCTIONS BO|
			00000660  4f 53 54 5f 45 4e 41 42  4c 45 5f 41 53 53 45 52  |OST_ENABLE_ASSER|
			00000670  54 5f 44 45 42 55 47 5f  48 41 4e 44 4c 45 52 20  |T_DEBUG_HANDLER |
			00000680  42 4f 4f 53 54 5f 4c 4f  47 5f 4e 4f 5f 53 48 4f  |BOOST_LOG_NO_SHO|
			00000690  52 54 48 41 4e 44 5f 4e  41 4d 45 53 20 42 4f 4f  |RTHAND_NAMES BOO|
			000006a0  53 54 5f 4c 4f 47 5f 55  53 45 5f 4e 41 54 49 56  |ST_LOG_USE_NATIV|
			000006b0  45 5f 53 59 53 4c 4f 47  20 42 4f 4f 53 54 5f 4c  |E_SYSLOG BOOST_L|
			000006c0  4f 47 5f 57 49 54 48 4f  55 54 5f 54 48 52 45 41  |OG_WITHOUT_THREA|
			000006d0  44 5f 41 54 54 52 20 41  42 53 4c 5f 46 4f 52 43  |D_ATTR ABSL_FORC|
			000006e0  45 5f 41 4c 49 47 4e 45  44 5f 41 43 43 45 53 53  |E_ALIGNED_ACCESS|
			000006f0  00 00 10 62 69 74 73 00  40 00 00 00 08 64 65 62  |...bits.@....deb|
			00000700  75 67 00 00 10 6d 61 78  42 73 6f 6e 4f 62 6a 65  |ug...maxBsonObje|
			00000710  63 74 53 69 7a 65 00 00  00 00 01 04 73 74 6f 72  |ctSize......stor|
			00000720  61 67 65 45 6e 67 69 6e  65 73 00 3e 00 00 00 02  |ageEngines.>....|
			00000730  30 00 08 00 00 00 64 65  76 6e 75 6c 6c 00 02 31  |0.....devnull..1|
			00000740  00 11 00 00 00 65 70 68  65 6d 65 72 61 6c 46 6f  |.....ephemeralFo|
			00000750  72 54 65 73 74 00 02 32  00 0b 00 00 00 77 69 72  |rTest..2.....wir|
			00000760  65 64 54 69 67 65 72 00  00 01 6f 6b 00 00 00 00  |edTiger...ok....|
			00000770  00 00 00 f0 3f 00                                 |....?.|`,
		),
		j: `{"$k":["version","gitVersion","modules","allocator","javascriptEngine","sysInfo","versionArray","openssl","buildEnvironment","bits","debug","maxBsonObjectSize","storageEngines","ok"],"version":"5.0.0","gitVersion":"1184f004a99660de6f5e745573419bda8a28c0e9","modules":[],"allocator":"tcmalloc","javascriptEngine":"mozjs","sysInfo":"deprecated","versionArray":[5,0,0,0],"openssl":{"$k":["running","compiled"],"running":"OpenSSL 1.1.1f  31 Mar 2020","compiled":"OpenSSL 1.1.1f  31 Mar 2020"},"buildEnvironment":{"$k":["distmod","distarch","cc","ccflags","cxx","cxxflags","linkflags","target_arch","target_os","cppdefines"],"distmod":"ubuntu2004","distarch":"x86_64","cc":"/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0","ccflags":"-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer -fno-strict-aliasing -O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 -Wno-unused-local-typedefs -Wno-unused-function -Wno-deprecated-declarations -Wno-unused-const-variable -Wno-unused-but-set-variable -Wno-missing-braces -fstack-protector-strong -Wa,--nocompress-debug-sections -fno-builtin-memcmp","cxx":"/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0","cxxflags":"-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17","linkflags":"-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong -Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack -Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags","target_arch":"x86_64","target_os":"linux","cppdefines":"SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE _REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG BOOST_LOG_WITHOUT_THREAD_ATTR ABSL_FORCE_ALIGNED_ACCESS"},"bits":64,"debug":false,"maxBsonObjectSize":16777216,"storageEngines":["devnull","ephemeralForTest","wiredTiger"],"ok":{"$f":"1"}}`,
	}

	all = fuzzTestCase{
		name: "all",
		v: NewDocument(types.MakeDocument(
			"binary", types.Array{
				types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
				types.Binary{Subtype: types.BinaryGeneric, B: []byte{}},
			},
			"bool", types.Array{true, false},
			"datetime", types.Array{time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC), time.Time{}},
			"double", types.Array{42.13, 0.0},
			"int32", types.Array{int32(42), int32(0)},
			"int64", types.Array{int64(42), int64(0)},
			"objectID", types.Array{types.ObjectID{0x42}, types.ObjectID{}},
			"string", types.Array{"foo", ""},
			"timestamp", types.Array{types.Timestamp(42), types.Timestamp(0)},
		)),
		b: testutil.MustParseDump(`
			00000000  2d 01 00 00 04 62 69 6e  61 72 79 00 16 00 00 00  |-....binary.....|
			00000010  05 30 00 01 00 00 00 80  42 05 31 00 00 00 00 00  |.0......B.1.....|
			00000020  00 00 04 62 6f 6f 6c 00  0d 00 00 00 08 30 00 01  |...bool......0..|
			00000030  08 31 00 00 00 04 64 61  74 65 74 69 6d 65 00 1b  |.1....datetime..|
			00000040  00 00 00 09 30 00 2b e6  51 e7 7a 01 00 00 09 31  |....0.+.Q.z....1|
			00000050  00 00 28 d3 ed 7c c7 ff  ff 00 04 64 6f 75 62 6c  |..(..|.....doubl|
			00000060  65 00 1b 00 00 00 01 30  00 71 3d 0a d7 a3 10 45  |e......0.q=....E|
			00000070  40 01 31 00 00 00 00 00  00 00 00 00 00 04 69 6e  |@.1...........in|
			00000080  74 33 32 00 13 00 00 00  10 30 00 2a 00 00 00 10  |t32......0.*....|
			00000090  31 00 00 00 00 00 00 04  69 6e 74 36 34 00 1b 00  |1.......int64...|
			000000a0  00 00 12 30 00 2a 00 00  00 00 00 00 00 12 31 00  |...0.*........1.|
			000000b0  00 00 00 00 00 00 00 00  00 04 6f 62 6a 65 63 74  |..........object|
			000000c0  49 44 00 23 00 00 00 07  30 00 42 00 00 00 00 00  |ID.#....0.B.....|
			000000d0  00 00 00 00 00 00 07 31  00 00 00 00 00 00 00 00  |.......1........|
			000000e0  00 00 00 00 00 00 04 73  74 72 69 6e 67 00 18 00  |.......string...|
			000000f0  00 00 02 30 00 04 00 00  00 66 6f 6f 00 02 31 00  |...0.....foo..1.|
			00000100  01 00 00 00 00 00 04 74  69 6d 65 73 74 61 6d 70  |.......timestamp|
			00000110  00 1b 00 00 00 11 30 00  2a 00 00 00 00 00 00 00  |......0.*.......|
			00000120  11 31 00 00 00 00 00 00  00 00 00 00 00           |.1...........|
		`),
		j: `{"$k":["binary","bool","datetime","double","int32","int64","objectID","string","timestamp"],"binary":[{"$b":"Qg==","s":128},{"$b":"","s":0}],"bool":[true,false],"datetime":[{"$d":"1627378542123"},{"$d":"-62135596800000"}],"double":[{"$f":"42.13"},{"$f":"0"}],"int32":[42,0],"int64":[{"$l":"42"},{"$l":"0"}],"objectID":[{"$o":"420000000000000000000000"},{"$o":"000000000000000000000000"}],"string":["foo",""],"timestamp":[{"$t":"42"},{"$t":"0"}]}`,
	}

	documentTestcases = []fuzzTestCase{handshake1, handshake2, handshake3, handshake4, all}
)

func TestDocument(t *testing.T) {
	// t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		// t.Parallel()
		testBinary(t, documentTestcases, func() bsontype { return new(Document) })
	})

	t.Run("JSON", func(t *testing.T) {
		// t.Parallel()
		testJSON(t, documentTestcases, func() bsontype { return new(Document) })
	})
}

func FuzzDocumentBinary(f *testing.F) {
	fuzzBinary(f, documentTestcases, func() bsontype { return new(Document) })
}

func FuzzDocumentJSON(f *testing.F) {
	fuzzJSON(f, documentTestcases, func() bsontype { return new(Document) })
}

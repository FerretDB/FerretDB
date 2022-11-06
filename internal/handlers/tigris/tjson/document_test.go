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

package tjson

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func convertDocument(d *types.Document) *documentType {
	res := documentType(*d)
	return &res
}

func prepareTestCases() []testCase {
	handshake1doc := must.NotFail(types.NewDocument(
		"_id", "handshake1",
		"ismaster", true,
		"client", must.NotFail(types.NewDocument(
			"driver", must.NotFail(types.NewDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			)),
			"os", must.NotFail(types.NewDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			)),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", must.NotFail(types.NewDocument(
				"name", "mongosh 1.0.1",
			)),
		)),
		// TODO Support arrays: https://github.com/FerretDB/FerretDB/issues/908
		// "compression", must.NotFail(types.NewArray("none")),
		"loadBalanced", false,
	))
	handshake1 := testCase{
		name:   "handshake1",
		v:      convertDocument(handshake1doc),
		schema: must.NotFail(DocumentSchema(handshake1doc)),
		j: `{"$k":["_id","ismaster","client","loadBalanced"],` +
			`"_id":"handshake1","ismaster":true,` +
			`"client":{"$k":["driver","os","platform","application"],"driver":{"$k":["name","version"],` +
			`"name":"nodejs","version":"4.0.0-beta.6"},"os":{"$k":["type","name","architecture","version"],` +
			`"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},` +
			`"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",` +
			`"application":{"$k":["name"],"name":"mongosh 1.0.1"}},"loadBalanced":false}`,
	}

	handshake2doc := must.NotFail(types.NewDocument(
		"_id", "handshake2",
		"ismaster", true,
		"client", must.NotFail(types.NewDocument(
			"driver", must.NotFail(types.NewDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			)),
			"os", must.NotFail(types.NewDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			)),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", must.NotFail(types.NewDocument(
				"name", "mongosh 1.0.1",
			)),
		)),
		// TODO Support arrays: https://github.com/FerretDB/FerretDB/issues/908
		// "compression", must.NotFail(types.NewArray("none")),
		"loadBalanced", false,
	))
	handshake2 := testCase{
		name:   "handshake2",
		v:      convertDocument(handshake2doc),
		schema: must.NotFail(DocumentSchema(handshake2doc)),
		j: `{"$k":["_id","ismaster","client","loadBalanced"],` +
			`"_id":"handshake2","ismaster":true,` +
			`"client":{"$k":["driver","os","platform","application"],"driver":{"$k":["name","version"],` +
			`"name":"nodejs","version":"4.0.0-beta.6"},"os":{"$k":["type","name","architecture","version"],` +
			`"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},` +
			`"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",` +
			`"application":{"$k":["name"],"name":"mongosh 1.0.1"}},"loadBalanced":false}`,
	}

	handshake3doc := must.NotFail(types.NewDocument(
		"_id", "handshake3",
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
	))
	handshake3 := testCase{
		name:   "handshake3",
		v:      convertDocument(handshake3doc),
		schema: must.NotFail(DocumentSchema((handshake3doc))),
		j: `{"$k":["_id","buildInfo","lsid","$db"],"_id":"handshake3","buildInfo":1,` +
			`"lsid":{"$k":["id"],"id":{"$b":"oxnytKF1QMe456OjLsJWvg==","s":4}},"$db":"admin"}`,
	}

	handshake4doc := must.NotFail(types.NewDocument(
		"_id", "handshake4",
		"version", "5.0.0",
		"gitVersion", "1184f004a99660de6f5e745573419bda8a28c0e9",
		// TODO Support arrays: https://github.com/FerretDB/FerretDB/issues/908
		// "modules", must.NotFail(types.NewArray()),
		"allocator", "tcmalloc",
		"javascriptEngine", "mozjs",
		"sysInfo", "deprecated",
		// TODO Support arrays: https://github.com/FerretDB/FerretDB/issues/908
		// "versionArray", must.NotFail(types.NewArray(int32(5), int32(0), int32(0), int32(0))),
		"openssl", must.NotFail(types.NewDocument(
			"running", "OpenSSL 1.1.1f  31 Mar 2020",
			"compiled", "OpenSSL 1.1.1f  31 Mar 2020",
		)),
		"buildEnvironment", must.NotFail(types.NewDocument(
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
		)),
		"bits", int32(64),
		"debug", false,
		"maxBsonObjectSize", int32(16777216),
		// TODO Support arrays:  https://github.com/FerretDB/FerretDB/issues/908
		// "storageEngines", must.NotFail(types.NewArray("devnull", "ephemeralForTest", "wiredTiger")),
		"ok", float64(1),
	))
	handshake4 := testCase{
		name:   "handshake4",
		v:      convertDocument(handshake4doc),
		schema: must.NotFail(DocumentSchema((handshake4doc))),
		j: `{"$k":["_id","version","gitVersion","allocator","javascriptEngine","sysInfo",` +
			`"openssl","buildEnvironment","bits","debug","maxBsonObjectSize","ok"],` +
			`"_id":"handshake4",` +
			`"version":"5.0.0","gitVersion":"1184f004a99660de6f5e745573419bda8a28c0e9",` +
			`"allocator":"tcmalloc","javascriptEngine":"mozjs","sysInfo":"deprecated",` +
			`"openssl":{"$k":["running","compiled"],"running":"OpenSSL 1.1.1f  31 Mar 2020",` +
			`"compiled":"OpenSSL 1.1.1f  31 Mar 2020"},` +
			`"buildEnvironment":{"$k":["distmod","distarch","cc","ccflags","cxx","cxxflags","linkflags",` +
			`"target_arch","target_os","cppdefines"],"distmod":"ubuntu2004","distarch":"x86_64",` +
			`"cc":"/opt/mongodbtoolchain/v3/bin/gcc: gcc (GCC) 8.5.0",` +
			`"ccflags":"-Werror -include mongo/platform/basic.h -fasynchronous-unwind-tables -ggdb -Wall ` +
			`-Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -fno-omit-frame-pointer -fno-strict-aliasing ` +
			`-O2 -march=sandybridge -mtune=generic -mprefer-vector-width=128 -Wno-unused-local-typedefs ` +
			`-Wno-unused-function -Wno-deprecated-declarations -Wno-unused-const-variable ` +
			`-Wno-unused-but-set-variable -Wno-missing-braces -fstack-protector-strong ` +
			`-Wa,--nocompress-debug-sections -fno-builtin-memcmp",` +
			`"cxx":"/opt/mongodbtoolchain/v3/bin/g++: g++ (GCC) 8.5.0",` +
			`"cxxflags":"-Woverloaded-virtual -Wno-maybe-uninitialized -fsized-deallocation -std=c++17",` +
			`"linkflags":"-Wl,--fatal-warnings -pthread -Wl,-z,now -fuse-ld=gold -fstack-protector-strong ` +
			`-Wl,--no-threads -Wl,--build-id -Wl,--hash-style=gnu -Wl,-z,noexecstack -Wl,--warn-execstack ` +
			`-Wl,-z,relro -Wl,--compress-debug-sections=none -Wl,-z,origin -Wl,--enable-new-dtags",` +
			`"target_arch":"x86_64","target_os":"linux",` +
			`"cppdefines":"SAFEINT_USE_INTRINSICS 0 PCRE_STATIC NDEBUG _XOPEN_SOURCE 700 _GNU_SOURCE ` +
			`_REENTRANT 1 _FORTIFY_SOURCE 2 BOOST_THREAD_VERSION 5 BOOST_THREAD_USES_DATETIME ` +
			`BOOST_SYSTEM_NO_DEPRECATED BOOST_MATH_NO_LONG_DOUBLE_MATH_FUNCTIONS BOOST_ENABLE_ASSERT_DEBUG_HANDLER ` +
			`BOOST_LOG_NO_SHORTHAND_NAMES BOOST_LOG_USE_NATIVE_SYSLOG BOOST_LOG_WITHOUT_THREAD_ATTR ` +
			`ABSL_FORCE_ALIGNED_ACCESS"},"bits":64,"debug":false,"maxBsonObjectSize":16777216,` +
			`"ok":1}`,
	}

	allDoc := must.NotFail(types.NewDocument(
		"_id", types.ObjectID(objectIDType{0x62, 0xea, 0x6a, 0x94, 0x3d, 0x44, 0xb1, 0x0e, 0x1b, 0x6b, 0x87, 0x97}),
		"binary", types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
		"bool", true,
		"double", 42.13,
		"int32", int32(42),
		"int64", int64(42),
		"timestamp", types.Timestamp(1562470521),
		"string", "foo",
		"object", must.NotFail(types.NewDocument("foo", "bar")),
		"regex", types.Regex{Pattern: "^foobar$", Options: "i"},
		"null", "foo",
	))
	allSchema := must.NotFail(DocumentSchema(allDoc))
	allDoc.Set("null", types.Null) // Check that `null` set in a valid document and can be handled correctly
	all := testCase{
		name:   "all",
		v:      convertDocument(allDoc),
		schema: allSchema,
		j: `{"$k":["_id","binary","bool","double","int32","int64","timestamp","string","object","regex","null"],` +
			`"_id":"YupqlD1EsQ4ba4eX","binary":{"$b":"Qg==","s":128},"bool":true,"double":42.13,"int32":42,` +
			`"int64":42,"timestamp":{"$t":"1562470521"},"string":"foo","object":{"$k":["foo"],"foo":"bar"},` +
			`"regex":{"$r":"^foobar$","o":"i"},"null":null}`,
	}

	// TODO Add a test case that contains arrays of various types (like in the pjson package):
	// https://github.com/FerretDB/FerretDB/issues/908

	eofDoc := must.NotFail(types.NewDocument("_id", "foo"))
	eof := testCase{
		name:   "EOF",
		schema: must.NotFail(DocumentSchema(eofDoc)),
		j:      `[`,
		jErr:   `unexpected EOF`,
	}

	mismatchedDoc := must.NotFail(types.NewDocument("_id", "foo"))
	mismatchedSchema := testCase{
		name:   "mismatched schema",
		v:      convertDocument(mismatchedDoc),
		schema: boolSchema,
		j:      `{"$k":["_id"],"_id":"foo"}`,
		sErr:   "json: cannot unmarshal object into Go value of type bool",
	}

	invalidDoc := must.NotFail(types.NewDocument("_id", "foo"))
	invalidSchema := testCase{
		name:   "mismatched schema",
		v:      convertDocument(invalidDoc),
		schema: &Schema{Type: "invalid"},
		j:      `{"$k":["_id"],"_id":"foo"}`,
		sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
	}

	return []testCase{handshake1, handshake2, handshake3, handshake4, all, eof, mismatchedSchema, invalidSchema}
}

func TestDocument(t *testing.T) {
	t.Parallel()
	testJSON(t, prepareTestCases(), func() tjsontype { return new(documentType) })
}

func FuzzDocument(f *testing.F) {
	fuzzJSON(f, prepareTestCases())
}

func BenchmarkDocument(b *testing.B) {
	benchmark(b, prepareTestCases())
}

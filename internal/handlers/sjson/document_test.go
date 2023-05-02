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

package sjson

import (
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func convertDocument(d *types.Document) *documentType {
	res := documentType(*d)
	return &res
}

var (
	handshake1 = testCase{
		name: "handshake1",
		v: convertDocument(must.NotFail(types.NewDocument(
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
			"compression", must.NotFail(types.NewArray("none")),
			"loadBalanced", false,
		))),
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{
					"ismaster": {Type: elemTypeBool},
					"client": {
						Type: elemTypeObject,
						Schema: &schema{
							Properties: map[string]*elem{
								"driver": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"name":    {Type: elemTypeString},
											"version": {Type: elemTypeString},
										},
										Keys: []string{"name", "version"},
									},
								},
								"os": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"type":         {Type: elemTypeString},
											"name":         {Type: elemTypeString},
											"architecture": {Type: elemTypeString},
											"version":      {Type: elemTypeString},
										},
										Keys: []string{"type", "name", "architecture", "version"},
									},
								},
								"platform": {
									Type: elemTypeString,
								},
								"application": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"name": {Type: elemTypeString},
										},
										Keys: []string{"name"},
									},
								},
							},
							Keys: []string{"driver", "os", "platform", "application"},
						},
					},
					"compression": {
						Type:  elemTypeArray,
						Items: []*elem{{Type: elemTypeString}},
					},
					"loadBalanced": {Type: elemTypeBool},
				},
				Keys: []string{"ismaster", "client", "compression", "loadBalanced"},
			},
		},
		j: `{"ismaster":true,` +
			`"client":{"driver":{` +
			`"name":"nodejs","version":"4.0.0-beta.6"},"os":{` +
			`"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},` +
			`"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",` +
			`"application":{"name":"mongosh 1.0.1"}},"compression":["none"],"loadBalanced":false}`,
	}

	handshake2 = testCase{
		name: "handshake2",
		v: convertDocument(must.NotFail(types.NewDocument(
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
			"compression", must.NotFail(types.NewArray("none")),
			"loadBalanced", false,
		))),
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{
					"ismaster": {
						Type: elemTypeBool,
					},
					"client": {
						Type: elemTypeObject,
						Schema: &schema{
							Properties: map[string]*elem{
								"driver": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"name": {
												Type: elemTypeString,
											},
											"version": {
												Type: elemTypeString,
											},
										},
										Keys: []string{"name", "version"},
									},
								},
								"os": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"type": {
												Type: elemTypeString,
											},
											"name": {
												Type: elemTypeString,
											},
											"architecture": {
												Type: elemTypeString,
											},
											"version": {
												Type: elemTypeString,
											},
										},
										Keys: []string{"type", "name", "architecture", "version"},
									},
								},
								"platform": {
									Type: elemTypeString,
								},
								"application": {
									Type: elemTypeObject,
									Schema: &schema{
										Properties: map[string]*elem{
											"name": {
												Type: elemTypeString,
											},
										},
										Keys: []string{"name"},
									},
								},
							},
							Keys: []string{"driver", "os", "platform", "application"},
						},
					},
					"compression": {
						Type:  elemTypeArray,
						Items: []*elem{stringSchema},
					},
					"loadBalanced": boolSchema,
				},
				Keys: []string{"ismaster", "client", "compression", "loadBalanced"},
			},
		},
		j: `{"ismaster":true,` +
			`"client":{"driver":{` +
			`"name":"nodejs","version":"4.0.0-beta.6"},"os":{` +
			`"type":"Darwin","name":"darwin","architecture":"x64","version":"20.6.0"},` +
			`"platform":"Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",` +
			`"application":{"name":"mongosh 1.0.1"}},"compression":["none"],"loadBalanced":false}`,
	}

	handshake3 = testCase{
		name: "handshake3",
		v: convertDocument(must.NotFail(types.NewDocument(
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
		))),
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{
					"buildInfo": intSchema,
					"lsid": {Type: elemTypeObject, Schema: &schema{
						Properties: map[string]*elem{
							"id": binDataSchema(types.BinaryUUID),
						},
						Keys: []string{"id"},
					}},
					"$db": stringSchema,
				},
				Keys: []string{"buildInfo", "lsid", "$db"},
			},
		},
		j: `{"buildInfo":1,"lsid":{"id":"oxnytKF1QMe456OjLsJWvg=="},"$db":"admin"}`,
	}

	handshake4 = testCase{
		name: "handshake4",
		v: convertDocument(must.NotFail(types.NewDocument(
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
			"storageEngines", must.NotFail(types.NewArray("devnull", "ephemeralForTest", "wiredTiger")),
			"ok", float64(1),
		))),
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{
					"version":          stringSchema,
					"gitVersion":       stringSchema,
					"modules":          {Type: elemTypeArray, Items: []*elem{}},
					"allocator":        stringSchema,
					"javascriptEngine": stringSchema,
					"sysInfo":          stringSchema,
					"versionArray":     {Type: elemTypeArray, Items: []*elem{intSchema, intSchema, intSchema, intSchema}},
					"openssl": {
						Type: elemTypeObject,
						Schema: &schema{
							Properties: map[string]*elem{"running": stringSchema, "compiled": stringSchema},
							Keys:       []string{"running", "compiled"},
						},
					},
					"buildEnvironment": {
						Type: elemTypeObject,
						Schema: &schema{
							Properties: map[string]*elem{
								"distmod":     stringSchema,
								"distarch":    stringSchema,
								"cc":          stringSchema,
								"ccflags":     stringSchema,
								"cxx":         stringSchema,
								"cxxflags":    stringSchema,
								"linkflags":   stringSchema,
								"target_arch": stringSchema,
								"target_os":   stringSchema,
								"cppdefines":  stringSchema,
							},
							Keys: []string{
								"distmod", "distarch", "cc", "ccflags", "cxx", "cxxflags", "linkflags",
								"target_arch", "target_os", "cppdefines",
							},
						},
					},
					"bits":              intSchema,
					"debug":             boolSchema,
					"maxBsonObjectSize": intSchema,
					"storageEngines":    {Type: elemTypeArray, Items: []*elem{stringSchema, stringSchema, stringSchema}},
					"ok":                doubleSchema,
				},
				Keys: []string{
					"version", "gitVersion", "modules", "allocator", "javascriptEngine", "sysInfo", "versionArray",
					"openssl", "buildEnvironment", "bits", "debug", "maxBsonObjectSize", "storageEngines", "ok",
				},
			},
		},
		j: `{` +
			`"version":"5.0.0","gitVersion":"1184f004a99660de6f5e745573419bda8a28c0e9","modules":[],` +
			`"allocator":"tcmalloc","javascriptEngine":"mozjs","sysInfo":"deprecated","versionArray":[5,0,0,0],` +
			`"openssl":{"running":"OpenSSL 1.1.1f  31 Mar 2020",` +
			`"compiled":"OpenSSL 1.1.1f  31 Mar 2020"},` +
			`"buildEnvironment":{"distmod":"ubuntu2004","distarch":"x86_64",` +
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
			`"storageEngines":["devnull","ephemeralForTest","wiredTiger"],"ok":1}`,
	}

	all = testCase{
		name: "all",
		v: convertDocument(must.NotFail(types.NewDocument(
			"binary", must.NotFail(types.NewArray(
				types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
				types.Binary{Subtype: types.BinaryGeneric, B: []byte{}},
			)),
			"bool", must.NotFail(types.NewArray(true, false)),
			"datetime", must.NotFail(types.NewArray(
				time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
				time.Time{}.Local(),
			)),
			"double", must.NotFail(types.NewArray(42.13, 0.0)),
			"int32", must.NotFail(types.NewArray(int32(42), int32(0))),
			"int64", must.NotFail(types.NewArray(int64(42), int64(0))),
			"objectID", must.NotFail(types.NewArray(types.ObjectID{0x42}, types.ObjectID{})),
			"string", must.NotFail(types.NewArray("foo", "")),
			"timestamp", must.NotFail(types.NewArray(types.Timestamp(42), types.Timestamp(0))),
			"null", must.NotFail(types.NewArray(types.Null, types.Null)),
		))),
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{
					"binary": {Type: elemTypeArray, Items: []*elem{
						binDataSchema(types.BinaryUser), binDataSchema(types.BinaryGeneric),
					}},
					"bool":      {Type: elemTypeArray, Items: []*elem{boolSchema, boolSchema}},
					"datetime":  {Type: elemTypeArray, Items: []*elem{dateSchema, dateSchema}},
					"double":    {Type: elemTypeArray, Items: []*elem{doubleSchema, doubleSchema}},
					"int32":     {Type: elemTypeArray, Items: []*elem{intSchema, intSchema}},
					"int64":     {Type: elemTypeArray, Items: []*elem{longSchema, longSchema}},
					"objectID":  {Type: elemTypeArray, Items: []*elem{objectIDSchema, objectIDSchema}},
					"string":    {Type: elemTypeArray, Items: []*elem{stringSchema, stringSchema}},
					"timestamp": {Type: elemTypeArray, Items: []*elem{timestampSchema, timestampSchema}},
					"null":      {Type: elemTypeArray, Items: []*elem{nullSchema, nullSchema}},
				},
				Keys: []string{
					"binary", "bool", "datetime", "double", "int32", "int64", "objectID", "string", "timestamp", "null",
				},
			},
		},
		j: `{` +
			`"binary":["Qg==",""],"bool":[true,false],` +
			`"datetime":[1627378542123,-62135596800000],"double":[42.13,0],` +
			`"int32":[42,0],"int64":[42,0],` +
			`"objectID":["420000000000000000000000","000000000000000000000000"],` +
			`"string":["foo",""],"timestamp":[42,0],"null":[null,null]}`,
	}

	eof = testCase{
		name: "EOF",
		sch: &elem{
			Type: elemTypeObject,
			Schema: &schema{
				Properties: map[string]*elem{},
				Keys:       []string{},
			},
		},
		j:    `[`,
		jErr: `unexpected EOF`,
	}

	nilSchema = testCase{
		name: "NilSchema",
		sch: &elem{
			Type:   elemTypeObject,
			Schema: nil,
		},
		j:    `{"foo": "bar"}`,
		jErr: `document schema is nil for non-empty document`,
	}

	emptySchema = testCase{
		name: "NilSchema",
		sch: &elem{
			Type:   elemTypeObject,
			Schema: new(schema),
		},
		j:    `{"foo": "bar"}`,
		jErr: `sjson.documentType.UnmarshalJSON: 0 elements in $k in the schema, 1 in the document`,
	}

	documentTestCases = []testCase{handshake1, handshake2, handshake3, handshake4, all, eof, nilSchema, emptySchema}
)

func TestDocument(t *testing.T) {
	t.Parallel()
	testJSON(t, documentTestCases, func() sjsontype { return new(documentType) })
}

func FuzzDocument(f *testing.F) {
	fuzzJSON(f, documentTestCases, func() sjsontype { return new(documentType) })
}

func BenchmarkDocument(b *testing.B) {
	benchmark(b, documentTestCases, func() sjsontype { return new(documentType) })
}

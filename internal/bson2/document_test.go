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

package bson2

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// testCase represents a single test case.
//
//nolint:vet // for readability
type testCase struct {
	name      string
	b         []byte
	doc       *types.Document
	decodeErr error
}

var (
	handshake1 = testCase{
		name: "handshake1",
		b:    testutil.MustParseDumpFile("testdata", "handshake1.hex"),
		doc: must.NotFail(types.NewDocument(
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
		)),
	}

	handshake2 = testCase{
		name: "handshake2",
		b:    testutil.MustParseDumpFile("testdata", "handshake2.hex"),
		doc: must.NotFail(types.NewDocument(
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
		)),
	}

	handshake3 = testCase{
		name: "handshake3",
		b:    testutil.MustParseDumpFile("testdata", "handshake3.hex"),
		doc: must.NotFail(types.NewDocument(
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
		)),
	}

	handshake4 = testCase{
		name: "handshake4",
		b:    testutil.MustParseDumpFile("testdata", "handshake4.hex"),
		doc: must.NotFail(types.NewDocument(
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
		)),
	}

	all = testCase{
		name: "all",
		b:    testutil.MustParseDumpFile("testdata", "all.hex"),
		doc: must.NotFail(types.NewDocument(
			"array", must.NotFail(types.NewArray(
				must.NotFail(types.NewArray("")),
				must.NotFail(types.NewArray("foo")),
			)),
			"binary", must.NotFail(types.NewArray(
				types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
				types.Binary{Subtype: types.BinaryGeneric, B: []byte{}},
			)),
			"bool", must.NotFail(types.NewArray(true, false)),
			"datetime", must.NotFail(types.NewArray(
				time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
				time.Time{}.Local(),
			)),
			"document", must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument("foo", "")),
				must.NotFail(types.NewDocument("", "foo")),
			)),
			"double", must.NotFail(types.NewArray(42.13, 0.0)),
			"int32", must.NotFail(types.NewArray(int32(42), int32(0))),
			"int64", must.NotFail(types.NewArray(int64(42), int64(0))),
			"objectID", must.NotFail(types.NewArray(types.ObjectID{0x42}, types.ObjectID{})),
			"string", must.NotFail(types.NewArray("foo", "")),
			"timestamp", must.NotFail(types.NewArray(types.Timestamp(42), types.Timestamp(0))),
		)),
	}

	eof = testCase{
		name:      "EOF",
		b:         []byte{0x00},
		decodeErr: ErrDecodeShortInput,
	}

	smallDoc = testCase{
		name: "smallDoc",
		b: []byte{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subdocument length and end of subdocument
			0x00, // end of document
		},
		doc: must.NotFail(types.NewDocument(
			"foo", must.NotFail(types.NewDocument()),
		)),
	}

	shortDoc = testCase{
		name: "shortDoc",
		b: []byte{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x06, 0x00, 0x00, 0x00, 0x00, // invalid subdocument length and end of subdocument
			0x00, // end of document
		},
		decodeErr: ErrDecodeShortInput,
	}

	smallArray = testCase{
		name: "smallArray",
		b: []byte{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x04, 0x66, 0x6f, 0x6f, 0x00, // subarray "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subarray length and end of subarray
			0x00, // end of document
		},
		doc: must.NotFail(types.NewDocument(
			"foo", must.NotFail(types.NewArray()),
		)),
	}

	shortArray = testCase{
		name: "shortArray",
		b: []byte{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x04, 0x66, 0x6f, 0x6f, 0x00, // subarray "foo"
			0x06, 0x00, 0x00, 0x00, 0x00, // invalid subarray length and end of subarray
			0x00, // end of document
		},
		decodeErr: ErrDecodeShortInput,
	}

	duplicateKeys = testCase{
		name: "duplicateKeys",
		b: []byte{
			0x0b, 0x00, 0x00, 0x00, // document length
			0x08, 0x00, 0x00, // "": false
			0x08, 0x00, 0x01, // "": true
			0x00, // end of document
		},
		doc: must.NotFail(types.NewDocument(
			"", false,
			"", true,
		)),
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3759
	/*
		fuzz1 = testCase{
			name: "fuzz1-702a93bb4d6e1425",
			b: []byte{
				0x4d, 0x01, 0x00, 0x00, // document length
				0x08,                                                 // bool
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x00, // "00000000"
				0x01,                   // true
				0x03,                   // nested document
				0x30, 0x30, 0x30, 0x30, // nested document length
				0x30, 0x30, 0x00, // "00"
				0x26, 0x01, 0x00, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x08, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x00, 0x00, 0x00,
			},
			decodeErr: ErrDecodeShortInput,
		}
	*/

	documentTestCases = []testCase{
		handshake1, handshake2, handshake3, handshake4, all,
		eof, smallDoc, shortDoc, smallArray, shortArray, duplicateKeys,
	}
)

func TestDocument(t *testing.T) {
	for _, tc := range documentTestCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			require.NotEqual(t, tc.doc == nil, tc.decodeErr == nil)

			t.Run("Encode", func(t *testing.T) {
				if tc.doc == nil {
					t.Skip()
				}

				t.Run("bson", func(t *testing.T) {
					doc, err := bson.ConvertDocument(tc.doc)
					require.NoError(t, err)

					actual, err := doc.MarshalBinary()
					require.NoError(t, err)
					assert.Equal(t, tc.b, actual, "actual:\n%s", hex.Dump(actual))
				})

				t.Run("bson2", func(t *testing.T) {
					doc, err := ConvertDocument(tc.doc)
					require.NoError(t, err)

					actual, err := encodeDocument(doc)
					require.NoError(t, err)
					assert.Equal(t, tc.b, actual, "actual:\n%s", hex.Dump(actual))
				})
			})

			t.Run("Decode", func(t *testing.T) {
				t.Run("bson", func(t *testing.T) {
					var doc bson.Document
					buf := bufio.NewReader(bytes.NewReader(tc.b))
					err := doc.ReadFrom(buf)

					if tc.decodeErr != nil {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)

					_, err = buf.ReadByte()
					assert.Equal(t, err, io.EOF)

					actual, err := types.ConvertDocument(&doc)
					require.NoError(t, err)
					testutil.AssertEqual(t, tc.doc, actual)
				})

				t.Run("bson2", func(t *testing.T) {
					doc, err := DecodeDocument(tc.b)

					if tc.decodeErr != nil {
						require.Error(t, err, "b:\n\n%s\n%#v", hex.Dump(tc.b), tc.b)
						require.ErrorIs(t, err, tc.decodeErr)
						return
					}
					require.NoError(t, err)

					actual, err := doc.Convert()
					require.NoError(t, err)
					testutil.AssertEqual(t, tc.doc, actual)
				})
			})
		})
	}
}

func FuzzDocument(f *testing.F) {
	for _, tc := range documentTestCases {
		f.Add(tc.b)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		t.Run("bson2", func(t *testing.T) {
			t.Parallel()

			doc, err := DecodeDocument(b)
			if err != nil {
				t.Skip()
			}

			actual, err := encodeDocument(doc)
			require.NoError(t, err)
			assert.Equal(t, b, actual, "actual:\n%s", hex.Dump(actual))
		})

		t.Run("cross", func(t *testing.T) {
			t.Parallel()

			br := bytes.NewReader(b)
			bufr := bufio.NewReader(br)

			var doc1 bson.Document
			err1 := doc1.ReadFrom(bufr)

			if err1 != nil {
				_, err2 := DecodeDocument(b)
				require.Error(t, err2, "bson1 err = %v", err1)
				return
			}

			// remove extra tail
			cb := b[:len(b)-bufr.Buffered()-br.Len()]

			doc2, err2 := DecodeDocument(cb)
			require.NoError(t, err2)

			d1, err := types.ConvertDocument(&doc1)
			require.NoError(t, err)

			d2, err := doc2.Convert()
			require.NoError(t, err)

			testutil.AssertEqual(t, d1, d2)
		})
	})
}

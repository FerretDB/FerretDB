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

// normalTestCase represents a single test case for successful decoding/encoding.
//
//nolint:vet // for readability
type normalTestCase struct {
	name string
	raw  RawDocument
	tdoc *types.Document
}

// decodeTestCase represents a single test case for unsuccessful decoding.
//
//nolint:vet // for readability
type decodeTestCase struct {
	name string
	raw  RawDocument

	oldOk bool

	findRawErr    error
	findRawL      int
	decodeErr     error
	decodeDeepErr error // defaults to decodeErr
}

// normalTestCases represents test cases for successful decoding/encoding.
var normalTestCases = []normalTestCase{
	{
		name: "handshake1",
		raw:  testutil.MustParseDumpFile("testdata", "handshake1.hex"),
		tdoc: must.NotFail(types.NewDocument(
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
	},
	{
		name: "handshake2",
		raw:  testutil.MustParseDumpFile("testdata", "handshake2.hex"),
		tdoc: must.NotFail(types.NewDocument(
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
	},
	{
		name: "handshake3",
		raw:  testutil.MustParseDumpFile("testdata", "handshake3.hex"),
		tdoc: must.NotFail(types.NewDocument(
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
	},
	{
		name: "handshake4",
		raw:  testutil.MustParseDumpFile("testdata", "handshake4.hex"),
		tdoc: must.NotFail(types.NewDocument(
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
	},
	{
		name: "all",
		raw:  testutil.MustParseDumpFile("testdata", "all.hex"),
		tdoc: must.NotFail(types.NewDocument(
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
	},
	{
		name: "float64Doc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x01, 0x66, 0x00,
			0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", float64(3.141592653589793),
		)),
	},
	{
		name: "stringDoc",
		raw: RawDocument{
			0x0e, 0x00, 0x00, 0x00,
			0x02, 0x66, 0x00,
			0x02, 0x00, 0x00, 0x00,
			0x76, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", "v",
		)),
	},
	{
		name: "binaryDoc",
		raw: RawDocument{
			0x0e, 0x00, 0x00, 0x00,
			0x05, 0x66, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x80,
			0x76,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", types.Binary{B: []byte("v"), Subtype: types.BinaryUser},
		)),
	},
	{
		name: "objectIDDoc",
		raw: RawDocument{
			0x14, 0x00, 0x00, 0x00,
			0x07, 0x66, 0x00,
			0x62, 0x56, 0xc5, 0xba, 0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40},
		)),
	},
	{
		name: "boolDoc",
		raw: RawDocument{
			0x09, 0x00, 0x00, 0x00,
			0x08, 0x66, 0x00,
			0x01,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", true,
		)),
	},
	{
		name: "timeDoc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x09, 0x66, 0x00,
			0x0b, 0xce, 0x82, 0x18, 0x8d, 0x01, 0x00, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", time.Date(2024, 1, 17, 17, 40, 42, 123000000, time.UTC),
		)),
	},
	{
		name: "nullDoc",
		raw: RawDocument{
			0x08, 0x00, 0x00, 0x00,
			0x0a, 0x66, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", types.Null,
		)),
	},
	{
		name: "regexDoc",
		raw: RawDocument{
			0x0c, 0x00, 0x00, 0x00,
			0x0b, 0x66, 0x00,
			0x70, 0x00,
			0x6f, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", types.Regex{Pattern: "p", Options: "o"},
		)),
	},
	{
		name: "int32Doc",
		raw: RawDocument{
			0x0c, 0x00, 0x00, 0x00,
			0x10, 0x66, 0x00,
			0xa1, 0xb0, 0xb9, 0x12,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", int32(314159265),
		)),
	},
	{
		name: "timestampDoc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x11, 0x66, 0x00,
			0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", types.Timestamp(42),
		)),
	},
	{
		name: "int64Doc",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00,
			0x12, 0x66, 0x00,
			0x21, 0x6d, 0x25, 0xa, 0x43, 0x29, 0xb, 0x00,
			0x00,
		},
		tdoc: must.NotFail(types.NewDocument(
			"f", int64(3141592653589793),
		)),
	},
	{
		name: "smallDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subdocument length and end of subdocument
			0x00, // end of document
		},
		tdoc: must.NotFail(types.NewDocument(
			"foo", must.NotFail(types.NewDocument()),
		)),
	},
	{
		name: "smallArray",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x04, 0x66, 0x6f, 0x6f, 0x00, // subarray "foo"
			0x05, 0x00, 0x00, 0x00, 0x00, // subarray length and end of subarray
			0x00, // end of document
		},
		tdoc: must.NotFail(types.NewDocument(
			"foo", must.NotFail(types.NewArray()),
		)),
	},
	{
		name: "duplicateKeys",
		raw: RawDocument{
			0x0b, 0x00, 0x00, 0x00, // document length
			0x08, 0x00, 0x00, // "": false
			0x08, 0x00, 0x01, // "": true
			0x00, // end of document
		},
		tdoc: must.NotFail(types.NewDocument(
			"", false,
			"", true,
		)),
	},
}

// decodeTestCases represents test cases for unsuccessful decoding.
var decodeTestCases = []decodeTestCase{
	{
		name:       "EOF",
		raw:        RawDocument{0x00},
		findRawErr: ErrDecodeShortInput,
		decodeErr:  ErrDecodeShortInput,
	},
	{
		name: "invalidLength",
		raw: RawDocument{
			0x00, 0x00, 0x00, 0x00, // invalid document length
			0x00, // end of document
		},
		findRawErr: ErrDecodeInvalidInput,
		decodeErr:  ErrDecodeInvalidInput,
	},
	{
		name: "extraByte",
		raw: RawDocument{
			0x05, 0x00, 0x00, 0x00, // document length
			0x00, // end of document
			0x00, // extra byte
		},
		oldOk:     true,
		findRawL:  5,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "unexpectedTag",
		raw: RawDocument{
			0x06, 0x00, 0x00, 0x00, // document length
			0xdd, // unexpected tag
			0x00, // end of document
		},
		findRawL:  6,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidTag",
		raw: RawDocument{
			0x06, 0x00, 0x00, 0x00, // document length
			0x00, // invalid tag
			0x00, // end of document
		},
		findRawL:  6,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "shortDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x06, 0x00, 0x00, 0x00, // invalid subdocument length
			0x00, // end of subdocument
			0x00, // end of document
		},
		findRawL:      15,
		decodeErr:     ErrDecodeShortInput,
		decodeDeepErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidDoc",
		raw: RawDocument{
			0x0f, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x05, 0x00, 0x00, 0x00, // subdocument length
			0x30, // invalid end of subdocument
			0x00, // end of document
		},
		findRawL:  15,
		decodeErr: ErrDecodeInvalidInput,
	},
	{
		name: "invalidDocTag",
		raw: RawDocument{
			0x10, 0x00, 0x00, 0x00, // document length
			0x03, 0x66, 0x6f, 0x6f, 0x00, // subdocument "foo"
			0x06, 0x00, 0x00, 0x00, // subdocument length
			0x00, // invalid tag
			0x00, // end of subdocument
			0x00, // end of document
		},
		findRawL:      16,
		decodeDeepErr: ErrDecodeInvalidInput,
	},
}

func TestNormal(t *testing.T) {
	prev := noNaN
	noNaN = false

	t.Cleanup(func() { noNaN = prev })

	for _, tc := range normalTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("bson", func(t *testing.T) {
				t.Run("ReadFrom", func(t *testing.T) {
					var doc bson.Document
					buf := bufio.NewReader(bytes.NewReader(tc.raw))
					err := doc.ReadFrom(buf)
					require.NoError(t, err)

					_, err = buf.ReadByte()
					assert.Equal(t, err, io.EOF)

					tdoc, err := types.ConvertDocument(&doc)
					require.NoError(t, err)
					testutil.AssertEqual(t, tc.tdoc, tdoc)
				})

				t.Run("MarshalBinary", func(t *testing.T) {
					doc, err := bson.ConvertDocument(tc.tdoc)
					require.NoError(t, err)

					raw, err := doc.MarshalBinary()
					require.NoError(t, err)
					assert.Equal(t, []byte(tc.raw), raw)
				})
			})

			t.Run("bson2", func(t *testing.T) {
				t.Run("FindRaw", func(t *testing.T) {
					ls := tc.raw.LogValue().Resolve().String()
					assert.NotContains(t, ls, "panicked")
					assert.NotContains(t, ls, "called too many times")

					l, err := FindRaw(tc.raw)
					require.NoError(t, err)
					require.Len(t, tc.raw, l)
				})

				t.Run("DecodeEncode", func(t *testing.T) {
					doc, err := tc.raw.Decode()
					require.NoError(t, err)

					ls := doc.LogValue().Resolve().String()
					assert.NotContains(t, ls, "panicked")
					assert.NotContains(t, ls, "called too many times")

					tdoc, err := doc.Convert()
					require.NoError(t, err)
					testutil.AssertEqual(t, tc.tdoc, tdoc)

					raw, err := doc.Encode()
					require.NoError(t, err)
					assert.Equal(t, tc.raw, raw)
				})

				t.Run("DecodeDeepEncode", func(t *testing.T) {
					doc, err := tc.raw.DecodeDeep()
					require.NoError(t, err)

					ls := doc.LogValue().Resolve().String()
					assert.NotContains(t, ls, "panicked")
					assert.NotContains(t, ls, "called too many times")

					tdoc, err := doc.Convert()
					require.NoError(t, err)
					testutil.AssertEqual(t, tc.tdoc, tdoc)

					raw, err := doc.Encode()
					require.NoError(t, err)
					assert.Equal(t, tc.raw, raw)
				})

				t.Run("ConvertEncode", func(t *testing.T) {
					doc, err := ConvertDocument(tc.tdoc)
					require.NoError(t, err)

					raw, err := doc.Encode()
					require.NoError(t, err)
					assert.Equal(t, tc.raw, raw)
				})
			})
		})
	}
}

func TestDecode(t *testing.T) {
	prev := noNaN
	noNaN = false

	t.Cleanup(func() { noNaN = prev })

	for _, tc := range decodeTestCases {
		if tc.decodeDeepErr == nil {
			tc.decodeDeepErr = tc.decodeErr
		}

		require.NotNil(t, tc.decodeDeepErr, "invalid test case %q", tc.name)

		t.Run(tc.name, func(t *testing.T) {
			t.Run("bson", func(t *testing.T) {
				t.Run("ReadFrom", func(t *testing.T) {
					var doc bson.Document
					buf := bufio.NewReader(bytes.NewReader(tc.raw))
					err := doc.ReadFrom(buf)

					if tc.oldOk {
						require.NoError(t, err)
						return
					}

					require.Error(t, err)
				})
			})

			t.Run("bson2", func(t *testing.T) {
				t.Run("FindRaw", func(t *testing.T) {
					l, err := FindRaw(tc.raw)

					if tc.findRawErr != nil {
						require.ErrorIs(t, err, tc.findRawErr)
						return
					}

					require.NoError(t, err)
					require.Equal(t, tc.findRawL, l)
				})

				t.Run("Decode", func(t *testing.T) {
					_, err := tc.raw.Decode()

					if tc.decodeErr != nil {
						require.ErrorIs(t, err, tc.decodeErr)
						return
					}

					require.NoError(t, err)
				})

				t.Run("DecodeDeep", func(t *testing.T) {
					_, err := tc.raw.DecodeDeep()
					require.ErrorIs(t, err, tc.decodeDeepErr)
				})
			})
		})
	}
}

func BenchmarkDocument(b *testing.B) {
	prev := noNaN
	noNaN = false

	b.Cleanup(func() { noNaN = prev })

	for _, tc := range normalTestCases {
		b.Run(tc.name, func(b *testing.B) {
			b.Run("bson", func(b *testing.B) {
				b.Run("ReadFrom", func(b *testing.B) {
					var doc bson.Document
					var buf *bufio.Reader
					var err error
					br := bytes.NewReader(tc.raw)

					b.ReportAllocs()
					b.ResetTimer()

					for range b.N {
						_, _ = br.Seek(0, io.SeekStart)
						buf = bufio.NewReader(br)
						err = doc.ReadFrom(buf)
					}

					b.StopTimer()

					require.NoError(b, err)
				})

				b.Run("MarshalBinary", func(b *testing.B) {
					doc, err := bson.ConvertDocument(tc.tdoc)
					require.NoError(b, err)

					var raw []byte

					b.ReportAllocs()
					b.ResetTimer()

					for range b.N {
						raw, err = doc.MarshalBinary()
					}

					b.StopTimer()

					require.NoError(b, err)
					assert.NotNil(b, raw)
				})
			})

			b.Run("bson2", func(b *testing.B) {
				var doc *Document
				var raw []byte
				var err error

				b.Run("Decode", func(b *testing.B) {
					b.ReportAllocs()

					for range b.N {
						doc, err = tc.raw.Decode()
					}

					b.StopTimer()

					require.NoError(b, err)
					require.NotNil(b, doc)
				})

				b.Run("Encode", func(b *testing.B) {
					doc, err = tc.raw.Decode()
					require.NoError(b, err)

					b.ReportAllocs()
					b.ResetTimer()

					for range b.N {
						raw, err = doc.Encode()
					}

					b.StopTimer()

					require.NoError(b, err)
					assert.NotNil(b, raw)
				})

				b.Run("DecodeDeep", func(b *testing.B) {
					b.ReportAllocs()

					for range b.N {
						doc, err = tc.raw.DecodeDeep()
					}

					b.StopTimer()

					require.NoError(b, err)
					require.NotNil(b, doc)
				})

				b.Run("EncodeDeep", func(b *testing.B) {
					doc, err = tc.raw.DecodeDeep()
					require.NoError(b, err)

					b.ReportAllocs()
					b.ResetTimer()

					for range b.N {
						raw, err = doc.Encode()
					}

					b.StopTimer()

					require.NoError(b, err)
					assert.NotNil(b, raw)
				})
			})
		})
	}
}

// testRawDocument tests a single RawDocument (that might or might not be valid).
// It is adapted from tests above.
func testRawDocument(t *testing.T, rawDoc RawDocument) {
	t.Helper()

	t.Run("bson2", func(t *testing.T) {
		t.Run("FindRaw", func(t *testing.T) {
			ls := rawDoc.LogValue().Resolve().String()
			assert.NotContains(t, ls, "panicked")
			assert.NotContains(t, ls, "called too many times")

			_, _ = FindRaw(rawDoc)
		})

		t.Run("DecodeEncode", func(t *testing.T) {
			doc, err := rawDoc.Decode()
			if err != nil {
				_, err = rawDoc.DecodeDeep()
				assert.Error(t, err) // it might be different

				return
			}

			ls := doc.LogValue().Resolve().String()
			assert.NotContains(t, ls, "panicked")
			assert.NotContains(t, ls, "called too many times")

			_, _ = doc.Convert()

			raw, err := doc.Encode()
			if err == nil {
				assert.Equal(t, rawDoc, raw)
			}
		})

		t.Run("DecodeDeepEncode", func(t *testing.T) {
			doc, err := rawDoc.DecodeDeep()
			if err != nil {
				return
			}

			ls := doc.LogValue().Resolve().String()
			assert.NotContains(t, ls, "panicked")
			assert.NotContains(t, ls, "called too many times")

			_, err = doc.Convert()
			require.NoError(t, err)

			raw, err := doc.Encode()
			require.NoError(t, err)
			assert.Equal(t, rawDoc, raw)
		})
	})

	t.Run("cross", func(t *testing.T) {
		br := bytes.NewReader(rawDoc)
		bufr := bufio.NewReader(br)

		var doc1 bson.Document
		err1 := doc1.ReadFrom(bufr)

		if err1 != nil {
			_, err2 := rawDoc.DecodeDeep()
			require.Error(t, err2, "bson1 err = %v", err1)

			return
		}

		// remove extra tail
		b := []byte(rawDoc[:len(rawDoc)-bufr.Buffered()-br.Len()])
		l, err := FindRaw(rawDoc)
		require.NoError(t, err)
		require.Equal(t, b, []byte(rawDoc[:l]))

		// decode

		bdoc2, err2 := RawDocument(b).DecodeDeep()
		require.NoError(t, err2)

		ls := bdoc2.LogValue().Resolve().String()
		assert.NotContains(t, ls, "panicked")
		assert.NotContains(t, ls, "called too many times")

		tdoc1, err := types.ConvertDocument(&doc1)
		require.NoError(t, err)

		tdoc2, err := bdoc2.Convert()
		require.NoError(t, err)

		testutil.AssertEqual(t, tdoc1, tdoc2)

		// encode

		doc1e, err := bson.ConvertDocument(tdoc1)
		require.NoError(t, err)

		doc2e, err := ConvertDocument(tdoc2)
		require.NoError(t, err)

		ls = doc2e.LogValue().Resolve().String()
		assert.NotContains(t, ls, "panicked")
		assert.NotContains(t, ls, "called too many times")

		b1, err := doc1e.MarshalBinary()
		require.NoError(t, err)

		b2, err := doc2e.Encode()
		require.NoError(t, err)

		assert.Equal(t, b1, []byte(b2))
		assert.Equal(t, b, []byte(b2))
	})
}

func FuzzDocument(f *testing.F) {
	prev := noNaN
	noNaN = false

	f.Cleanup(func() { noNaN = prev })

	for _, tc := range normalTestCases {
		f.Add([]byte(tc.raw))
	}

	for _, tc := range decodeTestCases {
		f.Add([]byte(tc.raw))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		testRawDocument(t, RawDocument(b))

		l, err := FindRaw(b)
		if err == nil {
			testRawDocument(t, RawDocument(b[:l]))
		}
	})
}

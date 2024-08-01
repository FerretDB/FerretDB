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

package types_test

import (
	"bytes"
	"context"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestLogging is a copy of [bson_test.TestLoggingNil] modified to fit `types` value types.
func TestLoggingNil(t *testing.T) {
	var doc *types.Document
	assert.Equal(t, doc.LogValue().String(), "Document<nil>")

	var arr *types.Array
	assert.Equal(t, arr.LogValue().String(), "Array<nil>")
}

// TestLogging is a copy of [bson_test.TestLogging] modified to fit `types` value types.
func TestLogging(t *testing.T) {
	ctx := context.Background()

	var cbuf, tbuf, jbuf bytes.Buffer
	clog := slog.New(logging.NewHandler(&cbuf, &logging.NewHandlerOpts{
		Base:          "console",
		RemoveTime:    true,
		RemoveLevel:   true,
		RemoveSource:  true,
		CheckMessages: true,
	}))
	tlog := slog.New(logging.NewHandler(&tbuf, &logging.NewHandlerOpts{
		Base:          "text",
		RemoveTime:    true,
		RemoveLevel:   true,
		RemoveSource:  true,
		CheckMessages: true,
	}))
	jlog := slog.New(logging.NewHandler(&jbuf, &logging.NewHandlerOpts{
		Base:          "json",
		RemoveTime:    true,
		RemoveLevel:   true,
		RemoveSource:  true,
		CheckMessages: true,
	}))

	for _, tc := range []struct {
		name string
		doc  any
		c    string
		t    string
		j    string
	}{
		{
			name: "Numbers",
			doc: must.NotFail(types.NewDocument(
				"f64", 42.0,
				"inf", float64(math.Inf(1)),
				"neg_inf", float64(math.Inf(-1)),
				"zero", math.Copysign(0, 1),
				"neg_zero", math.Copysign(0, -1),
				"nan", float64(math.NaN()),
				"i32", int32(42),
				"i64", int64(42),
			)),
			c: `	{"v":{"f64":42,"i32":42,"i64":42,"inf":"+Inf","nan":"NaN","neg_inf":"-Inf","neg_zero":-0,"zero":0}}`,
			t: `v.f64=42 v.inf=+Inf v.neg_inf=-Inf v.zero=0 v.neg_zero=-0 v.nan=NaN v.i32=42 v.i64=42`,
			j: `{"v":{"f64":42,"inf":"+Inf","neg_inf":"-Inf","zero":0,"neg_zero":-0,"nan":"NaN","i32":42,"i64":42}}`,
		},
		{
			name: "Scalars",
			doc: must.NotFail(types.NewDocument(
				"null", types.Null,
				"id", types.ObjectID{0x42},
				"bool", true,
				"time", time.Date(2023, 3, 6, 13, 14, 42, 123456789, time.FixedZone("", int(4*time.Hour.Seconds()))),
			)),
			c: `	{"v":{"bool":true,"id":"ObjectID(420000000000000000000000)","null":null,"time":"2023-03-06T09:14:42.123Z"}}`,
			t: `v.null=<nil> v.id=ObjectID(420000000000000000000000) v.bool=true v.time=2023-03-06T09:14:42.123Z`,
			j: `{"v":{"null":null,"id":"ObjectID(420000000000000000000000)","bool":true,"time":"2023-03-06T09:14:42.123Z"}}`,
		},
		{
			name: "Composites",
			doc: must.NotFail(types.NewDocument(
				"doc", must.NotFail(types.NewDocument(
					"foo", "bar",
					"baz", must.NotFail(types.NewDocument(
						"qux", "quux",
					)),
				)),
				"doc_empty", must.NotFail(types.NewDocument()),
				"array", must.NotFail(types.NewArray(
					"foo",
					"bar",
					must.NotFail(types.NewArray("baz", "qux")),
				)),
			)),
			c: `	{"v":{"array":{"0":"foo","1":"bar","2":{"0":"baz","1":"qux"}},` +
				`"doc":{"baz":{"qux":"quux"},"foo":"bar"}}}`,
			t: `v.doc.foo=bar v.doc.baz.qux=quux ` +
				`v.array.0=foo v.array.1=bar v.array.2.0=baz v.array.2.1=qux`,
			j: `{"v":{"doc":{"foo":"bar","baz":{"qux":"quux"}},` +
				`"array":{"0":"foo","1":"bar","2":{"0":"baz","1":"qux"}}}}`,
		},
		{
			name: "Nested",
			doc:  makeNested(false, 20).(*types.Document),
			c: `	{"v":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":` +
				`{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":null}}}}}}}}}}}}}}}}}}}}}`,
			t: `v.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0.f.0=<nil>`,
			j: `{"v":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":` +
				`{"f":{"0":{"f":{"0":{"f":{"0":{"f":{"0":null}}}}}}}}}}}}}}}}}}}}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clog.InfoContext(ctx, "", slog.Any("v", tc.doc))
			assert.Equal(t, tc.c+"\n", cbuf.String(), "console output mismatch")
			cbuf.Reset()

			tlog.InfoContext(ctx, "", slog.Any("v", tc.doc))
			assert.Equal(t, tc.t+"\n", tbuf.String(), "text output mismatch")
			tbuf.Reset()

			jlog.InfoContext(ctx, "", slog.Any("v", tc.doc))
			assert.Equal(t, tc.j+"\n", jbuf.String(), "json output mismatch")
			jbuf.Reset()
		})
	}
}

// makeNested creates a nested document or array with the given depth.
func makeNested(array bool, depth int) any {
	if depth < 1 {
		panic("depth must be at least 1")
	}

	var child any = types.Null

	if depth > 1 {
		child = makeNested(!array, depth-1)
	}

	if array {
		return must.NotFail(types.NewArray(child))
	}

	return must.NotFail(types.NewDocument("f", child))
}
